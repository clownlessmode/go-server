package proxy

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
)

const (
	rocketbankBalancesPath = "/v1/products/balances"
	rocketbankCardsPath    = "/v1/cards/list"
)

func (s *Service) applyRocketbankBalanceChangeScript(req *http.Request, res *http.Response) {
	if !isRocketbankBalancesRequest(req, res) || s.rocketbankRepo == nil {
		return
	}

	config, err := s.rocketbankRepo.GetConfig(req.Context())
	if err != nil {
		proxyLog.Warnf("rocketbank balance change config read failed: err=%v", err)
		return
	}
	if config.Balance == nil {
		return
	}

	rawBody, err := io.ReadAll(res.Body)
	if err != nil {
		proxyLog.Warnf("rocketbank balance change response read failed: err=%v", err)
		return
	}
	if err := res.Body.Close(); err != nil {
		proxyLog.Warnf("rocketbank balance change response close failed: err=%v", err)
	}

	changedBody, changed, err := rocketbankBalanceChangedBody(rawBody, res.Header.Get("Content-Encoding"), *config.Balance)
	if err != nil {
		proxyLog.Warnf("rocketbank balance change failed: err=%v", err)
		res.Body = io.NopCloser(bytes.NewReader(rawBody))
		return
	}
	if !changed {
		res.Body = io.NopCloser(bytes.NewReader(rawBody))
		return
	}

	res.Body = io.NopCloser(bytes.NewReader(changedBody))
	res.ContentLength = int64(len(changedBody))
	res.Header.Set("Content-Length", strconv.Itoa(len(changedBody)))

	proxyLog.Infof("rocketbank balance change applied: route=%s balance=%.2f", pathForLog(req), *config.Balance)
}

func isRocketbankBalancesRequest(req *http.Request, res *http.Response) bool {
	path := pathForLog(req)

	return req.Method == http.MethodGet &&
		res.StatusCode == http.StatusOK &&
		isRocketbankHost(req.Host) &&
		(path == rocketbankBalancesPath || path == rocketbankCardsPath)
}

func rocketbankBalanceChangedBody(rawBody []byte, encoding string, balance float64) ([]byte, bool, error) {
	body := rawBody
	encoded := false

	if strings.EqualFold(encoding, "gzip") {
		reader, err := gzip.NewReader(bytes.NewReader(rawBody))
		if err != nil {
			return nil, false, err
		}
		defer reader.Close()

		body, err = io.ReadAll(reader)
		if err != nil {
			return nil, false, err
		}
		encoded = true
	} else if strings.TrimSpace(encoding) != "" {
		return nil, false, nil
	}

	var response any
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, false, err
	}
	if !replaceRocketbankBalanceFields(response, balance) {
		return nil, false, nil
	}

	changedBody, err := json.Marshal(response)
	if err != nil {
		return nil, false, err
	}

	if !encoded {
		return changedBody, true, nil
	}

	var compressed bytes.Buffer
	writer := gzip.NewWriter(&compressed)
	if _, err := writer.Write(changedBody); err != nil {
		_ = writer.Close()
		return nil, false, err
	}
	if err := writer.Close(); err != nil {
		return nil, false, err
	}

	return compressed.Bytes(), true, nil
}

func replaceRocketbankBalanceFields(value any, balance float64) bool {
	changed := false

	switch typed := value.(type) {
	case []any:
		for _, item := range typed {
			if replaceRocketbankBalanceFields(item, balance) {
				changed = true
			}
		}
	case map[string]any:
		for key, item := range typed {
			if key == "balance" {
				typed[key] = balance
				changed = true
				continue
			}
			if replaceRocketbankBalanceFields(item, balance) {
				changed = true
			}
		}
	}

	return changed
}
