package proxy

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	beelinedomain "project/internal/modules/banks/beeline/domain"
)

const (
	beelineUserInfoPath = "/mobile/api/v1/profile/userInfo"
	beelineProductsPath = "/mobile/api/v1/profile/products"
)

func (s *Service) captureBeelineActiveSim(req *http.Request, res *http.Response) {
	if s.beelineRepo == nil || !isBeelineHost(req.Host) || res.StatusCode != http.StatusOK || req.Method != http.MethodGet {
		return
	}

	path := pathForLog(req)
	if path != beelineUserInfoPath && path != beelineProductsPath {
		return
	}

	rawBody, err := io.ReadAll(res.Body)
	if err != nil {
		proxyLog.Warnf("beeline active sim response read failed: route=%s err=%v", path, err)
		return
	}
	if err := res.Body.Close(); err != nil {
		proxyLog.Warnf("beeline active sim response close failed: route=%s err=%v", path, err)
	}

	body, _, err := decodeBeelineResponseBody(rawBody, res.Header.Get("Content-Encoding"))
	if err != nil || body == nil {
		proxyLog.Warnf("beeline active sim response decode failed: route=%s err=%v", path, err)
		res.Body = io.NopCloser(bytes.NewReader(rawBody))
		return
	}

	switch path {
	case beelineProductsPath:
		s.setBeelineProductCTNs(extractBeelineMobileCTNs(body))
		s.resolveAndSetActiveBeelineSim(req.Context(), "", path)
	case beelineUserInfoPath:
		preferred, ok := extractBeelinePreferredSimFromUserInfo(body)
		if !ok {
			res.Body = io.NopCloser(bytes.NewReader(rawBody))
			return
		}
		s.resolveAndSetActiveBeelineSim(req.Context(), preferred, path)
	}

	res.Body = io.NopCloser(bytes.NewReader(rawBody))
}

func (s *Service) setBeelineProductCTNs(numbers []string) {
	s.beelineSimMu.Lock()
	defer s.beelineSimMu.Unlock()

	s.beelineProductCTNs = append([]string(nil), numbers...)
}

func (s *Service) resolveAndSetActiveBeelineSim(ctx context.Context, preferred, source string) {
	number, ok := s.pickConfiguredBeelineSim(ctx, preferred)
	if !ok {
		return
	}

	s.setActiveBeelineSim(number, source)
	if _, err := s.beelineRepo.EnsureSim(ctx, number); err != nil {
		proxyLog.Warnf("beeline sim ensure failed: number=%s err=%v", number, err)
	}
}

func (s *Service) pickConfiguredBeelineSim(ctx context.Context, preferred string) (string, bool) {
	candidates := s.beelineSimCandidates(preferred)
	if len(candidates) == 0 {
		return "", false
	}

	if configured, ok := s.beelineRepo.FindConfiguredSimAmong(ctx, candidates); ok {
		if len(candidates) > 1 {
			proxyLog.Infof(
				"beeline sim resolved from config: sim=%s candidates=%v preferred=%s",
				configured,
				candidates,
				preferred,
			)
		}
		return configured, true
	}

	if preferred != "" {
		if number, ok := normalizeBeelineSimNumber(preferred); ok {
			proxyLog.Infof("beeline sim fallback to preferred: sim=%s source=%s", number, preferred)
			return number, true
		}
	}

	proxyLog.Infof("beeline sim fallback to first product: sim=%s candidates=%v", candidates[0], candidates)
	return candidates[0], true
}

func (s *Service) beelineSimCandidates(preferred string) []string {
	s.beelineSimMu.Lock()
	productCTNs := append([]string(nil), s.beelineProductCTNs...)
	s.beelineSimMu.Unlock()

	seen := make(map[string]struct{}, len(productCTNs)+1)
	candidates := make([]string, 0, len(productCTNs)+1)

	if preferred != "" {
		if number, ok := normalizeBeelineSimNumber(preferred); ok {
			candidates = append(candidates, number)
			seen[number] = struct{}{}
		}
	}

	for _, number := range productCTNs {
		if _, exists := seen[number]; exists {
			continue
		}
		candidates = append(candidates, number)
		seen[number] = struct{}{}
	}

	return candidates
}

func (s *Service) beelineSimForProxy(ctx context.Context) string {
	if number, ok := s.pickConfiguredBeelineSim(ctx, s.activeBeelineSim()); ok {
		s.setActiveBeelineSim(number, "proxy-resolve")
		return number
	}

	return s.activeBeelineSim()
}

func extractBeelinePreferredSimFromUserInfo(body []byte) (string, bool) {
	var payload struct {
		Data struct {
			Contract struct {
				CTN   string `json:"ctn"`
				Phone struct {
					Number string `json:"number"`
				} `json:"phone"`
			} `json:"contract"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return "", false
	}

	number := payload.Data.Contract.Phone.Number
	if number == "" {
		number = payload.Data.Contract.CTN
	}

	return normalizeBeelineSimNumber(number)
}

func extractBeelineMobileCTNs(body []byte) []string {
	var payload struct {
		Data struct {
			Products []struct {
				CTN  string `json:"ctn"`
				Type string `json:"type"`
			} `json:"products"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil
	}

	mobileCTNs := make([]string, 0, len(payload.Data.Products))
	seen := make(map[string]struct{}, len(payload.Data.Products))
	for _, product := range payload.Data.Products {
		if !strings.EqualFold(product.Type, "mobile") {
			continue
		}

		number, ok := normalizeBeelineSimNumber(product.CTN)
		if !ok {
			continue
		}
		if _, exists := seen[number]; exists {
			continue
		}

		seen[number] = struct{}{}
		mobileCTNs = append(mobileCTNs, number)
	}

	return mobileCTNs
}

func normalizeBeelineSimNumber(number string) (string, bool) {
	number = beelinedomain.NormalizeSimNumber(number)
	if err := beelinedomain.ValidateSimNumber(number); err != nil {
		return "", false
	}

	return number, true
}

func (s *Service) setActiveBeelineSim(number, sourceRoute string) {
	s.beelineSimMu.Lock()
	defer s.beelineSimMu.Unlock()

	if s.activeBeelineSimNumber == number {
		return
	}

	s.activeBeelineSimNumber = number
	proxyLog.Infof("beeline active sim: %s source=%s", number, sourceRoute)
}

func (s *Service) activeBeelineSim() string {
	s.beelineSimMu.Lock()
	defer s.beelineSimMu.Unlock()

	return s.activeBeelineSimNumber
}
