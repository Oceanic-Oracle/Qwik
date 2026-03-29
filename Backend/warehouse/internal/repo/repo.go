package repo

import (
    "crypto/sha256"
    "log/slog"
    "warehouse/internal/repo/product"
    "warehouse/internal/repo/review"
    "warehouse/internal/repo/shelf"
    "warehouse/pkg"

    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/redis/go-redis/v9"
)

type Repo struct {
    Product product.ProductInterface
    Review  review.ReviewInterface
    Shelf   *shelf.ShelfRepo
}

func NewRepo(connPool *pkg.ConnectionPool, redisClient *redis.Client, log *slog.Logger) *Repo {
    writeConn := func(s string) (*pgxpool.Pool, pkg.ShardNum) {
        num := getShardNum(s, len(*connPool))
        return (*connPool)[num].WriteNode, num
    }
    readConn := func(s string) (*pgxpool.Pool, pkg.ShardNum) {
        num := getShardNum(s, len(*connPool))
        return (*connPool)[num].ReadNode, num
    }
    allReadConn := func() []*pgxpool.Pool {
        conns := make([]*pgxpool.Pool, 0, len(*connPool))
        for _, val := range *connPool {
            conns = append(conns, val.ReadNode)
        }
        return conns
    }

    return &Repo{
        Product: product.NewProduct(writeConn, readConn, allReadConn, log),
        Review:  review.NewReview(writeConn, readConn, allReadConn, log),
        // ИСПРАВЛЕНИЕ: Передаем Redis-клиент в репо полок
        Shelf:   shelf.NewShelfRepo(writeConn, readConn, allReadConn, redisClient, log),
    }
}

func getShardNum(data string, num int) pkg.ShardNum {
    hash := sha256.Sum256([]byte(data))
    hashUint := uint8(hash[0])
    return pkg.ShardNum(hashUint % uint8(num))
}