package proxy

import (
	"context"
	"time"

	beelinedomain "project/internal/modules/banks/beeline/domain"
)

func (s *Service) beelineDisplayBalance(
	ctx context.Context,
	simNumber string,
	liveAPIBalance *float64,
) (float64, bool) {
	outgoing, err := s.beelineRepo.SumPaymentTotals(ctx, simNumber)
	if err != nil {
		proxyLog.Warnf("beeline display balance outgoing failed: sim=%s err=%v", simNumber, err)
		return 0, false
	}

	incoming, err := s.beelineRepo.SumIncomingTotals(ctx, simNumber)
	if err != nil {
		proxyLog.Warnf("beeline display balance incoming failed: sim=%s err=%v", simNumber, err)
		return 0, false
	}

	apiBalance := liveAPIBalance
	if apiBalance == nil {
		snapshot, err := s.beelineRepo.GetDetalizationSnapshot(ctx, simNumber)
		if err == nil && snapshot.APIBalance != nil {
			apiBalance = snapshot.APIBalance
		}
	}

	if display := beelinedomain.DisplayBalanceFromAPI(apiBalance, outgoing, incoming); display != nil {
		return *display, true
	}

	snapshot, err := s.beelineRepo.GetDetalizationSnapshot(ctx, simNumber)
	if err != nil {
		return 0, false
	}

	baseData, err := decodeDetalizationSnapshotData(snapshot.Data)
	if err != nil {
		return 0, false
	}

	_, balance, err := s.buildBeelineDetalizationView(
		ctx,
		simNumber,
		baseData,
		snapshot.PeriodStart,
		snapshot.PeriodEnd,
		time.Now().UTC(),
	)
	if err != nil {
		return 0, false
	}

	return balance, true
}
