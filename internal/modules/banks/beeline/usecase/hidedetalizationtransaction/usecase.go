package hidedetalizationtransaction

import (
	"context"
	"errors"

	"project/internal/modules/banks/beeline/domain"
)

type Input struct {
	Number        string
	TransactionID string
}

type Output struct{}

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

	if payment, err := uc.repo.GetPayment(ctx, input.Number, input.TransactionID); err == nil && payment.ID != "" {
		return nil, domain.ErrCannotHidePaymentTransaction
	} else if err != nil && !errors.Is(err, domain.ErrPaymentNotFound) {
		return nil, err
	}

	if err := uc.repo.HideTransaction(ctx, input.Number, input.TransactionID); err != nil {
		return nil, err
	}

	return &Output{}, nil
}
