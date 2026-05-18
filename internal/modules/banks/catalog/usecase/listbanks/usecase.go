package listbanks

import (
	"context"

	"project/internal/modules/banks/catalog/domain"
)

type UseCase struct {
	repo domain.Repository
}

func New(repo domain.Repository) *UseCase {
	return &UseCase{repo: repo}
}

func (uc *UseCase) Execute(ctx context.Context, input Input) (*Output, error) {
	banks, err := uc.repo.List(ctx)
	if err != nil {
		return nil, err
	}

	out := &Output{
		Banks: make([]BankOutput, 0, len(banks)),
	}
	for _, bank := range banks {
		out.Banks = append(out.Banks, BankOutput{
			ID:        bank.ID,
			Code:      bank.Code,
			Name:      bank.Name,
			CreatedAt: bank.CreatedAt,
			UpdatedAt: bank.UpdatedAt,
		})
	}

	return out, nil
}
