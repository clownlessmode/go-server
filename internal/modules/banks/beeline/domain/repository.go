package domain

import "context"

type Repository interface {
	GetConfig(ctx context.Context) (*Config, error)
}
