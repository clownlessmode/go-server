package listusers

import "time"

type UserOutput struct {
	ID        int64
	Login     string
	Password  string
	Role      string
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Output struct {
	Users []UserOutput
}
