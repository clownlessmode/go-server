package proxy

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"

	"project/internal/modules/banks/rocketbank/domain"
)

func (s *Service) applyRocketbankCardInfoChangeScript(req *http.Request, res *http.Response) {
	if !isRocketbankCardsRequest(req, res) || s.rocketbankRepo == nil {
		return
	}

	config, err := s.rocketbankRepo.GetConfig(req.Context())
	if err != nil {
		proxyLog.Warnf("rocketbank card info change config read failed: err=%v", err)
		return
	}
	if config.ClientInfo.CardNumber == nil || strings.TrimSpace(*config.ClientInfo.CardNumber) == "" {
		return
	}

	rawBody, err := io.ReadAll(res.Body)
	if err != nil {
		proxyLog.Warnf("rocketbank card info change response read failed: err=%v", err)
		return
	}
	if err := res.Body.Close(); err != nil {
		proxyLog.Warnf("rocketbank card info change response close failed: err=%v", err)
	}

	changedBody, changed, err := rocketbankCardInfoChangedBody(rawBody, res.Header.Get("Content-Encoding"), config.ClientInfo)
	if err != nil {
		proxyLog.Warnf("rocketbank card info change failed: err=%v", err)
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

	proxyLog.Infof("rocketbank card info change applied: route=%s", pathForLog(req))
}

func isRocketbankCardsRequest(req *http.Request, res *http.Response) bool {
	return req.Method == http.MethodGet &&
		res.StatusCode == http.StatusOK &&
		isRocketbankHost(req.Host) &&
		pathForLog(req) == rocketbankCardsPath
}

func rocketbankCardInfoChangedBody(rawBody []byte, encoding string, clientInfo domain.ClientInfo) ([]byte, bool, error) {
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
	if !replaceRocketbankCardInfoFields(response, strings.TrimSpace(*clientInfo.CardNumber)) {
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

func replaceRocketbankCardInfoFields(value any, cardNumber string) bool {
	changed := false
	cardSuffix := rocketbankCardLastDigits(cardNumber)
	maskedNumber := "2200********" + cardSuffix

	switch typed := value.(type) {
	case []any:
		for _, item := range typed {
			if replaceRocketbankCardInfoFields(item, cardNumber) {
				changed = true
			}
		}
	case map[string]any:
		for key, item := range typed {
			switch key {
			case "accountNumber":
				typed[key] = cardNumber
				changed = true
			case "maskedNumber":
				typed[key] = maskedNumber
				changed = true
			case "suffix":
				typed[key] = cardSuffix
				changed = true
			default:
				if replaceRocketbankCardInfoFields(item, cardNumber) {
					changed = true
				}
			}
		}
	}

	return changed
}
