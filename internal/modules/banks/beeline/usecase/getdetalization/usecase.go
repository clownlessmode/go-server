package getdetalization

import (
	"context"
	"errors"
	"time"

	"project/internal/modules/banks/beeline/detalization"
	"project/internal/modules/banks/beeline/domain"
)

type Input struct {
	Number string
}

type Output struct {
	Number      string
	PeriodStart time.Time
	PeriodEnd   time.Time
	Balance     *float64
	Data        map[string]any
	CreatedAt   time.Time
	UpdatedAt   time.Time
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

	snapshot, err := uc.repo.GetDetalizationSnapshot(ctx, input.Number)
	if errors.Is(err, domain.ErrDetalizationSnapshotNotFound) {
		return nil, domain.ErrDetalizationSnapshotNotFound
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

	hiddenNet := detalization.HiddenTransactionsNetChange(baseData, hiddenIDs)

	viewData, balanceValue, err := detalization.BuildView(baseData, payments, hiddenIDs, nil, input.Number, time.Now().UTC())
	if err != nil {
		return nil, err
	}

	balance := domain.RoundMoney(balanceValue)
	allOutgoing, err := uc.repo.SumPaymentTotals(ctx, input.Number)
	if err != nil {
		return nil, err
	}
	allIncoming, err := uc.repo.SumIncomingTotals(ctx, input.Number)
	if err != nil {
		return nil, err
	}
	if synced, ok := detalization.SyncDisplayBalance(viewData, snapshot.APIBalance, allOutgoing, allIncoming, hiddenNet); ok {
		balance = synced
	} else if display := domain.DisplayBalanceFromAPI(snapshot.APIBalance, allOutgoing, allIncoming, hiddenNet); display != nil {
		balance = *display
	}

	return &Output{
		Number:      sim.Number,
		PeriodStart: snapshot.PeriodStart,
		PeriodEnd:   snapshot.PeriodEnd,
		Balance:     &balance,
		Data:        viewData,
		CreatedAt:   sim.CreatedAt,
		UpdatedAt:   sim.UpdatedAt,
	}, nil
}
