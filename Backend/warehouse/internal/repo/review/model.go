package review

import "time"

type Review struct {
	Id          string
	Login       string
	Grade       int
	Description string
	CreatedAt   time.Time
}
