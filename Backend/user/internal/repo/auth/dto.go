package auth

import "github.com/google/uuid"

type (
	CreateUserReq struct {
		Login    string
		Password string
		Email    string
	}
)

type (
	CreateUserRes struct {
		Id       uuid.UUID
		Login    string
		Password string
	}

	GetUserRes struct {
		Id       string
		Password string
	}
)
