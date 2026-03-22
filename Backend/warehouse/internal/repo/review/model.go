package review

import "time"

type Review struct {
	Id          int
	Login       *string
	Grade       int
	Description string
	CreatedAt   time.Time
}
