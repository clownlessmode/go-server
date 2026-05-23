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
	sim, err := uc.repo.GetSim(ctx, input.Number)
	if err != nil {
		return nil, err
	}

	effectiveBalance, err := uc.repo.GetEffectiveBalance(ctx, input.Number)
	if err != nil {
		return nil, err
	}

	paymentsTotal, err := uc.repo.SumPaymentTotals(ctx, input.Number)
	if err != nil {
		return nil, err
	}

	return &Output{
		Number:        sim.Number,
		Balance:       effectiveBalance,
		BaseBalance:   sim.Balance,
		PaymentsTotal: paymentsTotal,
		CreatedAt:     sim.CreatedAt,
		UpdatedAt:     sim.UpdatedAt,
	}, nil
}
