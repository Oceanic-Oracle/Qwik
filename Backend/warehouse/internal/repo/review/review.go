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

// CREATE TABLE review (
//     id SERIAL PRIMARY KEY,
//     login TEXT NOT NULL,
//     grade INTEGER NOT NULL CHECK (grade BETWEEN 1 AND 5),
//     description TEXT,
//     product_id UUID NOT NULL REFERENCES product(id) ON DELETE CASCADE,
//     created_at TIMESTAMPTZ DEFAULT NOW(),
//     UNIQUE (login, product_id)
// );
