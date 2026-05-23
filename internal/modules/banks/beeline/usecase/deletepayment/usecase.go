package deletepayment

import (
	"context"

	"project/internal/modules/banks/beeline/domain"
)

type Input struct {
	Number string
	ID     string
}

type UseCase struct {
	repo domain.Repository
}

func New(repo domain.Repository) *UseCase {
	return &UseCase{repo: repo}
}

func (uc *UseCase) Execute(ctx context.Context, input Input) error {
	return uc.repo.DeletePayment(ctx, input.Number, input.ID)
}
