package repo

import (
	"auth/internal/repo/auth"
	"auth/internal/repo/profile"
	"auth/pkg"
	"crypto/sha256"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repo struct {
	User    auth.AuthInterface
	Profile profile.ProfileInterface
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
	AllReadConn := func() []*pgxpool.Pool {
		conns := make([]*pgxpool.Pool, 0, len(*connPool))
		for _, val := range *connPool {
			conns = append(conns, val.ReadNode)
		}
		return conns
	}

	return &Repo{
		User:    auth.NewAuthRepo(writeConn, readConn, AllReadConn, log),
		Profile: profile.NewProfileRepo(writeConn, readConn, AllReadConn, log),
	}
}

func getShardNum(data string, num int) pkg.ShardNum {
	hash := sha256.Sum256([]byte(data))

	hashUint := uint8(hash[0])

	return pkg.ShardNum(hashUint % uint8(num))
}
