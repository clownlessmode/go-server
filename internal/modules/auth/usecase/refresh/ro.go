package refresh

import "time"

type Output struct {
	User         UserOutput
	AccessToken  string
	RefreshToken string
}

type UserOutput struct {
	ID        int64
	Login     string
	Password  string
	Role      string
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}
