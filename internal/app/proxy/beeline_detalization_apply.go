package proxy

import (
	"context"
	"encoding/json"
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
) (map[string]any, float64, error) {
	payments, err := s.beelineRepo.ListPaymentsInPeriod(ctx, simNumber, periodStart, periodEnd)
	if err != nil {
		return nil, 0, err
	}

	hiddenIDs, err := s.beelineRepo.ListHiddenTransactionIDs(ctx, simNumber)
	if err != nil {
		return nil, 0, err
	}

	return detalization.BuildView(baseData, payments, hiddenIDs)
}

func (s *Service) saveBeelineDetalizationBaseline(
	ctx context.Context,
	simNumber string,
	baseData map[string]any,
	periodStart, periodEnd time.Time,
	computedBalance float64,
) error {
	storedData := baseData
	hiddenIDs, err := s.beelineRepo.ListHiddenTransactionIDs(ctx, simNumber)
	if err != nil {
		return err
	}
	if len(hiddenIDs) > 0 {
		purgedData, balance, err := detalization.PurgeHiddenFromData(baseData, hiddenIDs)
		if err != nil {
			return err
		}
		storedData = purgedData
		if balance != nil {
			computedBalance = *balance
		}
	}

	raw, err := encodeDetalizationSnapshotData(storedData)
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
