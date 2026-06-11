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

var beelineActiveSimCapturePaths = map[string]struct{}{
	beelineUserInfoPath:       {},
	beelineProductsPath:       {},
	beelineProfileContextPath: {},
	beelineSlaveAccountsPath:  {},
}

func (s *Service) captureBeelineActiveSim(req *http.Request, res *http.Response) {
	if s.beelineRepo == nil || !isBeelineHost(req.Host) || res.StatusCode != http.StatusOK || req.Method != http.MethodGet {
		return
	}

	path := pathForLog(req)
	if _, ok := beelineActiveSimCapturePaths[path]; !ok {
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

	if requestCTN := beelineCTNFromRequest(req); requestCTN != "" {
		capture := beelineSimCapture{Preferred: requestCTN, CTNs: []string{requestCTN}}
		proxyLog.Infof(
			"beeline sim capture: route=%s source=request-header preferred=%s activeBefore=%s",
			path,
			requestCTN,
			s.activeBeelineSim(),
		)
		s.addBeelineSessionCTNs(capture.CTNs)
		s.resolveAndSetActiveBeelineSim(req.Context(), capture.Preferred, path+"/request")
		res.Body = io.NopCloser(bytes.NewReader(rawBody))
		return
	}

	capture := extractBeelineSimCapture(path, body)
	proxyLog.Infof(
		"beeline sim capture: route=%s preferred=%q ctns=%v activeBefore=%s",
		path,
		capture.Preferred,
		capture.CTNs,
		s.activeBeelineSim(),
	)
	if len(capture.CTNs) == 0 && (path == beelineProfileContextPath || path == beelineSlaveAccountsPath) {
		logBeelineSimCaptureDebug(path, body)
	}

	if len(capture.CTNs) > 0 {
		s.addBeelineSessionCTNs(capture.CTNs)
	}
	if path == beelineProductsPath {
		s.setBeelineProductCTNs(capture.CTNs)
	}

	preferred := capture.Preferred
	if preferred == "" && path == beelineUserInfoPath {
		phoneRaw, ctnRaw := extractBeelineUserInfoRaw(body)
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

	if preferred != "" || len(capture.CTNs) > 0 {
		s.resolveAndSetActiveBeelineSim(req.Context(), preferred, path)
	}

	res.Body = io.NopCloser(bytes.NewReader(rawBody))
}

func extractBeelineUserInfoRaw(body []byte) (phoneRaw, ctnRaw string) {
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
		return "", ""
	}

	return strings.TrimSpace(payload.Data.Contract.Phone.Number), strings.TrimSpace(payload.Data.Contract.CTN)
}

func (s *Service) addBeelineSessionCTNs(numbers []string) {
	s.beelineSimMu.Lock()
	defer s.beelineSimMu.Unlock()

	s.beelineSessionCTNs = mergeBeelineCTNs(s.beelineSessionCTNs, numbers)
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
			"beeline sim resolve skipped: source=%s preferred=%s active=%s session=%v products=%v",
			source,
			preferred,
			s.activeBeelineSim(),
			s.currentBeelineSessionCTNs(),
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
}

func (s *Service) pickConfiguredBeelineSim(ctx context.Context, preferred string) (string, string, bool) {
	candidates := s.beelineSimCandidates(preferred)
	active := s.activeBeelineSim()
	products := s.currentBeelineProductCTNs()
	session := s.currentBeelineSessionCTNs()
	s.logBeelineSimCandidateConfig(ctx, "pick", preferred, active, session, candidates)

	if len(candidates) == 0 {
		proxyLog.Warnf(
			"beeline sim pick: no candidates preferred=%q active=%q session=%v products=%v",
			preferred,
			active,
			session,
			products,
		)
		return "", "", false
	}

	if configured, ok := s.beelineRepo.FindConfiguredSimAmong(ctx, candidates); ok {
		reason, _ := s.simConfigurationReason(ctx, configured)
		return configured, "configured:" + reason, true
	}

	for _, candidate := range candidates {
		inDB, err := s.simExistsInDB(ctx, candidate)
		if err != nil {
			proxyLog.Warnf("beeline sim pick: db lookup failed sim=%s err=%v", candidate, err)
			continue
		}
		if !inDB {
			continue
		}

		reason := "db-candidate"
		if candidate == preferred {
			reason = "db-preferred"
		}
		return candidate, reason, true
	}

	if len(session) > 0 {
		if sim, ok := s.singleRegisteredBeelineSim(ctx); ok {
			proxyLog.Warnf(
				"beeline sim pick: using single registered sim=%s beeline session=%v candidates=%v",
				sim,
				session,
				candidates,
			)
			return sim, "single-registered-override", true
		}
	}

	proxyLog.Warnf(
		"beeline sim pick: no configured or registered candidates preferred=%q active=%q session=%v products=%v candidates=%v",
		preferred,
		active,
		session,
		products,
		candidates,
	)
	return "", "", false
}

func (s *Service) singleRegisteredBeelineSim(ctx context.Context) (string, bool) {
	if s.beelineRepo == nil {
		return "", false
	}

	sims, err := s.beelineRepo.ListSims(ctx)
	if err != nil || len(sims) != 1 {
		return "", false
	}

	number := sims[0].Number
	inDB, err := s.simExistsInDB(ctx, number)
	if err != nil || !inDB {
		return "", false
	}

	return number, true
}

func (s *Service) logBeelineSimCandidateConfig(
	ctx context.Context,
	stage, preferred, active string,
	session, candidates []string,
) {
	if s.beelineRepo == nil {
		return
	}

	labels := make([]string, 0, len(candidates))
	for _, number := range candidates {
		labels = append(labels, s.simCandidateLabel(ctx, number))
	}

	proxyLog.Infof(
		"beeline sim candidates: stage=%s preferred=%q active=%q session=%v candidates=%v config=[%s]",
		stage,
		preferred,
		active,
		session,
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

func (s *Service) currentBeelineSessionCTNs() []string {
	s.beelineSimMu.Lock()
	defer s.beelineSimMu.Unlock()

	return append([]string(nil), s.beelineSessionCTNs...)
}

func (s *Service) beelineSimCandidates(preferred string) []string {
	s.beelineSimMu.Lock()
	sessionCTNs := append([]string(nil), s.beelineSessionCTNs...)
	s.beelineSimMu.Unlock()

	seen := make(map[string]struct{}, len(sessionCTNs)+1)
	candidates := make([]string, 0, len(sessionCTNs)+1)

	if preferred != "" {
		if number, ok := normalizeBeelineSimNumber(preferred); ok {
			candidates = append(candidates, number)
			seen[number] = struct{}{}
		}
	}

	for _, number := range sessionCTNs {
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
		return ""
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
