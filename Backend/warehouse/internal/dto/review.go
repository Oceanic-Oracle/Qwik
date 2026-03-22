package dto

import "time"

type Review struct {
	Id          int       `json:"id"`
	Login       *string   `json:"login,omitempty"`
	Grade       int       `json:"grade"`
	Description string    `json:"description"`
	ProductID   string    `json:"product_id"`
	CreatedAt   time.Time `json:"created_at"`
}
