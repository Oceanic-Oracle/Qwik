	package pkg

	import (
		"auth/internal/config"
		"context"
		"log/slog"
		"strings"
		"time"

		"github.com/jackc/pgx/v5/pgxpool"
	)

	type ShardNum uint8

	type pair struct {
		WriteNode *pgxpool.Pool
		ReadNode  *pgxpool.Pool
	}

	type ConnectionPool map[ShardNum]*pair

	func GetPgConnectionPool(cfg config.PgStorage, sslmode string, log *slog.Logger) *ConnectionPool {
		urlsWrite := strings.Split(cfg.PostgresUrlsWrite, ",")
		urlsRead := strings.Split(cfg.PostgresUrlsRead, ",")
		pool := make(ConnectionPool, len(urlsWrite))

		for i := 0; i < len(urlsWrite); i++ {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			connWrite, err := loneConn(ctx,urlsWrite[i], log)
			if err != nil {
				panic(err)
			}

			connRead, err := loneConn(ctx, urlsRead[i], log)
			if err != nil {
				panic(err)
			}
			pool[ShardNum(i)] = &pair{
				WriteNode: connWrite,
				ReadNode: connRead,
			}
		}

		return &pool
	}

	func loneConn(ctx context.Context, url string, log *slog.Logger) (*pgxpool.Pool, error) {
		conn, err := pgxpool.New(ctx, url)
		if err != nil {
			log.Error("failed to connection to postgres", "err", err)
			return nil, err
		}

		if err := conn.Ping(ctx); err != nil {
			log.Error("failed to ping database",
				"err", err,
				"url", url)
			return nil, err
		}

		return conn, nil
	}
