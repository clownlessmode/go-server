package domain

import "context"

type Repository interface {
	Create(ctx context.Context, user *User) (*User, error)
	List(ctx context.Context) ([]*User, error)
	GetByID(ctx context.Context, id int64) (*User, error)
	GetByLogin(ctx context.Context, login string) (*User, error)
	Update(ctx context.Context, user *User) (*User, error)
	Delete(ctx context.Context, id int64) error
}
