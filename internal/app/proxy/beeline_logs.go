package proxy

import (
	"bytes"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/AdguardTeam/gomitmproxy"
)

const (
	beelineLogFile              = "data/logs/beeline.json"
	beelineRequestBodyProp      = "beelineRequestBody"
	beelineRequestEncodingProp  = "beelineRequestEncoding"
)

var beelineLogExcludedHosts = []string{
	"report.appmetrica.yandex.net",
	"firebaselogging-pa.googleapis.com",
	"inapps.appsflyersdk.com",
	"launches.appsflyersdk.com",
	"static.beeline.ru",
	"www.google.com",
}

func isBeelineLogExcludedHost(requestHost string) bool {
	host := strings.ToLower(hostForLog(requestHost))
	for _, excluded := range beelineLogExcludedHosts {
		if host == excluded {
			return true
		}
	}

	return false
}

type beelineResponseLog struct {
	CapturedAt string `json:"capturedAt"`
	Method     string `json:"method"`
	Host       string `json:"host"`
	Route      string `json:"route"`
	Query      string `json:"query,omitempty"`
	Request    any    `json:"request,omitempty"`
	Status     int    `json:"status"`
	Response   any    `json:"response"`
}

func (s *Service) captureBeelineRequestForLog(session *gomitmproxy.Session, req *http.Request) {
	if !s.cfg.BeelineLogs || s.isMagicHost(req.Host) || isBeelineLogExcludedHost(req.Host) {
		return
	}
	if req.Body == nil {
		return
	}

	rawBody, err := io.ReadAll(req.Body)
	if err != nil {
		proxyLog.Warnf("beeline request log read failed: route=%s err=%v", pathForLog(req), err)
		return
	}
	req.Body = io.NopCloser(bytes.NewReader(rawBody))

	session.SetProp(beelineRequestBodyProp, rawBody)
	session.SetProp(beelineRequestEncodingProp, req.Header.Get("Content-Encoding"))
}

func beelineRequestForLog(session *gomitmproxy.Session) any {
	rawBody, ok := session.GetProp(beelineRequestBodyProp)
	if !ok {
		return nil
	}

	body, ok := rawBody.([]byte)
	if !ok || len(body) == 0 {
		return nil
	}

	encoding, _ := session.GetProp(beelineRequestEncodingProp)
	encodingValue, _ := encoding.(string)

	return responseForLog(responseBodyForLog(body, encodingValue))
}

func (s *Service) writeBeelineResponseLog(session *gomitmproxy.Session, req *http.Request, res *http.Response) {
	rawBody, err := io.ReadAll(res.Body)
	if err != nil {
		proxyLog.Warnf("beeline response log read failed: route=%s err=%v", pathForLog(req), err)
		return
	}
	if err := res.Body.Close(); err != nil {
		proxyLog.Warnf("beeline response body close failed: route=%s err=%v", pathForLog(req), err)
	}
	res.Body = io.NopCloser(bytes.NewReader(rawBody))

	responseBody := responseBodyForLog(rawBody, res.Header.Get("Content-Encoding"))
	entry := beelineResponseLog{
		CapturedAt: time.Now().UTC().Format(time.RFC3339),
		Method:     req.Method,
		Host:       hostForLog(req.Host),
		Route:      pathForLog(req),
		Query:      req.URL.RawQuery,
		Request:    beelineRequestForLog(session),
		Status:     res.StatusCode,
		Response:   responseForLog(responseBody),
	}

	s.beelineLogMu.Lock()
	defer s.beelineLogMu.Unlock()

	if err := appendBeelineJSONEntry(beelineLogFile, entry); err != nil {
		proxyLog.Warnf("beeline response log write failed: host=%s route=%s err=%v", entry.Host, entry.Route, err)
		return
	}

	proxyLog.Infof("proxy response saved: host=%s route=%s status=%d", entry.Host, entry.Route, entry.Status)
}

func isBeelineHost(requestHost string) bool {
	host := requestHost
	if splitHost, _, err := net.SplitHostPort(requestHost); err == nil {
		host = splitHost
	}

	return strings.EqualFold(host, "api.beeline.ru")
}

func isRuruHost(requestHost string) bool {
	host := strings.ToLower(hostForLog(requestHost))

	return strings.Contains(host, "ruru.ru")
}

func appendBeelineJSONEntry(path string, entry beelineResponseLog) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	entries := make([]beelineResponseLog, 0)
	body, err := os.ReadFile(path)
	if err == nil && strings.TrimSpace(string(body)) != "" {
		if err := json.Unmarshal(body, &entries); err != nil {
			return err
		}
	} else if err != nil && !os.IsNotExist(err) {
		return err
	}

	entries = append(entries, entry)

	body, err = json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return err
	}
	body = append(body, '\n')

	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, body, 0o644); err != nil {
		return err
	}

	return os.Rename(tmpPath, path)
}
