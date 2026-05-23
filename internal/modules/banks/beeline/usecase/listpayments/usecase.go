package listpayments

import (
	"context"

	"project/internal/modules/banks/beeline/domain"
)

type Input struct {
	Number string
}

type Output struct {
	Payments []domain.Payment
}

type UseCase struct {
	repo domain.Repository
}

func New(repo domain.Repository) *UseCase {
	return &UseCase{repo: repo}
}

func (uc *UseCase) Execute(ctx context.Context, input Input) (*Output, error) {
	if _, err := uc.repo.GetSim(ctx, input.Number); err != nil {
		return nil, err
	}

	payments, err := uc.repo.ListPayments(ctx, input.Number)
	if err != nil {
		return nil, err
	}

	return &Output{Payments: payments}, nil
}
