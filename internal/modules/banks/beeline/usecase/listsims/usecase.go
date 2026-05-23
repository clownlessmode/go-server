package listsims

import (
	"context"

	"project/internal/modules/banks/beeline/domain"
)

type Input struct{}

type Output struct {
	Sims []domain.Sim
}

type UseCase struct {
	repo domain.Repository
}

func New(repo domain.Repository) *UseCase {
	return &UseCase{repo: repo}
}

func (uc *UseCase) Execute(ctx context.Context, input Input) (*Output, error) {
	sims, err := uc.repo.ListSims(ctx)
	if err != nil {
		return nil, err
	}

	return &Output{Sims: sims}, nil
}
