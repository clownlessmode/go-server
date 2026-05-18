package updatecashtransfer

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
	current, err := uc.repo.GetHistoryItem(ctx, input.ID)
	if err != nil {
		return nil, err
	}

	cashTransfer, ok := domain.CashTransferInputFromHistoryItem(current)
	if !ok {
		return nil, domain.ErrHistoryItemNotFound
	}

	if input.Amount != nil {
		cashTransfer.Amount = domain.NormalizeHistoryAmount(*input.Amount)
	}
	if input.BalanceBefore != nil {
		cashTransfer.BalanceBefore = domain.NormalizeHistoryAmount(*input.BalanceBefore)
	}
	if input.Direction != nil {
		cashTransfer.Direction = domain.NormalizeHistoryDirection(*input.Direction)
	}
	if input.Time != nil {
		cashTransfer.Time = *input.Time
	}
	if !domain.IsValidCashTransferBalance(cashTransfer.Amount, cashTransfer.BalanceBefore, cashTransfer.Direction) {
		return nil, domain.ErrInsufficientBalance
	}

	item := domain.NewCashTransferHistoryItem(cashTransfer)
	updated, err := uc.repo.UpdateHistoryItem(ctx, input.ID, item)
	if err != nil {
		return nil, err
	}

	return &Output{Item: updated}, nil
}
