package auth

import (
	pgstorage "auth/internal/storage/postgres"
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type auth struct {
	getConn func(string) (*pgxpool.Pool, pgstorage.ShardNum)
	log     *slog.Logger
}

func (u *auth) GetUser(ctx context.Context, login string) (*GetUserRes, error) {
	q := `SELECT id, password FROM users WHERE login = $1`

	conn, shardNum := u.getConn(login)
	u.log.Debug("read from a shard", "num", shardNum)

	res := &GetUserRes{}
	if err := conn.QueryRow(ctx, q, login).Scan(&res.Id, &res.Password); err != nil {
		return nil, err
	}

	return res, nil
}

func (u *auth) Create(ctx context.Context, userdto *CreateUserReq) (*CreateUserRes, error) {
	q1 := `INSERT INTO users (id, login, password, shardEmail) VALUES ($1, $2, $3, $4)`
	q2 := `INSERT INTO users_email (email) VALUES ($1) RETURNING id_users`
	q3 := `INSERT INTO usersroles (user_id, role_id) VALUES ($1, $2)`

	connLogin, shardNumLogin := u.getConn(userdto.Login)
	connEmail, shardNumEmail := u.getConn(userdto.Email)
	u.log.Debug("write to shards", "login shard num", shardNumLogin, "email shard num", shardNumEmail)

	trEmail, err := connEmail.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer trEmail.Rollback(ctx)

	var newUserId uuid.UUID
	if err := trEmail.QueryRow(ctx, q2, userdto.Email).Scan(&newUserId); err != nil {
		return nil, err
	}

	trLogin, err := connLogin.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer trLogin.Rollback(ctx)

	if _, err := trLogin.Exec(ctx, q1, newUserId, userdto.Login, userdto.Password, shardNumEmail); err != nil {
		return nil, err
	}

	if _, err := trLogin.Exec(ctx, q3, newUserId, 1); err != nil {
		return nil, err
	}

	if err := trEmail.Commit(ctx); err != nil {
		return nil, err
	}
	if err := trLogin.Commit(ctx); err != nil {
		conn, _ := u.getConn(userdto.Email)
		if _, err := conn.Exec(ctx, `DELETE FROM users_email WHERE id_users = $1`, newUserId); err != nil {
			return nil, fmt.Errorf("CRIRICAL: user creation completely failed. Manual intervention required. Main: %w, User id: %s", err, newUserId)
		}
		return nil, err
	}

	return &CreateUserRes{
		Id:       newUserId,
		Login:    userdto.Login,
		Password: userdto.Password,
	}, nil
}

func NewAuthRepo(getConn func(string) (*pgxpool.Pool, pgstorage.ShardNum),
	log *slog.Logger) AuthInterface {
	return &auth{
		getConn: getConn,
		log:     log,
	}
}
