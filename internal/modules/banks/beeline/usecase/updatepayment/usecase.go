package updatepayment

import (
	"context"
	"time"

	"project/internal/modules/banks/beeline/domain"
)

type Input struct {
	Number       string
	ID           string
	Direction    *domain.PaymentDirection
	ReceiverCard *string
	Amount       *float64
	PaidAt       *time.Time
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
	current, err := uc.repo.GetPayment(ctx, input.Number, input.ID)
	if err != nil {
		return nil, err
	}

	updated, err := current.ApplyUpdate(input.Direction, input.ReceiverCard, input.Amount, input.PaidAt)
	if err != nil {
		return nil, err
	}

	saved, err := uc.repo.UpdatePayment(ctx, input.Number, updated)
	if err != nil {
		return nil, err
	}

	return &Output{Payment: saved}, nil
}
