package proxy

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const rocketbankLogFile = "data/logs/rocketbank.json"

type rocketbankResponseLog struct {
	CapturedAt string `json:"capturedAt"`
	Method     string `json:"method"`
	Host       string `json:"host"`
	Route      string `json:"route"`
	Query      string `json:"query,omitempty"`
	Status     int    `json:"status"`
	Response   any    `json:"response"`
}

func (s *Service) writeRocketbankResponseLog(req *http.Request, res *http.Response) {
	rawBody, err := io.ReadAll(res.Body)
	if err != nil {
		proxyLog.Warnf("rocketbank response log read failed: route=%s err=%v", pathForLog(req), err)
		return
	}
	if err := res.Body.Close(); err != nil {
		proxyLog.Warnf("rocketbank response body close failed: route=%s err=%v", pathForLog(req), err)
	}
	res.Body = io.NopCloser(bytes.NewReader(rawBody))

	responseBody := responseBodyForLog(rawBody, res.Header.Get("Content-Encoding"))
	entry := rocketbankResponseLog{
		CapturedAt: time.Now().UTC().Format(time.RFC3339),
		Method:     req.Method,
		Host:       hostForLog(req.Host),
		Route:      pathForLog(req),
		Query:      req.URL.RawQuery,
		Status:     res.StatusCode,
		Response:   responseForLog(responseBody),
	}

	if err := appendJSONEntry(rocketbankLogFile, entry); err != nil {
		proxyLog.Warnf("rocketbank response log write failed: route=%s err=%v", entry.Route, err)
		return
	}

	proxyLog.Infof("rocketbank response saved: route=%s status=%d", entry.Route, entry.Status)
}

func isRocketbankHost(requestHost string) bool {
	host := requestHost
	if splitHost, _, err := net.SplitHostPort(requestHost); err == nil {
		host = splitHost
	}

	return strings.EqualFold(host, "dbo.rocketbank.ru")
}

func pathForLog(req *http.Request) string {
	path := req.URL.Path
	if path == "" {
		path = "/"
	}

	return path
}

func hostForLog(requestHost string) string {
	host := requestHost
	if splitHost, _, err := net.SplitHostPort(requestHost); err == nil {
		host = splitHost
	}

	return host
}

func responseBodyForLog(rawBody []byte, encoding string) string {
	if strings.EqualFold(encoding, "gzip") {
		reader, err := gzip.NewReader(bytes.NewReader(rawBody))
		if err == nil {
			defer reader.Close()

			body, err := io.ReadAll(reader)
			if err == nil {
				return string(body)
			}
		}
	}

	return string(rawBody)
}

func responseForLog(body string) any {
	if strings.TrimSpace(body) == "" {
		return nil
	}

	var response any
	if err := json.Unmarshal([]byte(body), &response); err == nil {
		return response
	}

	return body
}

func appendJSONEntry(path string, entry rocketbankResponseLog) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	entries := make([]rocketbankResponseLog, 0)
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

	return os.WriteFile(path, body, 0o644)
}
