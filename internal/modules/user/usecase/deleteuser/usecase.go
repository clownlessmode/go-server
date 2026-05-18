package deleteuser

import (
	"context"

	"project/internal/modules/user/domain"
)

type UseCase struct {
	repo domain.Repository
}

func New(repo domain.Repository) *UseCase {
	return &UseCase{repo: repo}
}

func (uc *UseCase) Execute(ctx context.Context, input Input) (*Output, error) {
	if err := uc.repo.Delete(ctx, input.ID); err != nil {
		return nil, err
	}

	return &Output{}, nil
}
