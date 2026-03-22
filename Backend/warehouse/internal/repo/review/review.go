package review

import (
	"context"
	"log/slog"
	"warehouse/pkg"

	"github.com/jackc/pgx/v5/pgxpool"
)

type review struct {
	writeConn   func(s string) (*pgxpool.Pool, pkg.ShardNum)
	readConn    func(s string) (*pgxpool.Pool, pkg.ShardNum)
	allReadConn func() []*pgxpool.Pool

	log *slog.Logger
}

func (r *review) GetReviews(ctx context.Context, productId string) ([]*Review, error) {
	conns := r.allReadConn()
	reviews := make([]*Review, 0)

	sql := `
	SELECT	id
			,login
			,grade
			,description
			,created_at
	FROM review
	WHERE product_id = $1
	`

	for _, conn := range conns {
		row, err := conn.Query(ctx, sql, productId)
		if err != nil {
			return nil, err
		}

		review := &Review{}

		for row.Next() {
			if err := row.Scan(&review.Id, &review.Grade, &review.Description, &review.CreatedAt); err != nil {
				return nil, err
			}
		}
		reviews = append(reviews, review)
	}

	return reviews, nil
}

func (r *review) CreateReview(ctx context.Context, productID string, rev Review) (int, error) {
	query := `
		INSERT INTO review (login, grade, description, product_id)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`

	conn, _ := r.writeConn(productID)

	var id int
	err := conn.QueryRow(ctx, query,
		rev.Login,
		rev.Grade,
		rev.Description,
		productID,
	).Scan(&id)

	if err != nil {
		return 0, err
	}

	return id, nil
}

func NewReview(writeConn func(s string) (*pgxpool.Pool, pkg.ShardNum),
	readConn func(s string) (*pgxpool.Pool, pkg.ShardNum),
	allReadConn func() []*pgxpool.Pool,
	log *slog.Logger) ReviewInterface {
	return &review{
		writeConn:   writeConn,
		readConn:    readConn,
		allReadConn: allReadConn,
		log:         log,
	}
}