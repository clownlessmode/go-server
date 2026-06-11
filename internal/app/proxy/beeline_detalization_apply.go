package proxy

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"project/internal/modules/banks/beeline/detalization"
	beelinedomain "project/internal/modules/banks/beeline/domain"
)

func cloneDetalizationData(data map[string]any) (map[string]any, error) {
	return detalization.CloneData(data)
}

func decodeDetalizationSnapshotData(raw []byte) (map[string]any, error) {
	return detalization.DecodeSnapshotData(raw)
}

func encodeDetalizationSnapshotData(data map[string]any) ([]byte, error) {
	raw, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("marshal detalization snapshot: %w", err)
	}

	return raw, nil
}

func (s *Service) buildBeelineDetalizationView(
	ctx context.Context,
	simNumber string,
	baseData map[string]any,
	periodStart, periodEnd time.Time,
	now time.Time,
) (map[string]any, float64, error) {
	payments, err := s.beelineRepo.ListPaymentsInPeriod(ctx, simNumber, periodStart, periodEnd)
	if err != nil {
		return nil, 0, err
	}

	hiddenIDs, err := s.beelineHiddenTransactionIDs(ctx, simNumber)
	if err != nil {
		return nil, 0, err
	}

	hiddenNet := detalization.HiddenTransactionsNetChange(baseData, hiddenIDs)

	view, balance, err := detalization.BuildView(baseData, payments, hiddenIDs, nil, simNumber, now)
	if err != nil {
		return nil, 0, err
	}

	if synced, ok := s.syncBeelineDetalizationDisplayBalance(ctx, simNumber, view, hiddenNet); ok {
		balance = synced
	}

	for _, summary := range detalization.BalanceReturnTransactionSummaries(view) {
		proxyLog.Infof("beeline detalization balance_return tx: sim=%s %s", simNumber, summary)
	}

	return view, balance, nil
}

func (s *Service) syncBeelineDetalizationDisplayBalance(ctx context.Context, simNumber string, data map[string]any, hiddenNetChange float64) (float64, bool) {
	outgoing, err := s.beelineRepo.SumPaymentTotals(ctx, simNumber)
	if err != nil {
		return 0, false
	}

	incoming, err := s.beelineRepo.SumIncomingTotals(ctx, simNumber)
	if err != nil {
		return 0, false
	}

	var apiBalance *float64
	snapshot, err := s.beelineRepo.GetDetalizationSnapshot(ctx, simNumber)
	if err == nil {
		apiBalance = snapshot.APIBalance
	}

	return detalization.SyncDisplayBalance(data, apiBalance, outgoing, incoming, hiddenNetChange)
}

func (s *Service) saveBeelineDetalizationBaseline(
	ctx context.Context,
	simNumber string,
	baseData map[string]any,
	periodStart, periodEnd time.Time,
	computedBalance float64,
) error {
	hiddenIDs, err := s.beelineHiddenTransactionIDs(ctx, simNumber)
	if err != nil {
		return err
	}

	raw, err := encodeDetalizationSnapshotData(baseData)
	if err != nil {
		return err
	}

	balance := beelinedomain.RoundMoney(computedBalance)
	if _, err := s.beelineRepo.SaveDetalizationSnapshot(ctx, beelinedomain.DetalizationSnapshot{
		SimNumber:       simNumber,
		PeriodStart:     periodStart,
		PeriodEnd:       periodEnd,
		Data:            raw,
		ComputedBalance: &balance,
	}); err != nil {
		return err
	}

	proxyLog.Infof(
		"beeline detalization monthly saved: sim=%s period=%s..%s balance=%.2f transactions=%d snapshotBytes=%d hidden=%d",
		simNumber,
		periodStart.Format("2006-01-02"),
		periodEnd.Format("2006-01-02"),
		balance,
		detalization.CountReportTransactions(baseData),
		len(raw),
		len(hiddenIDs),
	)

	return nil
}

func (s *Service) beelineHiddenTransactionIDs(ctx context.Context, simNumber string) ([]string, error) {
	hiddenIDs, err := s.beelineRepo.ListHiddenTransactionIDs(ctx, simNumber)
	if errors.Is(err, beelinedomain.ErrSimNotFound) {
		return []string{}, nil
	}
	return hiddenIDs, err
}
