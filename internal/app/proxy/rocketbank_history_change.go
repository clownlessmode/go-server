package proxy

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"project/internal/modules/banks/rocketbank/domain"
)

const (
	rocketbankHistoryPath       = "/v1/history/list"
	rocketbankHistoryTimeLayout = "2006-01-02T15:04:05-0700"
	rocketbankHistoryPeriod     = "01-2006"
)

func (s *Service) applyRocketbankHistoryChangeScript(req *http.Request, res *http.Response) {
	if !isRocketbankHistoryRequest(req, res) || s.rocketbankRepo == nil {
		return
	}

	config, err := s.rocketbankRepo.GetConfig(req.Context())
	if err != nil {
		proxyLog.Warnf("rocketbank history change config read failed: err=%v", err)
		return
	}
	if len(config.History) == 0 && len(config.HiddenHistoryIDs) == 0 {
		return
	}

	rawBody, err := io.ReadAll(res.Body)
	if err != nil {
		proxyLog.Warnf("rocketbank history change response read failed: err=%v", err)
		return
	}
	if err := res.Body.Close(); err != nil {
		proxyLog.Warnf("rocketbank history change response close failed: err=%v", err)
	}

	changedBody, changed, err := rocketbankHistoryChangedBody(rawBody, res.Header.Get("Content-Encoding"), config.History, config.ClientInfo, config.HiddenHistoryIDs)
	if err != nil {
		proxyLog.Warnf("rocketbank history change failed: err=%v", err)
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

	proxyLog.Infof("rocketbank history change applied: items=%d hidden=%d", len(config.History), len(config.HiddenHistoryIDs))
}

func isRocketbankHistoryRequest(req *http.Request, res *http.Response) bool {
	return req.Method == http.MethodPost &&
		res.StatusCode == http.StatusOK &&
		isRocketbankHost(req.Host) &&
		pathForLog(req) == rocketbankHistoryPath
}

func rocketbankHistoryChangedBody(rawBody []byte, encoding string, history []domain.HistoryItem, clientInfo domain.ClientInfo, hiddenHistoryIDs []string) ([]byte, bool, error) {
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

	var response []map[string]any
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, false, err
	}
	if len(response) == 0 {
		return nil, false, nil
	}

	response, filtered := filterRocketbankHiddenHistory(response, hiddenHistoryIDs)
	response, added := mergeRocketbankHistory(response, history, clientInfo)
	if !filtered && added == 0 {
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

func mergeRocketbankHistory(response []map[string]any, history []domain.HistoryItem, clientInfo domain.ClientInfo) ([]map[string]any, int) {
	added := 0

	for _, item := range history {
		operation, ok := rocketbankHistoryOperation(item, clientInfo)
		if !ok {
			continue
		}

		transactionTime, ok := rocketbankHistoryOperationTime(operation)
		if !ok {
			continue
		}

		period := transactionTime.Format(rocketbankHistoryPeriod)
		periodIndex := rocketbankHistoryPeriodIndex(response, period)
		if periodIndex == -1 {
			response = append(response, newRocketbankHistoryPeriod(response, period, operation))
			sortRocketbankHistoryPeriods(response)
			added++
			continue
		}

		operations, ok := response[periodIndex]["operationsList"].([]any)
		if !ok {
			continue
		}
		if rocketbankHistoryHasOperation(operations, operation) {
			continue
		}

		operations = append(operations, operation)
		sortRocketbankHistoryOperations(operations)
		response[periodIndex]["operationsList"] = operations
		added++
	}

	return response, added
}

func filterRocketbankHiddenHistory(response []map[string]any, hiddenHistoryIDs []string) ([]map[string]any, bool) {
	if len(hiddenHistoryIDs) == 0 {
		return response, false
	}

	hiddenSet := make(map[string]struct{}, len(hiddenHistoryIDs))
	for _, hiddenID := range hiddenHistoryIDs {
		hiddenID = strings.TrimSpace(hiddenID)
		if hiddenID == "" {
			continue
		}
		hiddenSet[hiddenID] = struct{}{}
	}
	if len(hiddenSet) == 0 {
		return response, false
	}

	changed := false
	filteredResponse := make([]map[string]any, 0, len(response))
	for _, period := range response {
		operations, ok := period["operationsList"].([]any)
		if !ok {
			filteredResponse = append(filteredResponse, period)
			continue
		}

		filteredOperations := make([]any, 0, len(operations))
		for _, operation := range operations {
			operationMap, ok := operation.(map[string]any)
			if !ok {
				filteredOperations = append(filteredOperations, operation)
				continue
			}
			if _, hidden := hiddenSet[domain.LegacyHistoryOperationID(operationMap)]; hidden {
				changed = true
				continue
			}
			filteredOperations = append(filteredOperations, operation)
		}

		if len(filteredOperations) == 0 {
			changed = true
			continue
		}

		updatedPeriod := make(map[string]any, len(period))
		for key, value := range period {
			updatedPeriod[key] = value
		}
		updatedPeriod["operationsList"] = filteredOperations
		filteredResponse = append(filteredResponse, updatedPeriod)
	}

	return filteredResponse, changed
}

func newRocketbankHistoryPeriod(response []map[string]any, period string, operation map[string]any) map[string]any {
	periodItem := map[string]any{
		"currencyInfo":   []any{map[string]any{"code": "RUR", "currency": "810", "symbol": "₽"}},
		"isLastPeriod":   false,
		"operationsList": []any{operation},
		"period":         period,
	}
	if len(response) == 0 {
		return periodItem
	}

	if currencyInfo, ok := cloneRocketbankHistoryValue(response[0]["currencyInfo"]); ok {
		periodItem["currencyInfo"] = currencyInfo
	}

	return periodItem
}

func rocketbankHistoryPeriodIndex(response []map[string]any, period string) int {
	for index, item := range response {
		if itemPeriod, ok := item["period"].(string); ok && itemPeriod == period {
			return index
		}
	}

	return -1
}

func sortRocketbankHistoryPeriods(response []map[string]any) {
	sort.SliceStable(response, func(i int, j int) bool {
		left, leftOK := rocketbankHistoryPeriodTime(response[i])
		right, rightOK := rocketbankHistoryPeriodTime(response[j])
		if !leftOK {
			return false
		}
		if !rightOK {
			return true
		}

		return left.After(right)
	})
}

func rocketbankHistoryPeriodTime(period map[string]any) (time.Time, bool) {
	rawPeriod, ok := period["period"].(string)
	if !ok {
		return time.Time{}, false
	}

	periodTime, err := time.Parse(rocketbankHistoryPeriod, rawPeriod)
	return periodTime, err == nil
}

func rocketbankHistoryHasOperation(operations []any, item map[string]any) bool {
	itemID := domain.LegacyHistoryOperationID(item)
	if itemID == "" {
		return false
	}

	for _, operation := range operations {
		operationMap, ok := operation.(map[string]any)
		if !ok {
			continue
		}
		operationID := domain.LegacyHistoryOperationID(operationMap)
		if operationID == itemID {
			return true
		}
	}

	return false
}

func sortRocketbankHistoryOperations(operations []any) {
	sort.SliceStable(operations, func(i int, j int) bool {
		left, leftOK := rocketbankHistoryOperationTime(operations[i])
		right, rightOK := rocketbankHistoryOperationTime(operations[j])
		if !leftOK {
			return false
		}
		if !rightOK {
			return true
		}

		return left.After(right)
	})
}

func rocketbankHistoryOperationTime(operation any) (time.Time, bool) {
	operationMap, ok := operation.(map[string]any)
	if !ok {
		return time.Time{}, false
	}

	rawTime, ok := operationMap["transactionDateTime"].(string)
	if !ok {
		return time.Time{}, false
	}

	transactionTime, err := time.Parse(rocketbankHistoryTimeLayout, rawTime)
	return transactionTime, err == nil
}

func cloneRocketbankHistoryValue(value any) (any, bool) {
	raw, err := json.Marshal(value)
	if err != nil {
		return nil, false
	}

	var clone any
	if err := json.Unmarshal(raw, &clone); err != nil {
		return nil, false
	}

	return clone, true
}
