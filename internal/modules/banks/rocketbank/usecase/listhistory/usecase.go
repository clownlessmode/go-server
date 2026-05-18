package listhistory

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
	history, err := uc.repo.ListHistory(ctx)
	if err != nil {
		return nil, err
	}

	return &Output{History: history}, nil
}
