package updatecardtransfer

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

	cardTransfer, ok := domain.CardTransferInputFromHistoryItem(current)
	if !ok {
		return nil, domain.ErrHistoryItemNotFound
	}

	if input.Amount != nil {
		cardTransfer.Amount = domain.NormalizeHistoryAmount(*input.Amount)
	}
	if input.BalanceBefore != nil {
		cardTransfer.BalanceBefore = domain.NormalizeHistoryAmount(*input.BalanceBefore)
	}
	if input.Direction != nil {
		cardTransfer.Direction = domain.NormalizeHistoryDirection(*input.Direction)
	}
	if input.Time != nil {
		cardTransfer.Time = *input.Time
	}
	if input.BankID != nil {
		cardTransfer.BankID = *input.BankID
	}
	if input.RecipientCardNumber != nil {
		cardTransfer.RecipientCardNumber = *input.RecipientCardNumber
	}
	if !domain.IsValidCashTransferBalance(cardTransfer.Amount, cardTransfer.BalanceBefore, cardTransfer.Direction) {
		return nil, domain.ErrInsufficientBalance
	}

	item := domain.NewCardTransferHistoryItem(cardTransfer)
	updated, err := uc.repo.UpdateHistoryItem(ctx, input.ID, item)
	if err != nil {
		return nil, err
	}

	return &Output{Item: updated}, nil
}
