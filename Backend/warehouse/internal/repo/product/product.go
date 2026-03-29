package product

import (
    "context"
    "fmt"
    "log/slog"
    "sync"
    "warehouse/internal/repo/review"
    "warehouse/pkg"

    "github.com/google/uuid"
    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgxpool"
    "golang.org/x/sync/errgroup"
)

type product struct {
    writeConn   func(s string) (*pgxpool.Pool, pkg.ShardNum)
    readConn    func(s string) (*pgxpool.Pool, pkg.ShardNum)
    allReadConn func() []*pgxpool.Pool

    log *slog.Logger
}

func (p *product) GetProductById(ctx context.Context, id string) (*ProductWithAVG, []review.Review, error) {
    productID, err := uuid.Parse(id)
    if err != nil {
        return nil, nil, fmt.Errorf("invalid product ID format: %w", err)
    }

    const productSQL = `
        SELECT
            id,
            preview_url,
            name,
            description,
            price,
            width,
            height,
            depth,
            weight,
            volume,
            created_at,
            visibility,
            count,
            CASE WHEN avg IS NULL 
                THEN 0.0
                ELSE avg
            END AS avg,
            stock_quantity
        FROM
        (
            SELECT
                p.id::text,
                p.preview_url,
                p.name,
                p.description,
                p.price,
                p.width,
                p.height,
                p.depth,
                p.weight,
                p.volume,
                p.created_at,
                p.visibility,
                COUNT(r.id) AS count,
                AVG(r.grade) AS avg,
                COALESCE(SUM(sp.quantity), 0) AS stock_quantity
            FROM product AS p
                LEFT JOIN review AS r
                    ON p.id = r.product_id
                LEFT JOIN shelf_product AS sp
                    ON p.id = sp.product_id
            WHERE p.id = $1
            GROUP BY
                p.id
        ) AS a
    `

    var prod ProductWithAVG
    conn, _ := p.readConn(id)

    err = conn.QueryRow(ctx, productSQL, productID).Scan(
        &prod.Id,
        &prod.PreviewURL,
        &prod.Name,
        &prod.Description,
        &prod.Price,
        &prod.Width,
        &prod.Height,
        &prod.Depth,
        &prod.Weight,
        &prod.Volume,
        &prod.CreatedAt,
        &prod.Visibility,
        &prod.Count,
        &prod.Avg,
        &prod.StockQuantity,
    )
    if err != nil {
        if err == pgx.ErrNoRows {
            return nil, nil, fmt.Errorf("product not found")
        }
        return nil, nil, fmt.Errorf("failed to fetch product: %w", err)
    }

    const reviewsSQL = `
        SELECT id, login, grade, description, created_at
        FROM review
        WHERE product_id = $1
        ORDER BY created_at DESC
    `

    rows, err := conn.Query(ctx, reviewsSQL, productID)
    if err != nil {
        return nil, nil, fmt.Errorf("failed to query reviews: %w", err)
    }
    defer rows.Close()

    var reviews []review.Review
    for rows.Next() {
        var rev review.Review
        err := rows.Scan(&rev.Id, &rev.Login, &rev.Grade, &rev.Description, &rev.CreatedAt)
        if err != nil {
            return nil, nil, fmt.Errorf("failed to scan review: %w", err)
        }
        reviews = append(reviews, rev)
    }

    if err = rows.Err(); err != nil {
        return nil, nil, fmt.Errorf("row iteration error: %w", err)
    }

    return &prod, reviews, nil
}

func (p *product) GetProducts(ctx context.Context, visibility *bool) ([]*ProductWithAVG, error) {
    sql := `
        SELECT
            id,
            preview_url,
            name,
            description,
            price,
            width,
            height,
            depth,
            weight,
            volume,
            created_at,
            visibility,
            count,
            CASE WHEN avg IS NULL 
                THEN 0.0
                ELSE avg
            END AS avg,
            stock_quantity
        FROM
        (
            SELECT
                p.id::text,
                p.preview_url,
                p.name,
                p.description,
                p.price,
                p.width,
                p.height,
                p.depth,
                p.weight,
                p.volume,
                p.created_at,
                p.visibility,
                COUNT(r.id) AS count,
                AVG(r.grade) AS avg,
                COALESCE(SUM(sp.quantity), 0) AS stock_quantity
            FROM product AS p
                LEFT JOIN review AS r
                    ON p.id = r.product_id
                LEFT JOIN shelf_product AS sp
                    ON p.id = sp.product_id
            `
    if visibility != nil {
        sql += `
        WHERE visibility = $1
        `
    }
    sql += `
            GROUP BY
                p.id
        ) AS a
    `

    var answ []*ProductWithAVG
    conns := p.allReadConn()

    errGroup, _ := errgroup.WithContext(ctx)
    mtx := &sync.Mutex{}

    for _, conn := range conns {
        errGroup.Go(func() error {
            var rows pgx.Rows
            var err error
            if visibility != nil {
                rows, err = conn.Query(ctx, sql, visibility)
            } else {
                rows, err = conn.Query(ctx, sql)
            }
            if err != nil {
                return err
            }
            defer rows.Close()

            var idStr string
            mtx.Lock()
            defer mtx.Unlock()
            for rows.Next() {
                body := &ProductWithAVG{}
                if err := rows.Scan(
                    &idStr,
                    &body.PreviewURL,
                    &body.Name,
                    &body.Description,
                    &body.Price,
                    &body.Width,
                    &body.Height,
                    &body.Depth,
                    &body.Weight,
                    &body.Volume,
                    &body.CreatedAt,
                    &body.Visibility,
                    &body.Count,
                    &body.Avg,
                    &body.StockQuantity,
                ); err != nil {
                    return err
                }

                body.Id = idStr

                answ = append(answ, body)
            }

            return nil
        })
    }

    return answ, errGroup.Wait()
}

func (p *product) CreateProduct(ctx context.Context, req *CreateProduct) (*ProductWithAVG, error) {
    productID := uuid.New()

    conn, _ := p.writeConn(productID.String())

    const sql = `
        INSERT INTO product (
            id, 
            preview_url, 
            name, 
            description, 
            price, 
            width, 
            height, 
            depth, 
            weight, 
            volume, 
            created_at, 
            visibility
        )
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW(), $11)
        RETURNING 
            id, 
            preview_url, 
            name, 
            description, 
            price, 
            width, 
            height, 
            depth, 
            weight, 
            volume, 
            created_at, 
            visibility
    `

    var result ProductWithAVG
    err := conn.QueryRow(ctx, sql,
        productID,
        req.PreviewURL,
        req.Name,
        req.Description,
        req.Price,
        req.Width,
        req.Height,
        req.Depth,
        req.Weight,
        req.Volume,
        req.Visibility,
    ).Scan(
        &result.Id,
        &result.PreviewURL,
        &result.Name,
        &result.Description,
        &result.Price,
        &result.Width,
        &result.Height,
        &result.Depth,
        &result.Weight,
        &result.Volume,
        &result.CreatedAt,
        &result.Visibility,
    )
    if err != nil {
        p.log.ErrorContext(ctx, "failed to create product",
            "product_id", productID.String(),
            "error", err)
        return nil, fmt.Errorf("failed to create product: %w", err)
    }

    result.Avg = 0
    result.Count = 0
    result.StockQuantity = 0

    p.log.InfoContext(ctx, "product created successfully",
        "product_id", result.Id,
        "name", result.Name)

    return &result, nil
}

func NewProduct(
    writeConn func(s string) (*pgxpool.Pool, pkg.ShardNum),
    readConn func(s string) (*pgxpool.Pool, pkg.ShardNum),
    allReadConn func() []*pgxpool.Pool,
    log *slog.Logger,
) ProductInterface {
    return &product{
        writeConn:   writeConn,
        readConn:    readConn,
        allReadConn: allReadConn,
        log:         log,
    }
}