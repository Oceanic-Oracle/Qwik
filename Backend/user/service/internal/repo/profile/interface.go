package profile

import "context"

type ProfileInterface interface {
	GetProfile(context.Context, string) (*GetProfileRes, error)
}
