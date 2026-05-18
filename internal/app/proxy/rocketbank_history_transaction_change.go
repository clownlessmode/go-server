package proxy

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"project/internal/modules/banks/rocketbank/domain"
)

const rocketbankHistoryTransactionPath = "/v1/history/transaction"

func (s *Service) applyRocketbankHistoryTransactionChangeScript(req *http.Request, res *http.Response) {
	if !isRocketbankHistoryTransactionRequest(req, res) || s.rocketbankRepo == nil {
		return
	}

	transactionID := req.URL.Query().Get("transactionId")
	if transactionID == "" {
		return
	}

	item, err := s.rocketbankRepo.GetHistoryItem(req.Context(), transactionID)
	if err != nil {
		return
	}

	config, err := s.rocketbankRepo.GetConfig(req.Context())
	if err != nil {
		proxyLog.Warnf("rocketbank history transaction config read failed: err=%v", err)
		return
	}

	body, ok := rocketbankHistoryTransactionDetails(item, s.rocketbankCfg.Timezone, config.ClientInfo)
	if !ok {
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

	proxyLog.Infof("rocketbank history transaction change applied: transactionId=%s", domain.HistoryItemID(item))
}

func isRocketbankHistoryTransactionRequest(req *http.Request, res *http.Response) bool {
	return req.Method == http.MethodGet &&
		isRocketbankHost(req.Host) &&
		pathForLog(req) == rocketbankHistoryTransactionPath
}
