package repo

import (
	"crypto/sha256"
	"log/slog"
	"warehouse/internal/repo/product"
	"warehouse/pkg"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repo struct {
	Product product.ProductInterface
}

func NewRepo(connPool *pkg.ConnectionPool, log *slog.Logger) *Repo {
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
	}
}

func getShardNum(data string, num int) pkg.ShardNum {
	hash := sha256.Sum256([]byte(data))

	hashUint := uint8(hash[0])

	return pkg.ShardNum(hashUint % uint8(num))
}