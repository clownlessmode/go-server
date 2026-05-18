package gethistoryitem

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
	item, err := uc.repo.GetHistoryItem(ctx, input.ID)
	if err != nil {
		return nil, err
	}

	return &Output{Item: item}, nil
}
