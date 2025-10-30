package auth

import "context"

type AuthInterface interface {
	Create(context.Context, *CreateUserReq) (*CreateUserRes, error)
	GetUser(context.Context, string) (*GetUserRes, error)
}