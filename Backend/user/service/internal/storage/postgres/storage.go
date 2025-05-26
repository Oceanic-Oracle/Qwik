package pgstorage

import (
	"auth/internal/config"
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ShardNum uint8

type ConnectionPool map[ShardNum]*pgxpool.Pool

func GetConnectionPool(cfg config.PgStorage, sslmode string, log *slog.Logger) *ConnectionPool {
	hosts := strings.Split(cfg.Hosts, ",")
	pool := make(ConnectionPool, len(hosts))
	
	for i := 0; i < len(hosts); i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		conn, err := loneConn(ctx, cfg.Username, cfg.Password, hosts[i], cfg.Port, cfg.Database, "disable", log)
		if err != nil {
			panic(err);
		}
		pool[ShardNum(i)] = conn
	}

	return &pool
}

func loneConn(ctx context.Context, username, password, host, port, db, sslmode string, log *slog.Logger) (*pgxpool.Pool, error) {
	conn, err := pgxpool.New(ctx, fmt.Sprintf(
		"postgresql://%s:%s@%s:%s/%s?sslmode=%s",
		username, password, host, port, db, sslmode,
	))
	if err != nil {
		log.Error("failed to connection to postgres", "err", err)
		return nil, err
	}

	if err := conn.Ping(ctx); err != nil {
        log.Error("failed to ping database",
            "err", err,
            "host", host)
        return nil, err
    }

	return conn, nil
}
