package repo

import (
	"auth/internal/repo/auth"
	"auth/internal/repo/profile"
	pgstorage "auth/internal/storage/postgres"
	"crypto/sha256"
	"log/slog"
	"math/rand/v2"

	"github.com/jackc/pgx/v5/pgxpool"
)

type сonnectionRouter struct {
	writeConn func(string) (*pgxpool.Pool, pgstorage.ShardNum)
	readConn  func(string) (*pgxpool.Pool, pgstorage.ShardNum)
}

type Repo struct {
	User    auth.AuthInterface
	Profile profile.ProfileInterface
}

func NewRepo(connPool *pgstorage.ConnectionPool, log *slog.Logger) *Repo {
	cr := &сonnectionRouter{
		writeConn: func(data string) (*pgxpool.Pool, pgstorage.ShardNum) {
			shardNum := getShardNum(data, len(*connPool))

			return (*connPool)[shardNum].Main, shardNum
		},
		readConn: func(data string) (*pgxpool.Pool, pgstorage.ShardNum) {
			shardNum := getShardNum(data, len(*connPool))

			node := (*connPool)[shardNum].Slave
			indx := rand.IntN(len(node))

			return node[indx], shardNum
		},
	}

	return &Repo{
		User:    auth.NewAuthRepo(cr.writeConn, cr.readConn, log),
		Profile: profile.NewProfileRepo(cr.writeConn, cr.readConn, log),
	}
}

func getShardNum(data string, num int) pgstorage.ShardNum {
	hash := sha256.Sum256([]byte(data))

	hashUint := uint8(hash[0])

	return pgstorage.ShardNum(hashUint % uint8(num))
}
