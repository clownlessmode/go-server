package proxy

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strconv"

	beelinedomain "project/internal/modules/banks/beeline/domain"
)

const beelineMainBalancePath = "/mobile/api/v1/balance/main"

func (s *Service) applyBeelineBalanceChangeScript(req *http.Request, res *http.Response) {
	if !isBeelineMainBalanceRequest(req, res) || s.beelineRepo == nil {
		return
	}

	simNumber := s.beelineSimForProxy(req.Context())
	if simNumber == "" {
		return
	}

	hasSnapshot, err := s.beelineRepo.HasDetalizationSnapshot(req.Context(), simNumber)
	if err != nil {
		proxyLog.Warnf("beeline balance change snapshot check failed: sim=%s err=%v", simNumber, err)
		return
	}
	if !hasSnapshot {
		return
	}

	rawBody, err := io.ReadAll(res.Body)
	if err != nil {
		proxyLog.Warnf("beeline balance change response read failed: err=%v", err)
		return
	}
	if err := res.Body.Close(); err != nil {
		proxyLog.Warnf("beeline balance change response close failed: err=%v", err)
	}

	response, originalBody, encoded, err := readBeelineJSONResponse(rawBody, res.Header.Get("Content-Encoding"))
	if err != nil {
		proxyLog.Warnf("beeline balance change failed: err=%v", err)
		res.Body = io.NopCloser(bytes.NewReader(rawBody))
		return
	}
	if response == nil {
		res.Body = io.NopCloser(bytes.NewReader(originalBody))
		return
	}

	liveBalance, hasLiveBalance := beelineMainBalanceValue(response)
	var live *float64
	if hasLiveBalance {
		if err := s.beelineRepo.UpdateDetalizationAPIBalance(req.Context(), simNumber, liveBalance); err != nil {
			proxyLog.Warnf("beeline api balance persist failed: sim=%s err=%v", simNumber, err)
		}
		value := liveBalance
		live = &value
	}

	displayBalance, ok := s.computeBeelineDisplayBalance(req.Context(), simNumber, live)
	if !ok {
		res.Body = io.NopCloser(bytes.NewReader(originalBody))
		return
	}

	if !replaceBeelineMainBalanceValue(response, displayBalance) {
		res.Body = io.NopCloser(bytes.NewReader(originalBody))
		return
	}

	changedBody, changed, err := writeBeelineJSONResponse(response, originalBody, encoded)
	if err != nil || !changed {
		res.Body = io.NopCloser(bytes.NewReader(originalBody))
		return
	}

	res.Body = io.NopCloser(bytes.NewReader(changedBody))
	res.ContentLength = int64(len(changedBody))
	res.Header.Set("Content-Length", strconv.Itoa(len(changedBody)))

	proxyLog.Infof(
		"beeline balance change applied: route=%s sim=%s balance=%.2f source=api+payments",
		pathForLog(req),
		simNumber,
		displayBalance,
	)
}

func (s *Service) computeBeelineDisplayBalance(ctx context.Context, simNumber string, liveAPIBalance *float64) (float64, bool) {
	displayBalance, ok := s.beelineDisplayBalance(ctx, simNumber, liveAPIBalance)
	if !ok {
		return 0, false
	}

	if err := s.beelineRepo.UpdateDetalizationComputedBalance(ctx, simNumber, displayBalance); err != nil {
		proxyLog.Warnf("beeline snapshot balance persist failed: sim=%s err=%v", simNumber, err)
	}

	return displayBalance, true
}

func beelineMainBalanceValue(response map[string]any) (float64, bool) {
	data, ok := response["data"].(map[string]any)
	if !ok {
		return 0, false
	}
	parsed := jsonNumberFromAny(data["balanceValue"])
	if parsed == nil {
		return 0, false
	}

	return beelinedomain.RoundMoney(*parsed), true
}

func isBeelineMainBalanceRequest(req *http.Request, res *http.Response) bool {
	return req.Method == http.MethodGet &&
		res.StatusCode == http.StatusOK &&
		isBeelineHost(req.Host) &&
		pathForLog(req) == beelineMainBalancePath
}

func replaceBeelineMainBalanceValue(response map[string]any, balance float64) bool {
	data, ok := response["data"].(map[string]any)
	if !ok {
		return false
	}
	if _, exists := data["balanceValue"]; !exists {
		return false
	}

	data["balanceValue"] = balance
	return true
}
