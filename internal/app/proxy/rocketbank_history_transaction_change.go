package proxy

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"project/internal/modules/banks/rocketbank/domain"
)

const rocketbankHistoryTransactionPath = "/v1/history/transaction"

func (s *Service) applyRocketbankHistoryTransactionChangeScript(req *http.Request, res *http.Response) {
	if !isRocketbankHistoryTransactionRequest(req, res) {
		return
	}
	if s.rocketbankRepo == nil {
		proxyLog.Warnf("rocketbank history transaction skipped: repo is nil transactionId=%s", req.URL.Query().Get("transactionId"))
		return
	}

	transactionID := req.URL.Query().Get("transactionId")
	if transactionID == "" {
		proxyLog.Warnf("rocketbank history transaction skipped: empty transactionId route=%s", pathForLog(req))
		return
	}

	config, err := s.rocketbankRepo.GetConfig(req.Context())
	if err != nil {
		proxyLog.Warnf("rocketbank history transaction config read failed: transactionId=%s err=%v", transactionID, err)
		return
	}
	if domain.IsHiddenHistoryID(config.HiddenHistoryIDs, transactionID) {
		blockRocketbankHistoryTransaction(res)
		proxyLog.Infof("rocketbank history transaction hidden: transactionId=%s", transactionID)
		return
	}

	item, err := s.rocketbankRepo.GetHistoryItem(req.Context(), transactionID)
	if err != nil {
		if errors.Is(err, domain.ErrHistoryItemNotFound) {
			proxyLog.Infof("rocketbank history transaction pass-through: transactionId=%s status=%d (configured item not found)", transactionID, res.StatusCode)
		} else {
			proxyLog.Warnf("rocketbank history transaction lookup failed: transactionId=%s err=%v", transactionID, err)
		}
		return
	}

	body, ok := rocketbankHistoryTransactionDetails(item, s.rocketbankCfg.Timezone, config.ClientInfo)
	if !ok {
		proxyLog.Warnf("rocketbank history transaction details build failed: transactionId=%s type=%s", transactionID, item.Type)
		return
	}

	if res.Body != nil {
		if err := res.Body.Close(); err != nil {
			proxyLog.Warnf("rocketbank history transaction response close failed: err=%v", err)
		}
	}

	rawBody, err := json.Marshal(body)
	if err != nil {
		proxyLog.Warnf("rocketbank history transaction change failed: err=%v", err)
		return
	}

	res.StatusCode = http.StatusOK
	res.Status = "200 OK"
	res.Body = io.NopCloser(bytes.NewReader(rawBody))
	res.ContentLength = int64(len(rawBody))
	if res.Header == nil {
		res.Header = make(http.Header)
	}
	res.Header.Set("Content-Type", "application/json; charset=utf-8")
	res.Header.Set("Content-Length", strconv.Itoa(len(rawBody)))
	res.Header.Del("Content-Encoding")

	proxyLog.Infof(
		"rocketbank history transaction change applied: transactionId=%s type=%s chequeAllowed=%v",
		domain.HistoryItemID(item),
		item.Type,
		rocketbankHistoryChequeAllowed(body),
	)
}

func rocketbankHistoryChequeAllowed(body map[string]any) bool {
	cheque, ok := body["cheque"].(map[string]any)
	if !ok {
		return false
	}

	allowed, ok := cheque["allowed"].(bool)
	return ok && allowed
}

func isRocketbankHistoryTransactionRequest(req *http.Request, res *http.Response) bool {
	return req.Method == http.MethodGet &&
		isRocketbankHost(req.Host) &&
		pathForLog(req) == rocketbankHistoryTransactionPath
}

func blockRocketbankHistoryTransaction(res *http.Response) {
	if res.Body != nil {
		_ = res.Body.Close()
	}

	res.StatusCode = http.StatusNotFound
	res.Status = "404 Not Found"
	res.Body = io.NopCloser(bytes.NewReader([]byte(`{"error":"history item not found"}`)))
	res.ContentLength = int64(len(`{"error":"history item not found"}`))
	if res.Header == nil {
		res.Header = make(http.Header)
	}
	res.Header.Set("Content-Type", "application/json; charset=utf-8")
	res.Header.Set("Content-Length", strconv.Itoa(len(`{"error":"history item not found"}`)))
	res.Header.Del("Content-Encoding")
}
