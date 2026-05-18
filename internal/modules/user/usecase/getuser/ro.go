package getuser

import "time"

type Output struct {
	ID        int64
	Login     string
	Password  string
	Role      string
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}
