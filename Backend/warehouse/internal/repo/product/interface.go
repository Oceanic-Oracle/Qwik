package product

import (
	"context"
	"warehouse/internal/repo/review"
)

type ProductInterface interface {
	GetProductById(context.Context, string) (*ProductWithAVG, []review.Review, error)
	GetProducts(context.Context, *bool) ([]*ProductWithAVG, error)
	CreateProduct(ctx context.Context, req *CreateProduct) (*ProductWithAVG, error)
}
