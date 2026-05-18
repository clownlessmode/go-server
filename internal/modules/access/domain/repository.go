package domain

import "context"

type Repository interface {
	ListByUserID(ctx context.Context, userID int64) ([]*AccessGrant, error)
	HasActiveAccess(ctx context.Context, userID int64, bankID int64) (bool, error)
	Grant(ctx context.Context, access *AccessGrant) (*AccessGrant, error)
	Revoke(ctx context.Context, userID int64, bankID int64, reason string) (*AccessGrant, error)
}
