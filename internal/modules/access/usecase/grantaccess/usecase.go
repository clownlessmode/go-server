package grantaccess

import (
	"context"
	"time"

	accessdomain "project/internal/modules/access/domain"
	bankdomain "project/internal/modules/banks/catalog/domain"
	userdomain "project/internal/modules/user/domain"
)

type UseCase struct {
	accessRepo accessdomain.Repository
	bankRepo   bankdomain.Repository
	userRepo   userdomain.Repository
	now        func() time.Time
}

func New(accessRepo accessdomain.Repository, bankRepo bankdomain.Repository, userRepo userdomain.Repository) *UseCase {
	return &UseCase{
		accessRepo: accessRepo,
		bankRepo:   bankRepo,
		userRepo:   userRepo,
		now:        time.Now,
	}
}

func (uc *UseCase) Execute(ctx context.Context, input Input) (*Output, error) {
	if !input.ExpiresAt.After(uc.now().UTC()) {
		return nil, accessdomain.ErrInvalidExpiration
	}

	if _, err := uc.userRepo.GetByID(ctx, input.UserID); err != nil {
		return nil, err
	}
	if _, err := uc.bankRepo.GetByID(ctx, input.BankID); err != nil {
		return nil, err
	}

	hasActiveAccess, err := uc.accessRepo.HasActiveAccess(ctx, input.UserID, input.BankID)
	if err != nil {
		return nil, err
	}
	if hasActiveAccess {
		return nil, accessdomain.ErrAccessAlreadyExists
	}

	access, err := uc.accessRepo.Grant(ctx, &accessdomain.AccessGrant{
		UserID:      input.UserID,
		BankID:      input.BankID,
		ExpiresAt:   input.ExpiresAt,
		GrantReason: input.GrantReason,
	})
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
