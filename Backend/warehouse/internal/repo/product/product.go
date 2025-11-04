package product

import (
	"context"
	"log/slog"
	"sync"
	"warehouse/pkg"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/sync/errgroup"
)

type product struct {
	writeConn   func(s string) (*pgxpool.Pool, pkg.ShardNum)
	readConn    func(s string) (*pgxpool.Pool, pkg.ShardNum)
	allReadConn func() []*pgxpool.Pool

	log *slog.Logger
}

func (p *product) GetProductById(ctx context.Context, id string) (*Product, error) {
	const sql = `
		SELECT
			id::text
    		,preview_url
    		,name
    		,description
			,price
    		,created_at
    		,visibility
		FROM product
		WHERE id = $1
	`

	var idStr string
	dto := &Product{}
	conn, _ := p.readConn(id)
	if err := conn.QueryRow(ctx, sql, id).Scan(&idStr, &dto.Preview_url, &dto.Name,
		&dto.Description, &dto.Price, &dto.Created_at, &dto.Visibility); err != nil {
		return nil, err
	}

	var err error
	dto.Id, err = uuid.Parse(idStr)
	if err != nil {
		return nil, err
	}

	return dto, nil
}

func (p *product) GetProducts(ctx context.Context, visibility *bool) ([]*Product, error) {
	sql := `
		SELECT
			id::text
    		,preview_url
    		,name
    		,description
			,price
    		,created_at
		FROM product
		WHERE visibility = $1
	`
	var vis bool
	if *visibility {
		vis = *visibility
	} else {
		vis = *visibility
	}

	var answ []*Product
	conns := p.allReadConn()
	
	errGroup, _ := errgroup.WithContext(ctx)
	mtx := &sync.Mutex{}

	for _, conn := range conns {
		errGroup.Go(func() error {
			rows, err := conn.Query(ctx, sql, vis)
			if err != nil {
				return err	
			}
			defer rows.Close()

			var idStr string
			mtx.Lock()
			defer mtx.Unlock()
			for rows.Next() {
				body := &Product{}
				if err := rows.Scan(&idStr, &body.Preview_url, &body.Name, &body.Description, &body.Price, &body.Created_at);
					err != nil {
					return err
				}

				body.Id, err = uuid.Parse(idStr)
				if err != nil {
					return err
				}

				answ = append(answ, body)
			}

			return nil
		})
	}

	return answ, errGroup.Wait()
}

func NewProduct(writeConn func(s string) (*pgxpool.Pool, pkg.ShardNum),
	readConn func(s string) (*pgxpool.Pool, pkg.ShardNum),
	allReadConn func() []*pgxpool.Pool,
	log *slog.Logger) ProductInterface {
	return &product{
		writeConn:   writeConn,
		readConn:    readConn,
		allReadConn: allReadConn,
		log:         log,
	}
}
