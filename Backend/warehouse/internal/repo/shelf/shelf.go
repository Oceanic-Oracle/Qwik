package shelf

import (
    "context"
    "errors"
    "fmt"
    "log/slog"
    "warehouse/internal/dto"
    "warehouse/pkg"
    "time"

    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/redis/go-redis/v9"
)

var (
    ErrNoSpace         = errors.New("недостаточно места на складе")
    ErrNotEnoughStock  = errors.New("недостаточно товара для списания")
    ErrProductNotFound = errors.New("товар не найден")
    ErrNoVolume        = errors.New("у товара не указан объем (нужен для расчета места)")
    ErrShelfLocked     = errors.New("полка сейчас обрабатывается другим запросом, попробуйте позже")
)

type ShelfRepo struct {
    writeConn   func(s string) (*pgxpool.Pool, pkg.ShardNum)
    readConn    func(s string) (*pgxpool.Pool, pkg.ShardNum)
    allReadConn func() []*pgxpool.Pool
    redis       *redis.Client
    log         *slog.Logger
}

func NewShelfRepo(
    writeConn func(s string) (*pgxpool.Pool, pkg.ShardNum),
    readConn func(s string) (*pgxpool.Pool, pkg.ShardNum),
    allReadConn func() []*pgxpool.Pool,
    redisClient *redis.Client,
    log *slog.Logger,
) *ShelfRepo {
    return &ShelfRepo{
        writeConn:   writeConn,
        readConn:    readConn,
        allReadConn: allReadConn,
        redis:       redisClient,
        log:         log,
    }
}

// ShelfEntity внутренняя модель полки
type ShelfEntity struct {
    ID           string
    RackID       string
    Level        int
    Priority     float64
    MaxCapacity  float64
    UsedCapacity float64
}

// GetAvailableShelves возвращает полки с местом, гарантируя изоляцию SKU (на полке только 1 товар)
func (r *ShelfRepo) GetAvailableShelves(
    ctx context.Context,
    requiredCapacity float64,
    productID string,
) ([]*ShelfEntity, error) {
    query := `
        SELECT 
            s.id::text, s.rack_id::text, s.level, s.priority, 
            s.max_capacity, s.used_capacity
        FROM shelf s
        WHERE (s.max_capacity - s.used_capacity) >= $1
          AND NOT EXISTS (
              SELECT 1 FROM shelf_product sp 
              WHERE sp.shelf_id = s.id AND sp.product_id::text != $2
          )
        ORDER BY s.priority DESC, s.level ASC
        LIMIT 100
    `

    var shelves []*ShelfEntity
    conns := r.allReadConn()

    for _, conn := range conns {
        rows, err := conn.Query(ctx, query, requiredCapacity, productID)
        if err != nil {
            r.log.Warn("ошибка запроса полок на шарде", "error", err)
            continue
        }

        for rows.Next() {
            var shelf ShelfEntity
            err := rows.Scan(
                &shelf.ID, &shelf.RackID, &shelf.Level, &shelf.Priority,
                &shelf.MaxCapacity, &shelf.UsedCapacity,
            )
            if err != nil {
                r.log.Warn("ошибка сканирования полки", "error", err)
                continue
            }
            shelves = append(shelves, &shelf)
        }
        rows.Close()
    }

    return shelves, nil
}

// GetProductVolume возвращает объем товара по его ID (нужен для оптимизатора)
func (r *ShelfRepo) GetProductVolume(ctx context.Context, productID string) (float64, error) {
    conn, _ := r.readConn(productID)
    var volume float64
    
    err := conn.QueryRow(ctx, "SELECT COALESCE(volume, 0) FROM product WHERE id = $1", productID).Scan(&volume)
    if err != nil {
        if err == pgx.ErrNoRows {
            return 0, ErrProductNotFound
        }
        return 0, fmt.Errorf("ошибка получения объема товара: %w", err)
    }
    
    return volume, nil
}

// AllocateProduct размещает товар на конкретной полке с глобальной блокировкой
func (r *ShelfRepo) AllocateProduct(ctx context.Context, shelfID, productID string, quantity int) error {
    mutexKey := fmt.Sprintf("shelf_lock:%s", shelfID)
    lockTTL := 5 * time.Second

    acquired, err := r.redis.SetNX(ctx, mutexKey, "locked", lockTTL).Result()
    if err != nil {
        return fmt.Errorf("ошибка проверки блокировки полки: %w", err)
    }
    if !acquired {
        return ErrShelfLocked
    }
    
    defer func() {
        luaScript := redis.NewScript(`
            if redis.call("get", KEYS[1]) == ARGV[1] then
                return redis.call("del", KEYS[1])
            else
                return 0
            end
        `)
        luaScript.Run(context.Background(), r.redis, []string{mutexKey}, "locked")
    }()

    conn, _ := r.writeConn(productID)

    tx, err := conn.Begin(ctx)
    if err != nil {
        return fmt.Errorf("не удалось начать транзакцию: %w", err)
    }
    defer func() {
        if tx != nil {
            _ = tx.Rollback(ctx)
        }
    }()

    var maxCap, usedCap float64
    err = tx.QueryRow(ctx,
        "SELECT max_capacity, used_capacity FROM shelf WHERE id = $1 FOR UPDATE",
        shelfID,
    ).Scan(&maxCap, &usedCap)
    if err != nil {
        return fmt.Errorf("полка не найдена: %w", err)
    }

    var volume float64
    err = tx.QueryRow(ctx,
        "SELECT COALESCE(volume, 0) FROM product WHERE id = $1",
        productID,
    ).Scan(&volume)
    if err != nil {
        if err == pgx.ErrNoRows {
            return ErrProductNotFound
        }
        return fmt.Errorf("ошибка получения товара: %w", err)
    }

    requiredCap := float64(quantity) * volume
    if usedCap+requiredCap > maxCap {
        return ErrNoSpace
    }

    _, err = tx.Exec(ctx, `
        INSERT INTO shelf_product (shelf_id, product_id, quantity) 
        VALUES ($1, $2, $3) 
        ON CONFLICT (shelf_id, product_id) DO UPDATE SET quantity = shelf_product.quantity + $3
    `, shelfID, productID, quantity)

    if err != nil {
        return fmt.Errorf("ошибка размещения товара: %w", err)
    }

    if err = tx.Commit(ctx); err != nil {
        return fmt.Errorf("ошибка коммита: %w", err)
    }
    tx = nil

    return nil
}

// AutoAddStock распределяет товар с учетом изоляции (каждому товару своя полка)
func (r *ShelfRepo) AutoAddStock(ctx context.Context, productID string, quantity int) (*dto.AutoStockResponse, error) {
    conn, _ := r.readConn(productID)
    var volume float64
    if err := conn.QueryRow(ctx, "SELECT COALESCE(volume, 0) FROM product WHERE id = $1", productID).Scan(&volume); err != nil {
        if err == pgx.ErrNoRows {
            return nil, ErrProductNotFound
        }
        return nil, err
    }

    if volume == 0 {
        return nil, ErrNoVolume
    }

    shelves, err := r.GetAvailableShelves(ctx, volume, productID)
    if err != nil || len(shelves) == 0 {
        return nil, ErrNoSpace
    }

    resp := &dto.AutoStockResponse{
        ProductID:      productID,
        TotalRequested: quantity,
    }

    remaining := quantity

    for _, currentShelf := range shelves {
        if remaining <= 0 {
            break
        }

        freeSpace := currentShelf.MaxCapacity - currentShelf.UsedCapacity
        if freeSpace <= 0 {
            continue
        }

        fitQty := int(freeSpace / volume)
        if fitQty <= 0 {
            continue
        }

        actualQty := remaining
        if fitQty < remaining {
            actualQty = fitQty
        }

        err := r.AllocateProduct(ctx, currentShelf.ID, productID, actualQty)
        if err == nil {
            var newQty int
            var newUsedCap float64
            conn.QueryRow(ctx, `SELECT quantity FROM shelf_product WHERE shelf_id = $1 AND product_id = $2`, currentShelf.ID, productID).Scan(&newQty)
            conn.QueryRow(ctx, `SELECT used_capacity FROM shelf WHERE id = $1`, currentShelf.ID).Scan(&newUsedCap)

            resp.Results = append(resp.Results, &dto.StockResponse{
                ShelfID:      currentShelf.ID,
                ProductID:    productID,
                NewQuantity:  newQty,
                UsedCapacity: newUsedCap,
                MaxCapacity:  currentShelf.MaxCapacity,
                Message:      fmt.Sprintf("размещено %d шт. (стеллаж %s, ур. %d). Занято: %.3f м³ из %.1f м³", actualQty, currentShelf.RackID, currentShelf.Level, newUsedCap, currentShelf.MaxCapacity),
            })
            remaining -= actualQty
        } else if err == ErrNoSpace || err == ErrShelfLocked {
            continue
        } else {
            return nil, err
        }
    }

    resp.TotalPlaced = quantity - remaining

    if resp.TotalPlaced == 0 {
        return nil, ErrNoSpace
    }

    if remaining > 0 {
        resp.Message = fmt.Sprintf("⚠️ Склад заполнен (нет пустых или своих полок)! Размещено %d из %d. Не хватило места для %d шт.", resp.TotalPlaced, quantity, remaining)
    } else {
        resp.Message = "✅ Весь товар успешно распределен по полкам"
    }

    return resp, nil
}

// WithdrawStock списание товара с конкретной полки
func (r *ShelfRepo) WithdrawStock(ctx context.Context, shelfID, productID string, quantity int) (*dto.StockResponse, error) {
    mutexKey := fmt.Sprintf("shelf_lock:%s", shelfID)
    acquired, err := r.redis.SetNX(ctx, mutexKey, "locked", 5*time.Second).Result()
    if err != nil {
        return nil, fmt.Errorf("ошибка проверки блокировки: %w", err)
    }
    if !acquired {
        return nil, ErrShelfLocked
    }
    defer func() {
        luaScript := redis.NewScript(`
            if redis.call("get", KEYS[1]) == ARGV[1] then
                return redis.call("del", KEYS[1])
            else
                return 0
            end
        `)
        luaScript.Run(context.Background(), r.redis, []string{mutexKey}, "locked")
    }()

    conn, _ := r.writeConn(productID)

    tx, err := conn.Begin(ctx)
    if err != nil {
        return nil, err
    }
    defer func() {
        if tx != nil {
            _ = tx.Rollback(ctx)
        }
    }()

    var currentQty int
    err = tx.QueryRow(ctx, `
        SELECT quantity FROM shelf_product 
        WHERE shelf_id = $1 AND product_id = $2 FOR UPDATE
    `, shelfID, productID).Scan(&currentQty)

    if err != nil {
        if err == pgx.ErrNoRows {
            return nil, ErrNotEnoughStock
        }
        return nil, err
    }

    if currentQty < quantity {
        return nil, ErrNotEnoughStock
    }

    newQty := currentQty - quantity
    if newQty == 0 {
        _, err = tx.Exec(ctx, "DELETE FROM shelf_product WHERE shelf_id = $1 AND product_id = $2", shelfID, productID)
    } else {
        _, err = tx.Exec(ctx, "UPDATE shelf_product SET quantity = $1 WHERE shelf_id = $2 AND product_id = $3", newQty, shelfID, productID)
    }
    if err != nil {
        return nil, err
    }

    var newUsedCap float64
    tx.QueryRow(ctx, "SELECT used_capacity FROM shelf WHERE id = $1", shelfID).Scan(&newUsedCap)

    if err = tx.Commit(ctx); err != nil {
        return nil, err
    }
    tx = nil

    return &dto.StockResponse{
        ShelfID:      shelfID,
        ProductID:    productID,
        NewQuantity:  newQty,
        UsedCapacity: newUsedCap,
        Message:      "успешно списано",
    }, nil
}