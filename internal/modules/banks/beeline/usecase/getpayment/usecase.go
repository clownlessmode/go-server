package getpayment

import (
	"context"

	"project/internal/modules/banks/beeline/domain"
)

type Input struct {
	Number string
	ID     string
}

type Output struct {
	Payment domain.Payment
}

type UseCase struct {
	repo domain.Repository
}

func New(repo domain.Repository) *UseCase {
	return &UseCase{repo: repo}
}

func (uc *UseCase) Execute(ctx context.Context, input Input) (*Output, error) {
	payment, err := uc.repo.GetPayment(ctx, input.Number, input.ID)
	if err != nil {
		return nil, err
	}

	return &Output{Payment: payment}, nil
}
