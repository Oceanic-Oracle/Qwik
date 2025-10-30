package profile

import (
	"auth/pkg"
	"context"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type profile struct {
	writeConn func(s string) (*pgxpool.Pool, pkg.ShardNum)
	readConn func(s string) (*pgxpool.Pool, pkg.ShardNum)
	allReadConn func() []*pgxpool.Pool

	log       *slog.Logger
}

func (p *profile) GetProfile(ctx context.Context, login string) (*GetProfileRes, error) {
	q := `SELECT surname, name, patronymic, created_at FROM users WHERE login = $1`

	conn, shardNum := p.readConn(login)
	p.log.Debug("read from a shard", "num", shardNum)

	res := &GetProfileRes{}
	var ctime time.Time
	if err := conn.QueryRow(ctx, q, login).Scan(&res.Surname, &res.Name, &res.Patronymic, &ctime); err != nil {
		return nil, err
	}
	res.CreatedAt = ctime.Format("02.01.2006")

	return res, nil
}

func NewProfileRepo(writeConn func(s string) (*pgxpool.Pool, pkg.ShardNum),
	readConn func(s string) (*pgxpool.Pool, pkg.ShardNum),
	allReadConn func() []*pgxpool.Pool,
	log *slog.Logger) ProfileInterface {
	return &profile{
		writeConn: writeConn,
		readConn: readConn,
		allReadConn: allReadConn,
		log:     log,
	}
}
