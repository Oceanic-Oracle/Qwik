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
	Width       *float64
	Height      *float64
	Depth       *float64
	Weight      *float64
	Volume      *float64
	CreatedAt   time.Time
	Visibility  bool
}

type ProductWithAVG struct {
	Id          string
	PreviewURL  string
	Name        string
	Description string
	Price       int64
	Width       *float64
	Height      *float64
	Depth       *float64
	Weight      *float64
	Volume      *float64
	CreatedAt   time.Time
	Visibility  bool
	Count       int64
	Avg         float64
}

type CreateProduct struct {
	PreviewURL  string
	Name        string
	Description string
	Price       int64
	Width       *float64
	Height      *float64
	Depth       *float64
	Weight      *float64
	Volume      *float64
	Visibility  bool
}