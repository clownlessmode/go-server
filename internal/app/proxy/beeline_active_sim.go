package proxy

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
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
		productCTNs := extractBeelineMobileCTNs(body)
		proxyLog.Infof(
			"beeline sim capture: route=%s products=%v activeBefore=%s",
			path,
			productCTNs,
			s.activeBeelineSim(),
		)
		s.setBeelineProductCTNs(productCTNs)
		s.resolveAndSetActiveBeelineSim(req.Context(), "", path)
	case beelineUserInfoPath:
		preferred, phoneRaw, ctnRaw, ok := extractBeelinePreferredSimFromUserInfo(body)
		if !ok {
			proxyLog.Warnf(
				"beeline sim capture: route=%s preferred unavailable phone=%q ctn=%q activeBefore=%s",
				path,
				phoneRaw,
				ctnRaw,
				s.activeBeelineSim(),
			)
			res.Body = io.NopCloser(bytes.NewReader(rawBody))
			return
		}
		proxyLog.Infof(
			"beeline sim capture: route=%s preferred=%s phone=%q ctn=%q activeBefore=%s",
			path,
			preferred,
			phoneRaw,
			ctnRaw,
			s.activeBeelineSim(),
		)
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
	number, reason, ok := s.pickConfiguredBeelineSim(ctx, preferred)
	if !ok {
		proxyLog.Warnf(
			"beeline sim resolve skipped: source=%s preferred=%s active=%s products=%v",
			source,
			preferred,
			s.activeBeelineSim(),
			s.currentBeelineProductCTNs(),
		)
		return
	}

	s.setActiveBeelineSim(number, source)
	proxyLog.Infof(
		"beeline sim resolved: sim=%s reason=%s source=%s preferred=%s activeNow=%s",
		number,
		reason,
		source,
		preferred,
		s.activeBeelineSim(),
	)
	// Auto-create SIM disabled: add SIMs manually via POST /banks/beeline/sims.
	// if _, err := s.beelineRepo.EnsureSim(ctx, number); err != nil {
	// 	proxyLog.Warnf("beeline sim ensure failed: number=%s err=%v", number, err)
	// }
}

func (s *Service) pickConfiguredBeelineSim(ctx context.Context, preferred string) (string, string, bool) {
	candidates := s.beelineSimCandidates(preferred)
	active := s.activeBeelineSim()
	products := s.currentBeelineProductCTNs()
	s.logBeelineSimCandidateConfig(ctx, "pick", preferred, active, products, candidates)

	if len(candidates) == 0 {
		proxyLog.Warnf(
			"beeline sim pick: no candidates preferred=%q active=%q products=%v",
			preferred,
			active,
			products,
		)
		return "", "", false
	}

	if configured, ok := s.beelineRepo.FindConfiguredSimAmong(ctx, candidates); ok {
		reason, _ := s.simConfigurationReason(ctx, configured)
		return configured, "configured:" + reason, true
	}

	if preferred != "" {
		if number, ok := normalizeBeelineSimNumber(preferred); ok {
			inDB, _ := s.simExistsInDB(ctx, number)
			return number, fmt.Sprintf("fallback-preferred inDb=%t", inDB), true
		}
	}

	inDB, _ := s.simExistsInDB(ctx, candidates[0])
	return candidates[0], fmt.Sprintf("fallback-first-product inDb=%t", inDB), true
}

func (s *Service) logBeelineSimCandidateConfig(
	ctx context.Context,
	stage, preferred, active string,
	products, candidates []string,
) {
	if s.beelineRepo == nil {
		return
	}

	labels := make([]string, 0, len(candidates))
	for _, number := range candidates {
		labels = append(labels, s.simCandidateLabel(ctx, number))
	}

	proxyLog.Infof(
		"beeline sim candidates: stage=%s preferred=%q active=%q products=%v candidates=%v config=[%s]",
		stage,
		preferred,
		active,
		products,
		candidates,
		strings.Join(labels, ", "),
	)
}

func (s *Service) simCandidateLabel(ctx context.Context, number string) string {
	inDB, dbErr := s.simExistsInDB(ctx, number)
	hasSnapshot, snapshotErr := s.beelineRepo.HasDetalizationSnapshot(ctx, number)
	paymentsCount := 0
	if payments, err := s.beelineRepo.ListPayments(ctx, number); err == nil {
		paymentsCount = len(payments)
	}

	return fmt.Sprintf(
		"%s:inDb=%t(dbErr=%v) snapshot=%t(snapshotErr=%v) payments=%d configured=%s",
		number,
		inDB,
		dbErr,
		hasSnapshot,
		snapshotErr,
		paymentsCount,
		s.simConfigurationLabel(ctx, number),
	)
}

func (s *Service) simExistsInDB(ctx context.Context, number string) (bool, error) {
	_, err := s.beelineRepo.GetSim(ctx, number)
	if err == nil {
		return true, nil
	}
	if err == beelinedomain.ErrSimNotFound || errors.Is(err, beelinedomain.ErrSimNotFound) {
		return false, nil
	}
	return false, err
}

func (s *Service) simConfigurationReason(ctx context.Context, number string) (string, bool) {
	if s.beelineRepo == nil {
		return "no-repo", false
	}

	if hasSnapshot, err := s.beelineRepo.HasDetalizationSnapshot(ctx, number); err == nil && hasSnapshot {
		return "snapshot", true
	}

	sim, err := s.beelineRepo.GetSim(ctx, number)
	if err == nil && sim.Balance != nil {
		return "balance", true
	}

	payments, err := s.beelineRepo.ListPayments(ctx, number)
	if err == nil && len(payments) > 0 {
		return fmt.Sprintf("payments(%d)", len(payments)), true
	}

	return "none", false
}

func (s *Service) simConfigurationLabel(ctx context.Context, number string) string {
	reason, configured := s.simConfigurationReason(ctx, number)
	if configured {
		return reason
	}
	return "none"
}

func (s *Service) currentBeelineProductCTNs() []string {
	s.beelineSimMu.Lock()
	defer s.beelineSimMu.Unlock()

	return append([]string(nil), s.beelineProductCTNs...)
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
	activeBefore := s.activeBeelineSim()
	number, reason, ok := s.pickConfiguredBeelineSim(ctx, activeBefore)
	if !ok {
		proxyLog.Warnf("beeline sim for proxy: unresolved active=%s", activeBefore)
		return activeBefore
	}

	if number != activeBefore {
		s.setActiveBeelineSim(number, "proxy-resolve")
	}
	proxyLog.Infof(
		"beeline sim for proxy: sim=%s reason=%s activeBefore=%s activeAfter=%s",
		number,
		reason,
		activeBefore,
		s.activeBeelineSim(),
	)
	return number
}

func extractBeelinePreferredSimFromUserInfo(body []byte) (normalized, phoneRaw, ctnRaw string, ok bool) {
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
		return "", "", "", false
	}

	phoneRaw = strings.TrimSpace(payload.Data.Contract.Phone.Number)
	ctnRaw = strings.TrimSpace(payload.Data.Contract.CTN)
	number := phoneRaw
	if number == "" {
		number = ctnRaw
	}

	normalized, ok = normalizeBeelineSimNumber(number)
	return normalized, phoneRaw, ctnRaw, ok
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
