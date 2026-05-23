package proxy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const (
	beelineCardViewTermURLPath     = "/mobile/api/v1/mcpt/cardViewTermUrl"
	beelineClarifyPath             = "/mobile/api/mcpt/v1/catalog/transactions/clarify"
	beelineFinancialConditionsPath = "/mobile/api/mcpt/v1/catalog/transactions/financialConditions"
	beelineCommissionRate          = 0.065
	beelineSenderCardTemplate      = "beelinecatalog"
	beelineReceiverCardTemplate    = "beelinecatalogreceiver"
)

type beelinePaymentSnapshot struct {
	SenderCard   string
	ReceiverCard string
	Amount       *float64
	Commission   *float64
}

func (p *beelinePaymentSnapshot) finalize() beelinePaymentSnapshot {
	out := *p
	if out.Amount == nil && out.Commission != nil {
		amount := beelineAmountFromCommission(*out.Commission)
		out.Amount = &amount
	}

	return out
}

func beelineAmountFromCommission(commission float64) float64 {
	return math.Round(commission / beelineCommissionRate)
}

func (p *beelinePaymentSnapshot) totalAmount() *float64 {
	snapshot := p.finalize()
	if snapshot.Amount == nil || snapshot.Commission == nil {
		return nil
	}

	total := *snapshot.Amount + *snapshot.Commission
	return &total
}

func (s *Service) captureBeelinePaymentRequest(req *http.Request) {
	if isBeelineHost(req.Host) {
		s.captureBeelineAPIPaymentRequest(req)
		return
	}

	if isRuruHost(req.Host) {
		s.captureRuruCardTemplate(req)
	}
}

func (s *Service) captureBeelineAPIPaymentRequest(req *http.Request) {
	path := pathForLog(req)
	switch {
	case req.Method == http.MethodGet && path == beelineCardViewTermURLPath:
		maskedAccount := req.URL.Query().Get("masked_account")
		if maskedAccount == "" {
			return
		}

		decoded, err := url.QueryUnescape(maskedAccount)
		if err != nil {
			decoded = maskedAccount
		}

		card := formatBeelineCardDisplay(decoded)
		s.beelinePaymentMu.Lock()
		template := s.beelinePendingCardTemplate
		s.beelinePaymentMu.Unlock()

		s.updateBeelinePaymentSnapshot(func(p *beelinePaymentSnapshot) {
			switch template {
			case beelineSenderCardTemplate:
				p.SenderCard = card
			default:
				p.ReceiverCard = card
			}
		})
	case req.Method == http.MethodPut && path == beelineClarifyPath:
		body, err := readAndRestoreRequestBody(req)
		if err != nil || len(body) == 0 {
			return
		}

		s.mergeBeelinePaymentBody(body)
	case req.Method == http.MethodPost && path == beelineFinancialConditionsPath:
		body, err := readAndRestoreRequestBody(req)
		if err != nil || len(body) == 0 {
			return
		}

		s.mergeBeelinePaymentBody(body)
	}
}

func (s *Service) captureRuruCardTemplate(req *http.Request) {
	if req.Method != http.MethodGet || pathForLog(req) != "/secure/card/view" {
		return
	}

	template := strings.TrimSpace(req.URL.Query().Get("template"))
	if template == "" {
		return
	}

	s.beelinePaymentMu.Lock()
	s.beelinePendingCardTemplate = template
	s.beelinePaymentMu.Unlock()
}

func (s *Service) captureBeelinePaymentResponse(req *http.Request, res *http.Response) {
	if !isBeelineHost(req.Host) || res.StatusCode != http.StatusOK {
		return
	}
	if req.Method != http.MethodPost || pathForLog(req) != beelineFinancialConditionsPath {
		return
	}

	rawBody, err := io.ReadAll(res.Body)
	if err != nil {
		return
	}
	if err := res.Body.Close(); err != nil {
		proxyLog.Warnf("beeline payment context response close failed: err=%v", err)
	}
	res.Body = io.NopCloser(bytes.NewReader(rawBody))

	responseBody := responseBodyForLog(rawBody, res.Header.Get("Content-Encoding"))
	var parsed struct {
		Data struct {
			Commission *float64 `json:"commission"`
		} `json:"data"`
	}
	if err := json.Unmarshal([]byte(responseBody), &parsed); err != nil || parsed.Data.Commission == nil {
		return
	}

	s.updateBeelinePaymentSnapshot(func(p *beelinePaymentSnapshot) {
		p.Commission = parsed.Data.Commission
		if p.Amount == nil {
			amount := beelineAmountFromCommission(*parsed.Data.Commission)
			p.Amount = &amount
		}
	})
}

func (s *Service) prepareBeelineSMSPreview(transactionBody []byte) {
	s.mergeBeelinePaymentBody(transactionBody)
	s.logBeelineSMSPreview()
	s.sendBeelinePaymentSMS()
}

func (s *Service) mergeBeelinePaymentBody(body []byte) {
	if len(body) == 0 {
		return
	}

	details := parseBeelinePaymentBody(body)
	s.updateBeelinePaymentSnapshot(func(p *beelinePaymentSnapshot) {
		if details.SenderCard != "" {
			p.SenderCard = details.SenderCard
		}
		if details.ReceiverCard != "" {
			p.ReceiverCard = details.ReceiverCard
		}
		if details.Amount != nil {
			p.Amount = details.Amount
		}
		if details.Commission != nil {
			p.Commission = details.Commission
		}
	})
}

func (s *Service) updateBeelinePaymentSnapshot(update func(*beelinePaymentSnapshot)) {
	s.beelinePaymentMu.Lock()
	defer s.beelinePaymentMu.Unlock()

	update(&s.beelinePaymentContext)
}

func (s *Service) logBeelineSMSPreview() {
	s.beelinePaymentMu.Lock()
	snapshot := s.beelinePaymentContext.finalize()
	s.beelinePaymentMu.Unlock()

	proxyLog.Infof(`
════════════════════════════════════════
  BEELINE · SMS PREVIEW
════════════════════════════════════════
  Карта отправителя:  %s
  Карта получателя:   %s
  Сумма:              %s
  Комиссия:           %s
  Итого:              %s
────────────────────────────────────────
  SMS отправляется через ADB (если SMS_ENABLED=true)
════════════════════════════════════════`,
		displayOrDash(snapshot.SenderCard),
		displayOrDash(snapshot.ReceiverCard),
		formatBeelineAmount(snapshot.Amount),
		formatBeelineMoney(snapshot.Commission),
		formatBeelineMoney(snapshot.totalAmount()),
	)
}

func parseBeelinePaymentBody(body []byte) beelinePaymentSnapshot {
	var snapshot beelinePaymentSnapshot

	var raw map[string]any
	if err := json.Unmarshal(body, &raw); err != nil {
		return snapshot
	}

	if amount := jsonNumberFromMap(raw, "amount"); amount != nil {
		snapshot.Amount = amount
	}
	if commission := jsonNumberFromMap(raw, "commission"); commission != nil {
		snapshot.Commission = commission
	}
	if total := jsonNumberFromMap(raw, "totalAmount", "total", "amountTotal"); total != nil {
		if snapshot.Amount == nil && snapshot.Commission != nil {
			amount := *total - *snapshot.Commission
			snapshot.Amount = &amount
		}
	}

	for key, value := range raw {
		switch strings.ToLower(key) {
		case "cardnumber", "receivercard", "receivercardnumber", "recipientcard":
			if card := stringFromAny(value); card != "" {
				snapshot.ReceiverCard = formatBeelineCardDisplay(card)
			}
		case "sendercard", "sendercardnumber", "sourcecard":
			if card := stringFromAny(value); card != "" {
				snapshot.SenderCard = formatBeelineCardDisplay(card)
			}
		}
	}

	fields, ok := raw["fields"].([]any)
	if !ok {
		return snapshot
	}

	for _, item := range fields {
		field, ok := item.(map[string]any)
		if !ok {
			continue
		}

		id := strings.ToLower(stringFromAny(field["id"]))
		value := field["value"]
		switch id {
		case "amount":
			if amount := jsonNumberFromAny(value); amount != nil {
				snapshot.Amount = amount
			}
		case "cardnumber", "receivercard", "recipientcard":
			if card := stringFromAny(value); card != "" {
				snapshot.ReceiverCard = formatBeelineCardDisplay(card)
			}
		case "sendercard", "sourcecard":
			if card := stringFromAny(value); card != "" {
				snapshot.SenderCard = formatBeelineCardDisplay(card)
			}
		}
	}

	return snapshot
}

func readAndRestoreRequestBody(req *http.Request) ([]byte, error) {
	if req.Body == nil {
		return nil, nil
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}

	req.Body = io.NopCloser(bytes.NewReader(body))
	return body, nil
}

func jsonNumberFromMap(raw map[string]any, keys ...string) *float64 {
	for _, key := range keys {
		for mapKey, value := range raw {
			if !strings.EqualFold(mapKey, key) {
				continue
			}
			if number := jsonNumberFromAny(value); number != nil {
				return number
			}
		}
	}

	return nil
}

func jsonNumberFromAny(value any) *float64 {
	switch typed := value.(type) {
	case float64:
		return &typed
	case json.Number:
		if parsed, err := typed.Float64(); err == nil {
			return &parsed
		}
	case string:
		normalized := strings.ReplaceAll(strings.TrimSpace(typed), " ", "")
		normalized = strings.ReplaceAll(normalized, ",", ".")
		if normalized == "" {
			return nil
		}
		if parsed, err := strconv.ParseFloat(normalized, 64); err == nil {
			return &parsed
		}
	}

	return nil
}

func stringFromAny(value any) string {
	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed)
	case json.Number:
		return typed.String()
	case float64:
		return strconv.FormatFloat(typed, 'f', -1, 64)
	default:
		return ""
	}
}

func formatBeelineCardDisplay(card string) string {
	card = strings.TrimSpace(card)
	if card == "" {
		return ""
	}

	if strings.Contains(card, "*") || strings.Contains(card, " ") {
		return card
	}

	digits := strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, card)
	if len(digits) < 13 {
		return card
	}

	var parts []string
	for i := 0; i < len(digits); i += 4 {
		end := i + 4
		if end > len(digits) {
			end = len(digits)
		}
		parts = append(parts, digits[i:end])
	}

	return strings.Join(parts, " ")
}

func formatBeelineMoney(value *float64) string {
	if value == nil {
		return "—"
	}

	return fmt.Sprintf("%.2f ₽", *value)
}

func formatBeelineAmount(value *float64) string {
	if value == nil {
		return "—"
	}

	if math.Mod(*value, 1) == 0 {
		return fmt.Sprintf("%.0f ₽", *value)
	}

	return fmt.Sprintf("%.2f ₽", *value)
}

func displayOrDash(value string) string {
	if strings.TrimSpace(value) == "" {
		return "—"
	}

	return value
}
