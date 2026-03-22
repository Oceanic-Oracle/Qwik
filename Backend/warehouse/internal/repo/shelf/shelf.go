package shelf

import (
	"context"
	"fmt"
	"log/slog"

	"warehouse/internal/domain"
	"warehouse/pkg"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/twpayne/go-geom"
	"github.com/twpayne/go-geom/encoding/ewkb"
)

type ShelfRepo struct {
	writeConn   func(s string) (*pgxpool.Pool, pkg.ShardNum)
	readConn    func(s string) (*pgxpool.Pool, pkg.ShardNum)
	allReadConn func() []*pgxpool.Pool
	log         *slog.Logger
}

func NewShelfRepo(
	writeConn func(s string) (*pgxpool.Pool, pkg.ShardNum),
	readConn func(s string) (*pgxpool.Pool, pkg.ShardNum),
	allReadConn func() []*pgxpool.Pool,
	log *slog.Logger,
) *ShelfRepo {
	return &ShelfRepo{
		writeConn:   writeConn,
		readConn:    readConn,
		allReadConn: allReadConn,
		log:         log,
	}
}

// GetAvailableShelves возвращает полки с достаточной свободной ёмкостью
// Формула: I_shelf = W1*P_level + W2*P_priority - W3*ST_Distance(...)
func (r *ShelfRepo) GetAvailableShelves(
	ctx context.Context,
	requiredCapacity float64,
	productTags []string,
	warehouseID string,
	weights map[string]float64,
) ([]*domain.Shelf, error) {
	// Фильтр совместимости тегов: полка разрешает ВСЕ теги товара ИЛИ не имеет ограничений
	tagCondition := `
		(COALESCE(s.compatible_tags, ARRAY[]::text[]) = ARRAY[]::text[] 
		OR s.compatible_tags @> $3::text[])
	`

	// Важно: параметры запроса должны соответствовать порядку $1, $2, $3
	query := fmt.Sprintf(`
		SELECT 
			s.id, s.rack_id, s.level, s.priority, 
			s.max_capacity, s.used_capacity,
			ST_AsEWKB(s.location) as location_ewkb,
			s.compatible_tags
		FROM shelf s
		JOIN rack r ON s.rack_id = r.id
		WHERE r.warehouse_id = $1
			AND (s.max_capacity - s.used_capacity) >= $2
			AND %s
		ORDER BY s.priority DESC, s.level ASC
		LIMIT 100
	`, tagCondition)

	var shelves []*domain.Shelf
	conns := r.allReadConn()

	for _, conn := range conns {
		rows, err := conn.Query(ctx, query, warehouseID, requiredCapacity, productTags)
		if err != nil {
			r.log.Warn("ошибка запроса полок на шарде", "error", err)
			continue
		}

		for rows.Next() {
			var (
				shelf        domain.Shelf
				locationEWKB []byte
			)
			err := rows.Scan(
				&shelf.ID, &shelf.RackID, &shelf.Level, &shelf.Priority,
				&shelf.MaxCapacity, &shelf.UsedCapacity,
				&locationEWKB, &shelf.CompatibleTags,
			)
			if err != nil {
				r.log.Warn("ошибка сканирования полки", "error", err)
				continue
			}

			// Декодируем PostGIS-геометрию
			geomObj, err := ewkb.Unmarshal(locationEWKB)
			if err != nil {
				r.log.Warn("не удалось декодировать геометрию", "shelf_id", shelf.ID, "error", err)
				continue
			}

			// Корректная тип-ассерция для go-geom: *geom.Point
			if point, ok := geomObj.(*geom.Point); ok {
				shelf.Location = *point
			}

			shelves = append(shelves, &shelf)
		}
		rows.Close()
	}

	return shelves, nil
}

// ReserveCapacity атомарно резервирует место на полке (SELECT FOR UPDATE)
// Использует явный Begin/Commit/Rollback для совместимости
func (r *ShelfRepo) ReserveCapacity(ctx context.Context, shelfID string, capacity float64) error {
	conn, _ := r.writeConn(shelfID)

	// Явное начало транзакции
	tx, err := conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("не удалось начать транзакцию: %w", err)
	}
	// Дефер для отката в случае паники или ошибки до Commit
	defer func() {
		if tx != nil {
			_ = tx.Rollback(ctx) // Rollback после Commit — no-op
		}
	}()

	// Блокируем строку полки
	var currentUsed float64
	err = tx.QueryRow(ctx,
		"SELECT used_capacity FROM shelf WHERE id = $1 FOR UPDATE",
		shelfID,
	).Scan(&currentUsed)
	if err != nil {
		return fmt.Errorf("не удалось заблокировать полку: %w", err)
	}

	// Получаем максимальную ёмкость
	var maxCap float64
	err = tx.QueryRow(ctx,
		"SELECT max_capacity FROM shelf WHERE id = $1",
		shelfID,
	).Scan(&maxCap)
	if err != nil {
		return fmt.Errorf("не удалось получить ёмкость полки: %w", err)
	}

	// Проверка доступного места
	if currentUsed+capacity > maxCap {
		return fmt.Errorf("недостаточно места: требуется %.2f, доступно %.2f",
			capacity, maxCap-currentUsed)
	}

	// Обновляем used_capacity
	_, err = tx.Exec(ctx,
		"UPDATE shelf SET used_capacity = used_capacity + $1 WHERE id = $2",
		capacity, shelfID,
	)
	if err != nil {
		return fmt.Errorf("ошибка обновления ёмкости: %w", err)
	}

	// Коммит транзакции
	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("ошибка коммита транзакции: %w", err)
	}
	tx = nil // Отключаем defer-rollback после успешного коммита

	r.log.Debug("ёмкость зарезервирована",
		"shelf_id", shelfID,
		"capacity", capacity,
		"new_used", currentUsed+capacity)

	return nil
}

// CheckTagCompatibility проверяет совместимость тегов товара и полки
func (r *ShelfRepo) CheckTagCompatibility(ctx context.Context, shelfID string, productTags []string) (bool, error) {
	conn, _ := r.readConn(shelfID)

	var allowedTags []string
	err := conn.QueryRow(ctx,
		"SELECT compatible_tags FROM shelf WHERE id = $1",
		shelfID,
	).Scan(&allowedTags)
	if err != nil {
		return false, fmt.Errorf("не удалось получить теги полки: %w", err)
	}

	// Пустой список = все теги разрешены
	if len(allowedTags) == 0 {
		return true, nil
	}

	// Строгая проверка: полка должна разрешать ВСЕ теги товара
	tagSet := make(map[string]bool, len(allowedTags))
	for _, tag := range allowedTags {
		tagSet[tag] = true
	}

	for _, pTag := range productTags {
		if !tagSet[pTag] {
			r.log.Debug("несовместимость тегов",
				"shelf_id", shelfID,
				"rejected_tag", pTag,
				"allowed_tags", allowedTags)
			return false, nil
		}
	}

	return true, nil
}