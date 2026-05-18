package updatesbptransfer

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

	sbpTransfer, ok := domain.SBPTransferInputFromHistoryItem(current)
	if !ok {
		return nil, domain.ErrHistoryItemNotFound
	}

	if input.Amount != nil {
		sbpTransfer.Amount = domain.NormalizeHistoryAmount(*input.Amount)
	}
	if input.BalanceBefore != nil {
		sbpTransfer.BalanceBefore = domain.NormalizeHistoryAmount(*input.BalanceBefore)
	}
	if input.Direction != nil {
		sbpTransfer.Direction = domain.NormalizeHistoryDirection(*input.Direction)
	}
	if input.Time != nil {
		sbpTransfer.Time = *input.Time
	}
	if input.OperationFirstName != nil {
		sbpTransfer.OperationFirstName = *input.OperationFirstName
	}
	if input.OperationMiddleName != nil {
		sbpTransfer.OperationMiddleName = *input.OperationMiddleName
	}
	if input.OperationLastName != nil {
		sbpTransfer.OperationLastName = *input.OperationLastName
	}
	if input.BankID != nil {
		sbpTransfer.BankID = *input.BankID
	}
	if input.PhoneNumber != nil {
		sbpTransfer.PhoneNumber = *input.PhoneNumber
	}
	if !domain.IsValidCashTransferBalance(sbpTransfer.Amount, sbpTransfer.BalanceBefore, sbpTransfer.Direction) {
		return nil, domain.ErrInsufficientBalance
	}

	item := domain.NewSBPTransferHistoryItem(sbpTransfer)
	updated, err := uc.repo.UpdateHistoryItem(ctx, input.ID, item)
	if err != nil {
		return nil, err
	}

	return &Output{Item: updated}, nil
}
