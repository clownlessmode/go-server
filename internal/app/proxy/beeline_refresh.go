package proxy

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const (
	beelineAPIHost         = "https://api.beeline.ru"
	beelineRefreshTimeout  = 30 * time.Second
	beelineMonthPeriodDays = 29
)

func (s *Service) refreshBeelineAfterPayment(ctx context.Context, req *http.Request) {
	if s.beelineRepo == nil {
		return
	}

	s.captureBeelineSession(req)

	simNumber := s.beelineSimForProxy(ctx)
	if simNumber == "" {
		proxyLog.Warnf("beeline refresh skipped: active sim unknown")
		return
	}

	headers := s.beelineSessionHeaders()
	if len(headers) == 0 {
		proxyLog.Warnf("beeline refresh skipped: no session headers for sim=%s", simNumber)
		if _, ok := s.computeBeelineBalanceFromSnapshot(ctx, simNumber); ok {
			proxyLog.Infof("beeline refresh balance recomputed from snapshot: sim=%s", simNumber)
		}
		return
	}

	periodStart, periodEnd := beelineMonthPeriod(time.Now())

	if err := s.fetchBeelineDetalization(ctx, simNumber, headers, periodStart, periodEnd); err != nil {
		proxyLog.Warnf("beeline refresh detalization failed: sim=%s err=%v", simNumber, err)
	}

	if err := s.fetchBeelineBalanceMain(ctx, simNumber, headers); err != nil {
		proxyLog.Warnf("beeline refresh balance request failed: sim=%s err=%v", simNumber, err)
	}

	if balance, ok := s.computeBeelineBalanceFromSnapshot(ctx, simNumber); ok {
		proxyLog.Infof("beeline refresh completed: sim=%s balance=%.2f", simNumber, balance)
		return
	}

	proxyLog.Infof("beeline refresh completed: sim=%s (snapshot balance unavailable)", simNumber)
}

func (s *Service) fetchBeelineDetalization(
	ctx context.Context,
	simNumber string,
	headers map[string]string,
	periodStart, periodEnd time.Time,
) error {
	query := url.Values{}
	query.Set("periodStart", formatBeelinePeriodQueryParam(periodStart))
	query.Set("periodEnd", formatBeelinePeriodQueryParam(periodEnd))

	target := beelineAPIHost + beelineDetalizationPath + "?" + query.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return err
	}

	applyBeelineSessionHeaders(req, headers)
	req.Header.Set("Accept-Encoding", "gzip")

	resp, err := s.beelineRefreshClient().Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	rawBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	response, _, _, err := readBeelineJSONResponse(rawBody, resp.Header.Get("Content-Encoding"))
	if err != nil {
		return err
	}
	if response == nil {
		return fmt.Errorf("empty detalization response")
	}

	baseData, ok := response["data"].(map[string]any)
	if !ok {
		return fmt.Errorf("detalization response has no data")
	}

	finalBalance, err := s.prepareBeelineDetalizationFromBaseData(
		ctx,
		simNumber,
		baseData,
		periodStart,
		periodEnd,
		true,
	)
	if err != nil {
		return err
	}

	proxyLog.Infof(
		"beeline refresh detalization: sim=%s period=%s..%s balance=%.2f",
		simNumber,
		periodStart.In(beelineDetalizationLocation).Format("2006-01-02"),
		periodEnd.In(beelineDetalizationLocation).Format("2006-01-02"),
		finalBalance,
	)

	return nil
}

func (s *Service) fetchBeelineBalanceMain(ctx context.Context, simNumber string, headers map[string]string) error {
	target := beelineAPIHost + beelineMainBalancePath
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return err
	}

	applyBeelineSessionHeaders(req, headers)
	req.Header.Set("Accept-Encoding", "gzip")

	resp, err := s.beelineRefreshClient().Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	rawBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	response, _, _, err := readBeelineJSONResponse(rawBody, resp.Header.Get("Content-Encoding"))
	if err != nil {
		return err
	}

	data, _ := response["data"].(map[string]any)
	balanceValue := float64(0)
	if parsed := jsonNumberFromAny(data["balanceValue"]); parsed != nil {
		balanceValue = *parsed
	}

	proxyLog.Infof("beeline refresh balance/main: sim=%s apiBalance=%.2f", simNumber, balanceValue)

	return nil
}

func (s *Service) beelineRefreshClient() *http.Client {
	if s.beelineRefreshHTTPClient != nil {
		return s.beelineRefreshHTTPClient
	}

	s.beelineRefreshHTTPClient = &http.Client{Timeout: beelineRefreshTimeout}
	return s.beelineRefreshHTTPClient
}

func beelineMonthPeriod(now time.Time) (time.Time, time.Time) {
	now = now.In(beelineDetalizationLocation)
	end := beelineEndOfDay(now)
	start := beelineStartOfDay(now.AddDate(0, 0, -beelineMonthPeriodDays))

	return start.UTC(), end.UTC()
}

func formatBeelinePeriodQueryParam(value time.Time) string {
	return value.In(beelineDetalizationLocation).Format("2006-01-02 15:04:05")
}

func (s *Service) prepareBeelineDetalizationFromBaseData(
	ctx context.Context,
	simNumber string,
	baseData map[string]any,
	periodStart, periodEnd time.Time,
	saveBaseline bool,
) (float64, error) {
	if _, err := s.beelineRepo.EnsureSim(ctx, simNumber); err != nil {
		return 0, err
	}

	_, finalBalance, err := s.buildBeelineDetalizationView(
		ctx,
		simNumber,
		baseData,
		periodStart,
		periodEnd,
	)
	if err != nil {
		return 0, err
	}

	if saveBaseline {
		if err := s.saveBeelineDetalizationBaseline(
			ctx,
			simNumber,
			baseData,
			periodStart,
			periodEnd,
			finalBalance,
		); err != nil {
			return 0, err
		}
	}

	return finalBalance, nil
}
