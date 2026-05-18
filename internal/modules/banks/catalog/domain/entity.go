package domain

import (
	"time"
)

type Bank struct {
	ID        int64
	Code      string
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}
