package product

import "context"

type ProductInterface interface {
	GetProductById(context.Context, string) (*Product, error)
	GetProducts(context.Context, *bool) ([]*Product, error)
}