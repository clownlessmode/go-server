package deletehistoryitem

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
	if err := uc.repo.DeleteHistoryItem(ctx, input.ID); err != nil {
		return nil, err
	}

	return &Output{}, nil
}
