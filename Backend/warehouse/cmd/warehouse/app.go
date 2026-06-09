package main

import (
	"context"
	"crypto/sha256"
	"os"
	"os/signal"
	"syscall"
	"warehouse/internal/config"
	"warehouse/internal/repo"
	"warehouse/internal/server"
	"warehouse/internal/service"
	"warehouse/pkg"

	"github.com/jackc/pgx/v5/pgxpool"
)

// DBTurnoverProvider вычисляет метрики оборачиваемости из реальных данных БД.
// TurnoverRate  = кол-во отзывов × средняя оценка + кол-во единиц на полках × 0.5
// DemandCV      = STDDEV(оценок) / AVG(оценок) — коэффициент вариации спроса
type DBTurnoverProvider struct {
	pool *pkg.ConnectionPool
}

func (d *DBTurnoverProvider) readConn(productID string) *pgxpool.Pool {
	h := sha256.Sum256([]byte(productID))
	shard := pkg.ShardNum(h[0] % uint8(len(*d.pool)))
	return (*d.pool)[shard].ReadNode
}

func (d *DBTurnoverProvider) GetTurnoverRate(productID string) (float64, error) {
	var rate float64
	err := d.readConn(productID).QueryRow(context.Background(), `
		SELECT GREATEST(
			COALESCE(COUNT(r.id), 0)::float * COALESCE(AVG(r.grade), 3.0)
			+ COALESCE((SELECT SUM(quantity)::float FROM shelf_product WHERE product_id = $1::uuid), 0) * 0.5,
			1.0
		)
		FROM review r
		WHERE r.product_id = $1::uuid
	`, productID).Scan(&rate)
	if err != nil {
		return 1.0, nil
	}
	return rate, nil
}

func (d *DBTurnoverProvider) GetDemandVariability(productID string) (float64, error) {
	var cv float64
	err := d.readConn(productID).QueryRow(context.Background(), `
		SELECT COALESCE(STDDEV(grade::float) / NULLIF(AVG(grade::float), 0), 0.25)
		FROM review
		WHERE product_id = $1::uuid
	`, productID).Scan(&cv)
	if err != nil {
		return 0.25, nil
	}
	if cv == 0 {
		return 0.25, nil
	}
	return cv, nil
}

func main() {
	cfg := config.MustLoad()
	log := pkg.SetupLogger(cfg.Env)

	pool := pkg.GetPgConnectionPool(cfg.PgStorage, log)
	redisClient := pkg.GetRedisConnectionPool(cfg.RedisStorage, log)

	rep := repo.NewRepo(pool, redisClient, log)

	optimizer := service.NewOptimizationService(
		rep.Shelf,
		cfg.Optimizer,
		log,
		&DBTurnoverProvider{pool: pool},
	)

	srv := server.NewRestAPI(&cfg.HTTP, rep, log, optimizer)
	srvClose := srv.CreateServer()
	defer srvClose()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)
	<-stop
}
