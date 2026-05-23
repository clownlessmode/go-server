package proxy

import (
	"bytes"
	"context"
	"errors"
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

	computedBalance, ok := s.computeBeelineBalanceFromSnapshot(req.Context(), simNumber)
	if !ok {
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
	if response == nil || !replaceBeelineMainBalanceValue(response, computedBalance) {
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
		"beeline balance change applied: route=%s sim=%s balance=%.2f source=snapshot",
		pathForLog(req),
		simNumber,
		computedBalance,
	)
}

func (s *Service) computeBeelineBalanceFromSnapshot(ctx context.Context, simNumber string) (float64, bool) {
	snapshot, err := s.beelineRepo.GetDetalizationSnapshot(ctx, simNumber)
	if errors.Is(err, beelinedomain.ErrDetalizationSnapshotNotFound) {
		return 0, false
	}
	if err != nil {
		proxyLog.Warnf("beeline snapshot read failed: sim=%s err=%v", simNumber, err)
		return 0, false
	}

	baseData, err := decodeDetalizationSnapshotData(snapshot.Data)
	if err != nil {
		proxyLog.Warnf("beeline snapshot decode failed: sim=%s err=%v", simNumber, err)
		return 0, false
	}

	_, finalBalance, err := s.buildBeelineDetalizationView(
		ctx,
		simNumber,
		baseData,
		snapshot.PeriodStart,
		snapshot.PeriodEnd,
	)
	if err != nil {
		proxyLog.Warnf("beeline snapshot balance compute failed: sim=%s err=%v", simNumber, err)
		return 0, false
	}

	if err := s.beelineRepo.UpdateDetalizationComputedBalance(ctx, simNumber, finalBalance); err != nil {
		proxyLog.Warnf("beeline snapshot balance persist failed: sim=%s err=%v", simNumber, err)
	}

	return finalBalance, true
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
