package getconfig

import (
	"context"

	"project/internal/modules/banks/beeline/domain"
)

type UseCase struct {
	repo domain.Repository
}

func New(repo domain.Repository) *UseCase {
	return &UseCase{repo: repo}
}

func (uc *UseCase) Execute(ctx context.Context, input Input) (*Output, error) {
	config, err := uc.repo.GetConfig(ctx)
	if err != nil {
		return nil, err
	}

	return &Output{
		CreatedAt: config.CreatedAt,
		UpdatedAt: config.UpdatedAt,
	}, nil
}
