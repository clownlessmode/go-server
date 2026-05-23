package proxy

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const beelineCatalogTransactionsPath = "/mobile/api/mcpt/v1/catalog/transactions"

func isBeelineCatalogTransactionsRequest(req *http.Request) bool {
	return req.Method == http.MethodPost &&
		isBeelineHost(req.Host) &&
		pathForLog(req) == beelineCatalogTransactionsPath
}

func (s *Service) maybeSpoofBeelineCatalogTransaction(req *http.Request) *http.Response {
	if !isBeelineCatalogTransactionsRequest(req) {
		return nil
	}

	body, err := readAndRestoreRequestBody(req)
	if err != nil {
		proxyLog.Warnf("beeline transaction spoof: read request body failed: err=%v", err)
		return nil
	}

	s.prepareBeelineSMSPreview(body)

	delay := randomBeelineTransactionDelay()
	proxyLog.Infof("beeline transaction spoof: intercepting %s, waiting %s before fake success", beelineCatalogTransactionsPath, delay)
	time.Sleep(delay)

	responseBody, err := json.Marshal(beelineCatalogTransactionResponse{
		Data: beelineCatalogTransactionData{
			ID:    randomBeelineTransactionID(),
			Token: newRandomHexToken(),
		},
		Meta: beelineResponseMeta{
			Code:      200,
			CodeValue: "UNKNOWN",
			Message:   "OK",
			Status:    "OK",
		},
	})
	if err != nil {
		proxyLog.Warnf("beeline transaction spoof: marshal response failed: err=%v", err)
		return nil
	}

	go s.refreshBeelineAfterPayment(context.Background(), req)

	return jsonResponse(req, responseBody)
}

type beelineCatalogTransactionResponse struct {
	Data beelineCatalogTransactionData `json:"data"`
	Meta beelineResponseMeta           `json:"meta"`
}

type beelineCatalogTransactionData struct {
	ID    int64  `json:"id"`
	Token string `json:"token"`
}

type beelineResponseMeta struct {
	CachedAt  *string `json:"cachedAt"`
	Code      int     `json:"code"`
	CodeValue string  `json:"codeValue"`
	ErrorCode *string `json:"errorCode"`
	Errors    *string `json:"errors"`
	Message   string  `json:"message"`
	Status    string  `json:"status"`
}

func randomBeelineTransactionID() int64 {
	var b [4]byte
	_, _ = rand.Read(b[:])

	return 1_999_000_000 + int64(binary.BigEndian.Uint32(b[:])%9_999_999)
}

func newRandomHexToken() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)

	return fmt.Sprintf("%x", b)
}

func randomBeelineTransactionDelay() time.Duration {
	var b [4]byte
	_, _ = rand.Read(b[:])

	ms := 1000 + int(binary.BigEndian.Uint32(b[:])%1001)
	return time.Duration(ms) * time.Millisecond
}
