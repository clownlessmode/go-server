package proxy

import (
	"net/http"
	"strings"
)

var beelineSessionSkipHeaders = map[string]struct{}{
	"host":              {},
	"content-length":    {},
	"connection":        {},
	"transfer-encoding": {},
	"proxy-connection":  {},
	"keep-alive":        {},
	"te":                {},
	"trailer":           {},
	"upgrade":           {},
}

func (s *Service) captureBeelineSession(req *http.Request) {
	if req == nil || !isBeelineHost(req.Host) {
		return
	}

	headers := cloneBeelineSessionHeaders(req.Header)
	if len(headers) == 0 {
		return
	}

	s.beelineSessionMu.Lock()
	s.beelineSessionHeaderMap = headers
	s.beelineSessionMu.Unlock()
}

func (s *Service) beelineSessionHeaders() map[string]string {
	s.beelineSessionMu.Lock()
	defer s.beelineSessionMu.Unlock()

	if len(s.beelineSessionHeaderMap) == 0 {
		return nil
	}

	copied := make(map[string]string, len(s.beelineSessionHeaderMap))
	for key, value := range s.beelineSessionHeaderMap {
		copied[key] = value
	}

	return copied
}

func cloneBeelineSessionHeaders(header http.Header) map[string]string {
	if header == nil {
		return nil
	}

	result := make(map[string]string)
	for key, values := range header {
		if len(values) == 0 {
			continue
		}
		if _, skip := beelineSessionSkipHeaders[strings.ToLower(key)]; skip {
			continue
		}

		result[key] = values[0]
	}

	return result
}

func applyBeelineSessionHeaders(req *http.Request, headers map[string]string) {
	for key, value := range headers {
		req.Header.Set(key, value)
	}
}
