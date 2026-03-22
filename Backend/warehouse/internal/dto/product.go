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
	CreatedAt   time.Time `json:"created_at"`
	Visibility  bool      `json:"visibility"`
	Count       int64     `json:"count"`
	Avg         float64   `json:"avg"`
	Reviews     []Review  `json:"reviews,omitempty"`
}
