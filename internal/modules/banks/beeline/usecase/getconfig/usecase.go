package getconfig

import (
	"context"
	"errors"

	"project/internal/modules/banks/beeline/detalization"
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

	snapshot, err := uc.repo.GetDetalizationSnapshot(ctx, input.Number)
	if errors.Is(err, domain.ErrDetalizationSnapshotNotFound) {
		outgoingTotal, err := uc.repo.SumPaymentTotals(ctx, input.Number)
		if err != nil {
			return nil, err
		}
		incomingTotal, err := uc.repo.SumIncomingTotals(ctx, input.Number)
		if err != nil {
			return nil, err
		}

		return &Output{
			Number:        sim.Number,
			Balance:       nil,
			PaymentsTotal: outgoingTotal,
			IncomingTotal: incomingTotal,
			CreatedAt:     sim.CreatedAt,
			UpdatedAt:     sim.UpdatedAt,
		}, nil
	}
	if err != nil {
		return nil, err
	}

	baseData, err := detalization.DecodeSnapshotData(snapshot.Data)
	if err != nil {
		return nil, err
	}

	hiddenIDs, err := uc.repo.ListHiddenTransactionIDs(ctx, input.Number)
	if err != nil {
		return nil, err
	}

	payments, err := uc.repo.ListPaymentsInPeriod(ctx, input.Number, snapshot.PeriodStart, snapshot.PeriodEnd)
	if err != nil {
		return nil, err
	}

	outgoingTotal, incomingTotal := detalization.PaymentTotals(payments)

	var balance *float64
	if computedBalance, err := detalizationBuildBalance(baseData, payments, hiddenIDs); err == nil {
		value := domain.RoundMoney(computedBalance)
		balance = &value
	}

	return &Output{
		Number:        sim.Number,
		Balance:       balance,
		PaymentsTotal: outgoingTotal,
		IncomingTotal: incomingTotal,
		CreatedAt:     sim.CreatedAt,
		UpdatedAt:     sim.UpdatedAt,
	}, nil
}

func detalizationBuildBalance(baseData map[string]any, payments []domain.Payment, hiddenIDs []string) (float64, error) {
	_, balance, err := detalization.BuildView(baseData, payments, hiddenIDs)
	return balance, err
}
