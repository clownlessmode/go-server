package proxy

import (
	"encoding/json"
	"net/http"
	"strings"
)

const (
	beelineProfileContextPath = "/mobile/api/v1/profile/context"
	beelineSlaveAccountsPath  = "/mobile/api/v1/settings/sso/slaveAccounts"
)

var beelineCTNFieldNames = map[string]struct{}{
	"ctn":       {},
	"ctnfor":    {},
	"login":     {},
	"msisdn":    {},
	"number":    {},
	"phone":     {},
	"subscriber": {},
}

type beelineSimCapture struct {
	Preferred string
	CTNs      []string
}

func extractBeelineSimCapture(path string, body []byte) beelineSimCapture {
	switch path {
	case beelineUserInfoPath:
		preferred, _, _, ok := extractBeelinePreferredSimFromUserInfo(body)
		if !ok {
			return beelineSimCapture{}
		}
		return beelineSimCapture{Preferred: preferred, CTNs: []string{preferred}}
	case beelineProductsPath:
		return beelineSimCapture{CTNs: extractBeelineMobileCTNs(body)}
	case beelineProfileContextPath:
		return extractBeelineSimCaptureFromContext(body)
	case beelineSlaveAccountsPath:
		return extractBeelineSimCaptureFromSlaveAccounts(body)
	default:
		return beelineSimCapture{}
	}
}

func extractBeelineSimCaptureFromContext(body []byte) beelineSimCapture {
	var payload struct {
		Data struct {
			CTN      string `json:"ctn"`
			Contract struct {
				CTN   string `json:"ctn"`
				Phone struct {
					Number string `json:"number"`
				} `json:"phone"`
			} `json:"contract"`
			CurrentProduct struct {
				CTN string `json:"ctn"`
			} `json:"currentProduct"`
			ActiveProduct struct {
				CTN string `json:"ctn"`
			} `json:"activeProduct"`
			SelectedProduct struct {
				CTN string `json:"ctn"`
			} `json:"selectedProduct"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return beelineSimCapture{CTNs: collectBeelineCTNsFromJSON(body)}
	}

	preferred := firstNormalizedBeelineSim(
		payload.Data.SelectedProduct.CTN,
		payload.Data.ActiveProduct.CTN,
		payload.Data.CurrentProduct.CTN,
		payload.Data.Contract.Phone.Number,
		payload.Data.Contract.CTN,
		payload.Data.CTN,
	)

	capture := beelineSimCapture{
		Preferred: preferred,
		CTNs:      collectBeelineCTNsFromJSON(body),
	}
	if preferred != "" && !containsString(capture.CTNs, preferred) {
		capture.CTNs = append([]string{preferred}, capture.CTNs...)
	}

	return capture
}

func extractBeelineSimCaptureFromSlaveAccounts(body []byte) beelineSimCapture {
	var payload struct {
		Data struct {
			CurrentAccount string `json:"currentAccount"`
			ActiveAccount  string `json:"activeAccount"`
			SelectedCtn    string `json:"selectedCtn"`
			SlaveAccounts  []struct {
				CTN      string `json:"ctn"`
				Login    string `json:"login"`
				Phone    string `json:"phone"`
				IsActive bool   `json:"isActive"`
				Active   bool   `json:"active"`
				Current  bool   `json:"current"`
			} `json:"slaveAccounts"`
			Accounts []struct {
				CTN      string `json:"ctn"`
				Login    string `json:"login"`
				Phone    string `json:"phone"`
				IsActive bool   `json:"isActive"`
				Active   bool   `json:"active"`
				Current  bool   `json:"current"`
			} `json:"accounts"`
		} `json:"data"`
	}

	capture := beelineSimCapture{CTNs: collectBeelineCTNsFromJSON(body)}
	if err := json.Unmarshal(body, &payload); err != nil {
		return capture
	}

	appendAccount := func(ctn, login, phone string, active bool) {
		number := firstNormalizedBeelineSim(ctn, phone, login)
		if number == "" {
			return
		}
		if !containsString(capture.CTNs, number) {
			capture.CTNs = append(capture.CTNs, number)
		}
		if active && capture.Preferred == "" {
			capture.Preferred = number
		}
	}

	for _, account := range payload.Data.SlaveAccounts {
		appendAccount(account.CTN, account.Login, account.Phone, account.IsActive || account.Active || account.Current)
	}
	for _, account := range payload.Data.Accounts {
		appendAccount(account.CTN, account.Login, account.Phone, account.IsActive || account.Active || account.Current)
	}

	if capture.Preferred == "" {
		capture.Preferred = firstNormalizedBeelineSim(
			payload.Data.SelectedCtn,
			payload.Data.ActiveAccount,
			payload.Data.CurrentAccount,
		)
	}

	return capture
}

func collectBeelineCTNsFromJSON(body []byte) []string {
	var payload any
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil
	}

	seen := make(map[string]struct{})
	return collectBeelineCTNs(payload, seen)
}

func collectBeelineCTNs(value any, seen map[string]struct{}) []string {
	switch typed := value.(type) {
	case map[string]any:
		numbers := make([]string, 0)
		for key, nested := range typed {
			lowerKey := strings.ToLower(strings.TrimSpace(key))
			if _, ok := beelineCTNFieldNames[lowerKey]; ok {
				if number, ok := normalizeBeelineSimFromAny(nested); ok {
					numbers = appendUniqueCTN(numbers, seen, number)
				}
			}
			if lowerKey == "phone" {
				if phoneMap, ok := nested.(map[string]any); ok {
					if number, ok := normalizeBeelineSimFromAny(phoneMap["number"]); ok {
						numbers = appendUniqueCTN(numbers, seen, number)
					}
				}
			}
			numbers = append(numbers, collectBeelineCTNs(nested, seen)...)
		}
		return numbers
	case []any:
		numbers := make([]string, 0, len(typed))
		for _, item := range typed {
			numbers = append(numbers, collectBeelineCTNs(item, seen)...)
		}
		return numbers
	default:
		return nil
	}
}

func normalizeBeelineSimFromAny(value any) (string, bool) {
	raw, ok := value.(string)
	if !ok {
		return "", false
	}
	return normalizeBeelineSimNumber(strings.TrimSpace(raw))
}

func firstNormalizedBeelineSim(values ...string) string {
	for _, value := range values {
		if number, ok := normalizeBeelineSimNumber(strings.TrimSpace(value)); ok {
			return number
		}
	}
	return ""
}

func appendUniqueCTN(numbers []string, seen map[string]struct{}, number string) []string {
	if _, exists := seen[number]; exists {
		return numbers
	}
	seen[number] = struct{}{}
	return append(numbers, number)
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

var beelineRequestCTNHeaders = []string{
	"Ctn", "CTN", "X-Ctn", "X-CTN",
	"Msisdn", "MSISDN", "X-Msisdn", "X-MSISDN",
	"Login", "X-Login",
}

var beelineRequestCTNQueryKeys = []string{
	"ctn", "msisdn", "login", "phone", "subscriber",
}

func beelineCTNFromRequest(req *http.Request) string {
	if req == nil {
		return ""
	}

	for _, key := range beelineRequestCTNHeaders {
		if number, ok := normalizeBeelineSimNumber(req.Header.Get(key)); ok {
			return number
		}
	}

	for _, key := range beelineRequestCTNQueryKeys {
		if number, ok := normalizeBeelineSimNumber(req.URL.Query().Get(key)); ok {
			return number
		}
	}

	return ""
}

func logBeelineSimCaptureDebug(route string, body []byte) {
	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		proxyLog.Infof("beeline sim capture debug: route=%s decodeErr=%v body=%q", route, err, truncateBeelineLogString(string(body), 240))
		return
	}

	keys := make([]string, 0, len(payload))
	for key := range payload {
		keys = append(keys, key)
	}

	dataKeys := []string{}
	if data, ok := payload["data"].(map[string]any); ok {
		for key := range data {
			dataKeys = append(dataKeys, key)
		}
	}

	proxyLog.Infof(
		"beeline sim capture debug: route=%s keys=%v dataKeys=%v body=%q",
		route,
		keys,
		dataKeys,
		truncateBeelineLogString(string(body), 240),
	)
}

func truncateBeelineLogString(value string, limit int) string {
	value = strings.Join(strings.Fields(value), " ")
	if len(value) <= limit {
		return value
	}
	return value[:limit] + "..."
}

func mergeBeelineCTNs(existing, incoming []string) []string {
	seen := make(map[string]struct{}, len(existing)+len(incoming))
	merged := make([]string, 0, len(existing)+len(incoming))

	for _, number := range existing {
		if number == "" {
			continue
		}
		if _, exists := seen[number]; exists {
			continue
		}
		seen[number] = struct{}{}
		merged = append(merged, number)
	}

	for _, number := range incoming {
		if number == "" {
			continue
		}
		if _, exists := seen[number]; exists {
			continue
		}
		seen[number] = struct{}{}
		merged = append(merged, number)
	}

	return merged
}
