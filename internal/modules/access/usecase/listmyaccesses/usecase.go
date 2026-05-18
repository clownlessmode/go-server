package listmyaccesses

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
	accesses, err := uc.repo.ListByUserID(ctx, input.UserID)
	if err != nil {
		return nil, err
	}

	now := uc.now().UTC()
	out := &Output{
		Accesses: make([]AccessOutput, 0, len(accesses)),
	}
	for _, access := range accesses {
		out.Accesses = append(out.Accesses, AccessOutput{
			ID:           access.ID,
			UserID:       access.UserID,
			BankID:       access.BankID,
			BankCode:     access.BankCode,
			BankName:     access.BankName,
			GrantedAt:    access.GrantedAt,
			ExpiresAt:    access.ExpiresAt,
			RevokedAt:    access.RevokedAt,
			RevokeReason: access.RevokeReason,
			IsActive:     access.IsActive(now),
		})
	}

	return out, nil
}
