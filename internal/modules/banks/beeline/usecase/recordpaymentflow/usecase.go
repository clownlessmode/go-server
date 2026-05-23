package recordpaymentflow

import (
	"context"
	"time"

	"project/internal/modules/banks/beeline/domain"
)

type Input struct {
	SimNumber    string
	ReceiverCard string
	Amount       float64
	Commission   float64
	PaidAt       time.Time
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
	if _, err := uc.repo.EnsureSim(ctx, input.SimNumber); err != nil {
		return nil, err
	}

	payment, err := domain.NewPaymentFlowPayment(
		input.ReceiverCard,
		input.Amount,
		input.Commission,
		input.PaidAt,
	)
	if err != nil {
		return nil, err
	}

	created, err := uc.repo.CreatePayment(ctx, input.SimNumber, payment)
	if err != nil {
		return nil, err
	}

	return &Output{Payment: created}, nil
}
