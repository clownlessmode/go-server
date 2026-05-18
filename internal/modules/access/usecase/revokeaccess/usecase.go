package revokeaccess

import (
	"context"
	"time"

	"project/internal/modules/access/domain"
)

type UseCase struct {
	repo domain.Repository
	now  func() time.Time
}

func New(repo domain.Repository) *UseCase {
	return &UseCase{
		repo: repo,
		now:  time.Now,
	}
}

func (uc *UseCase) Execute(ctx context.Context, input Input) (*Output, error) {
	access, err := uc.repo.Revoke(ctx, input.UserID, input.BankID, input.RevokeReason)
	if err != nil {
		return nil, err
	}

	return &Output{
		ID:           access.ID,
		UserID:       access.UserID,
		BankID:       access.BankID,
		BankCode:     access.BankCode,
		BankName:     access.BankName,
		GrantedAt:    access.GrantedAt,
		ExpiresAt:    access.ExpiresAt,
		GrantReason:  access.GrantReason,
		RevokedAt:    access.RevokedAt,
		RevokeReason: access.RevokeReason,
		IsActive:     access.IsActive(uc.now().UTC()),
	}, nil
}
