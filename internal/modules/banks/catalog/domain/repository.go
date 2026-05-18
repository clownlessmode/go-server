package domain

import "context"

type Repository interface {
	List(ctx context.Context) ([]*Bank, error)
	GetByID(ctx context.Context, id int64) (*Bank, error)
}
