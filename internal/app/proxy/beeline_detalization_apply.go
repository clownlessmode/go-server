package proxy

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	beelinedomain "project/internal/modules/banks/beeline/domain"
)

func cloneDetalizationData(data map[string]any) (map[string]any, error) {
	raw, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("marshal detalization data: %w", err)
	}

	cloned := make(map[string]any)
	if err := json.Unmarshal(raw, &cloned); err != nil {
		return nil, fmt.Errorf("unmarshal detalization data: %w", err)
	}

	return cloned, nil
}

func decodeDetalizationSnapshotData(raw []byte) (map[string]any, error) {
	data := make(map[string]any)
	if err := json.Unmarshal(raw, &data); err != nil {
		return nil, fmt.Errorf("unmarshal detalization snapshot: %w", err)
	}

	return data, nil
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
) (map[string]any, float64, error) {
	working, err := cloneDetalizationData(baseData)
	if err != nil {
		return nil, 0, err
	}

	payments, err := s.beelineRepo.ListPaymentsInPeriod(ctx, simNumber, periodStart, periodEnd)
	if err != nil {
		return nil, 0, err
	}

	finalBalance, ok := applyBeelineDetalizationPayments(working, payments)
	if !ok {
		return nil, 0, fmt.Errorf("recalculate beeline detalization balances")
	}

	return working, finalBalance, nil
}

func (s *Service) saveBeelineDetalizationBaseline(
	ctx context.Context,
	simNumber string,
	baseData map[string]any,
	periodStart, periodEnd time.Time,
	computedBalance float64,
) error {
	raw, err := encodeDetalizationSnapshotData(baseData)
	if err != nil {
		return err
	}

	balance := beelinedomain.RoundMoney(computedBalance)
	_, err = s.beelineRepo.SaveDetalizationSnapshot(ctx, beelinedomain.DetalizationSnapshot{
		SimNumber:       simNumber,
		PeriodStart:     periodStart,
		PeriodEnd:       periodEnd,
		Data:            raw,
		ComputedBalance: &balance,
	})

	return err
}
