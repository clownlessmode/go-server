package createcashtransfer

import (
	"context"

	"project/internal/modules/banks/rocketbank/domain"
)

type UseCase struct {
	repo domain.Repository
}

func New(repo domain.Repository) *UseCase {
	return &UseCase{repo: repo}
}

func (uc *UseCase) Execute(ctx context.Context, input Input) (*Output, error) {
	item := domain.NewCashTransferHistoryItem(domain.CashTransferInput{
		Amount:        domain.NormalizeHistoryAmount(input.Amount),
		BalanceBefore: domain.NormalizeHistoryAmount(input.BalanceBefore),
		Direction:     domain.NormalizeHistoryDirection(input.Direction),
		Time:          input.Time,
	})

	created, err := uc.repo.CreateHistoryItem(ctx, item)
	if err != nil {
		return nil, err
	}

	return &Output{Item: created}, nil
}
