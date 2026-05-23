package getsim

import (
	"context"

	"project/internal/modules/banks/beeline/domain"
)

type Input struct {
	Number string
}

type Output struct {
	Sim domain.Sim
}

type UseCase struct {
	repo domain.Repository
}

func New(repo domain.Repository) *UseCase {
	return &UseCase{repo: repo}
}

func (uc *UseCase) Execute(ctx context.Context, input Input) (*Output, error) {
	sim, err := uc.repo.GetSim(ctx, input.Number)
	if err != nil {
		return nil, err
	}

	return &Output{Sim: sim}, nil
}
