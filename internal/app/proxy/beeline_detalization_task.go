package proxy

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/AdguardTeam/gomitmproxy"
)

const beelineDetalizationTaskPath = "/mobile/api/v1/detalization/task"

type beelineDetalizationTaskParams struct {
	RequestID   string
	CtnFor      string
	PeriodStart time.Time
	PeriodEnd   time.Time
}

func (s *Service) rememberBeelineDetalizationTask(session *gomitmproxy.Session, req *http.Request, res *http.Response) {
	if req == nil || res == nil ||
		req.Method != http.MethodPost ||
		!isBeelineHost(req.Host) ||
		pathForLog(req) != beelineDetalizationTaskPath ||
		res.StatusCode != http.StatusOK {
		return
	}

	rawRequestBody, _ := session.GetProp(beelineRequestBodyProp)
	requestBody, _ := rawRequestBody.([]byte)
	if len(requestBody) == 0 {
		return
	}

	var payload struct {
		CtnFor      string `json:"ctnFor"`
		PeriodStart string `json:"periodStart"`
		PeriodEnd   string `json:"periodEnd"`
	}
	if err := json.Unmarshal(requestBody, &payload); err != nil {
		proxyLog.Warnf("beeline detalization task decode failed: err=%v", err)
		return
	}

	periodStart, periodEnd, ok := parseBeelineDetalizationTaskPeriod(payload.PeriodStart, payload.PeriodEnd)
	if !ok {
		return
	}

	rawResponseBody, err := io.ReadAll(res.Body)
	if err != nil {
		return
	}
	if err := res.Body.Close(); err != nil {
		proxyLog.Warnf("beeline detalization task response close failed: err=%v", err)
	}
	res.Body = io.NopCloser(bytes.NewReader(rawResponseBody))

	var response struct {
		Data struct {
			RequestID string `json:"requestId"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rawResponseBody, &response); err != nil {
		proxyLog.Warnf("beeline detalization task response decode failed: err=%v", err)
		return
	}

	requestID := strings.TrimSpace(response.Data.RequestID)
	if requestID == "" {
		return
	}

	task := beelineDetalizationTaskParams{
		RequestID:   requestID,
		CtnFor:      strings.TrimSpace(payload.CtnFor),
		PeriodStart: periodStart,
		PeriodEnd:   periodEnd,
	}

	s.beelineDetalizationTaskMu.Lock()
	if s.beelineDetalizationTasks == nil {
		s.beelineDetalizationTasks = make(map[string]beelineDetalizationTaskParams)
	}
	s.beelineDetalizationTasks[requestID] = task
	s.lastBeelineDetalizationTask = task
	s.beelineDetalizationTaskMu.Unlock()
}

func (s *Service) beelineDetalizationTaskForReport(requestID string) (beelineDetalizationTaskParams, bool) {
	requestID = strings.TrimSpace(requestID)

	s.beelineDetalizationTaskMu.Lock()
	defer s.beelineDetalizationTaskMu.Unlock()

	if requestID != "" {
		if task, ok := s.beelineDetalizationTasks[requestID]; ok {
			return task, true
		}
	}

	if s.lastBeelineDetalizationTask.RequestID != "" {
		return s.lastBeelineDetalizationTask, true
	}

	return beelineDetalizationTaskParams{}, false
}

func parseBeelineDetalizationTaskPeriod(startRaw, endRaw string) (time.Time, time.Time, bool) {
	const layout = "2006-01-02"

	startRaw = strings.TrimSpace(startRaw)
	endRaw = strings.TrimSpace(endRaw)
	if startRaw == "" || endRaw == "" {
		return time.Time{}, time.Time{}, false
	}

	start, err := time.ParseInLocation(layout, startRaw, beelineDetalizationLocation)
	if err != nil {
		return time.Time{}, time.Time{}, false
	}
	end, err := time.ParseInLocation(layout, endRaw, beelineDetalizationLocation)
	if err != nil {
		return time.Time{}, time.Time{}, false
	}

	start = beelineStartOfDay(start)
	end = beelineEndOfDay(end)
	if end.Before(start) {
		end = beelineEndOfDay(start)
	}

	return start.UTC(), end.UTC(), true
}
