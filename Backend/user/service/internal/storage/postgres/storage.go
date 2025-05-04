package pgstorage

import (
	"auth/internal/config"
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ShardNum uint8

type ConnectionPool map[ShardNum]*connNode

type connNode struct {
	Main *pgxpool.Pool
	Slave []*pgxpool.Pool
}

func GetConnectionPool(cfg config.PgStorage, sslmode string, log *slog.Logger) *ConnectionPool {
	pool := make(ConnectionPool, 2)

	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		pool[0] = newConnect(ctx, cfg.Username, cfg.Password, cfg.HostShard0, cfg.Port, cfg.Database, "disable", log)
	}()
	go func() {
		defer wg.Done()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		pool[1] = newConnect(ctx, cfg.Username, cfg.Password, cfg.HostShard1, cfg.Port, cfg.Database, "disable", log)
	}()
	wg.Wait()

	return &pool
}

func newConnect(ctx context.Context, username, password, hosts, port, db, sslmode string, log *slog.Logger) (*connNode) {
	hostsArr := strings.Split(hosts, ",")
	if len(hostsArr) == 0 {
		panic("the main node is missing")
	}

	connMain, err := loneConn(ctx, username, password, hostsArr[0], port, db, sslmode, log)
	if err != nil {
		panic(err)
	}

	if len(hostsArr) == 1 {
		return &connNode{
			Main: connMain,
			Slave: []*pgxpool.Pool{connMain},
		}
	}

	res := &connNode{
		Main: connMain,
	}
	for i := 1; i < len(hostsArr); i++ {
		conn, err := loneConn(ctx, username, password, hostsArr[i], port, db, sslmode, log)
		if err != nil {
			panic(err)
		}

		res.Slave = append(res.Slave, conn)
	}

	return res
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
