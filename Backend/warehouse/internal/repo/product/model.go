package product

import (
	"time"
)

type Product struct {
	Id          string
	PreviewURL  string
	Name        string
	Description string
	Price       int64
	CreatedAt   time.Time
	Visibility  bool
}

type ProductWithAVG struct {
	Id          string
	PreviewURL  string
	Name        string
	Description string
	Price       int64
	CreatedAt   time.Time
	Visibility  bool
	Avg         float64
}

type CreateProduct struct {
	PreviewURL  string
	Name        string
	Description string
	Price       int64
	Visibility  bool
}
