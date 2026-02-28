package review

import "context"

type ReviewInterface interface {
	GetReviews(context.Context, string) ([]*Review, error)
}