package dto

import (
	"time"
)

type Product struct {
	Id          string    `json:"id"`
	PreviewUrl  string    `json:"preview_url,omitempty"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Price       int64     `json:"price"`
	Width       *float64  `json:"width,omitempty"`
	Height      *float64  `json:"height,omitempty"`
	Depth       *float64  `json:"depth,omitempty"`
	Weight      *float64  `json:"weight,omitempty"`
	Volume      *float64  `json:"volume,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	Visibility  bool      `json:"visibility"`
	Count       int64     `json:"count"`
	Avg         float64   `json:"avg"`
	Reviews     []Review  `json:"reviews,omitempty"`
}