package createpayment

import (
	"context"
	"time"

	"project/internal/modules/banks/beeline/domain"
)

type Input struct {
	Number       string
	Direction    domain.PaymentDirection
	ReceiverCard string
	Amount       float64
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
	if _, err := uc.repo.GetSim(ctx, input.Number); err != nil {
		return nil, err
	}

	payment, err := domain.NewManualPayment(input.Direction, input.ReceiverCard, input.Amount, input.PaidAt)
	if err != nil {
		return nil, err
	}

	created, err := uc.repo.CreatePayment(ctx, input.Number, payment)
	if err != nil {
		return nil, err
	}

	return &Output{Payment: created}, nil
}
