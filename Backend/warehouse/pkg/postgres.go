package pkg

import (
    "context"
    "log/slog"
    "strings"
    "time"
    "warehouse/internal/config"

    "github.com/jackc/pgx/v5/pgxpool"
)

type ShardNum uint8

type pair struct {
    WriteNode *pgxpool.Pool
    ReadNode  *pgxpool.Pool
}

type ConnectionPool map[ShardNum]*pair

func GetPgConnectionPool(cfg config.PgStorage, log *slog.Logger) *ConnectionPool {
    urlsWrite := strings.Split(cfg.PostgresUrlsWrite, ",")
    urlsRead := strings.Split(cfg.PostgresUrlsRead, ",")
    pool := make(ConnectionPool, len(urlsWrite))

    for i := 0; i < len(urlsWrite); i++ {
        // Создаем отдельный контекст с таймаутом для каждой итерации цикла,
        // чтобы отмена одного не ломала создание следующих пулов
        ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

        connWrite, err := loneConn(ctx, urlsWrite[i], log)
        if err != nil {
            cancel()
            panic(err)
        }

        connRead, err := loneConn(ctx, urlsRead[i], log)
        if err != nil {
            cancel()
            panic(err)
        }

        // Контекст можно отменять сразу после успешного создания пула,
        // так как pgxpool управляет своим жизненным циклом независимо
        cancel()

        pool[ShardNum(i)] = &pair{
            WriteNode: connWrite,
            ReadNode:  connRead,
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