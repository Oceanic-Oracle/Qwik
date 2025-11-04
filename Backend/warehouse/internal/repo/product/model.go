package product

import (
	"time"

	"github.com/google/uuid"
)

type Product struct {
	Id          uuid.UUID
	Preview_url string
	Name        string
	Description string
	Price       int
	Created_at  time.Time
	Visibility  bool
}
