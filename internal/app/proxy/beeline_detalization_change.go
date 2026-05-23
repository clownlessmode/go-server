package proxy

import (
	"bytes"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const beelineDetalizationPath = "/mobile/api/v2/detalization"

var beelineDetalizationLocation = time.FixedZone("MSK", 3*60*60)

func (s *Service) applyBeelineDetalizationChangeScript(req *http.Request, res *http.Response) {
	if !isBeelineDetalizationRequest(req, res) || s.beelineRepo == nil {
		return
	}

	simNumber := s.beelineSimForProxy(req.Context())
	if simNumber == "" {
		return
	}

	periodStart, periodEnd, ok := parseBeelineDetalizationPeriod(req.URL.Query())
	if !ok {
		return
	}

	rawBody, err := io.ReadAll(res.Body)
	if err != nil {
		proxyLog.Warnf("beeline detalization response read failed: err=%v", err)
		return
	}
	if err := res.Body.Close(); err != nil {
		proxyLog.Warnf("beeline detalization response close failed: err=%v", err)
	}

	response, originalBody, encoded, err := readBeelineJSONResponse(rawBody, res.Header.Get("Content-Encoding"))
	if err != nil {
		proxyLog.Warnf("beeline detalization decode failed: err=%v", err)
		res.Body = io.NopCloser(bytes.NewReader(rawBody))
		return
	}
	if response == nil {
		res.Body = io.NopCloser(bytes.NewReader(rawBody))
		return
	}

	baseData, ok := response["data"].(map[string]any)
	if !ok {
		res.Body = io.NopCloser(bytes.NewReader(originalBody))
		return
	}

	if _, err := s.beelineRepo.EnsureSim(req.Context(), simNumber); err != nil {
		proxyLog.Warnf("beeline sim ensure failed: number=%s err=%v", simNumber, err)
	}

	prepPeriod := isBeelineDetalizationPrepPeriod(periodStart, periodEnd)

	viewData, finalBalance, err := s.buildBeelineDetalizationView(
		req.Context(),
		simNumber,
		baseData,
		periodStart,
		periodEnd,
	)
	if err != nil {
		proxyLog.Warnf("beeline detalization prepare failed: sim=%s err=%v", simNumber, err)
		res.Body = io.NopCloser(bytes.NewReader(originalBody))
		return
	}

	if prepPeriod {
		if err := s.saveBeelineDetalizationBaseline(
			req.Context(),
			simNumber,
			baseData,
			periodStart,
			periodEnd,
			finalBalance,
		); err != nil {
			proxyLog.Warnf("beeline detalization snapshot save failed: sim=%s err=%v", simNumber, err)
			res.Body = io.NopCloser(bytes.NewReader(originalBody))
			return
		}
	}

	response["data"] = viewData

	changedBody, wrote, err := writeBeelineJSONResponse(response, originalBody, encoded)
	if err != nil || !wrote {
		res.Body = io.NopCloser(bytes.NewReader(originalBody))
		return
	}

	res.Body = io.NopCloser(bytes.NewReader(changedBody))
	res.ContentLength = int64(len(changedBody))
	res.Header.Set("Content-Length", strconv.Itoa(len(changedBody)))

	proxyLog.Infof(
		"beeline detalization prepared: route=%s sim=%s period=%s..%s prep=%t balance=%.2f",
		pathForLog(req),
		simNumber,
		periodStart.In(beelineDetalizationLocation).Format("2006-01-02"),
		periodEnd.In(beelineDetalizationLocation).Format("2006-01-02"),
		prepPeriod,
		finalBalance,
	)
}

func isBeelineDetalizationPrepPeriod(start, end time.Time) bool {
	return end.Sub(start) >= 25*24*time.Hour
}

func isBeelineDetalizationRequest(req *http.Request, res *http.Response) bool {
	return req.Method == http.MethodGet &&
		res.StatusCode == http.StatusOK &&
		isBeelineHost(req.Host) &&
		pathForLog(req) == beelineDetalizationPath
}

func parseBeelineDetalizationPeriod(query url.Values) (time.Time, time.Time, bool) {
	const layout = "2006-01-02 15:04:05"

	startRaw := strings.TrimSpace(query.Get("periodStart"))
	endRaw := strings.TrimSpace(query.Get("periodEnd"))
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

func beelineStartOfDay(value time.Time) time.Time {
	value = value.In(beelineDetalizationLocation)
	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, beelineDetalizationLocation)
}

func beelineEndOfDay(value time.Time) time.Time {
	value = value.In(beelineDetalizationLocation)
	return time.Date(value.Year(), value.Month(), value.Day(), 23, 59, 59, int(time.Second-time.Nanosecond), beelineDetalizationLocation)
}
