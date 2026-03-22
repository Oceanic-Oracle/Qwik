package review

import "context"

type ReviewInterface interface {
	GetReviews(context.Context, string) ([]*Review, error)
	CreateReview(ctx context.Context, productID string, rev Review) (int, error)
}