package proxy

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	beelinedomain "project/internal/modules/banks/beeline/domain"
	"project/internal/modules/banks/beeline/detalization"
	beelinedetalization "project/internal/modules/banks/beeline/infrastructure/detalization"
)

const (
	beelineDetalizationReportPath   = "/mobile/api/v1/detalization/report"
	beelineDetalizationReportPDFDir = "data/reports/beeline/cheques"
)

var beelineDetalizationReportService = beelinedetalization.NewService()

func (s *Service) applyBeelineDetalizationReportScript(req *http.Request, res *http.Response) bool {
	if !isBeelineDetalizationReportRequest(req) {
		return false
	}

	params, ok := s.buildBeelineDetalizationReportParams(req.Context(), req)
	if !ok {
		proxyLog.Warnf("beeline detalization report params unavailable: route=%s", beelineDetalizationReportPath)
		return false
	}

	body, err := beelineDetalizationReportService.GenerateReportPDF(params)
	if err != nil {
		proxyLog.Warnf("beeline detalization report pdf failed: err=%v", err)
		return false
	}

	s.saveBeelineDetalizationReportPDF(body, req, params)

	if res.Body != nil {
		if err := res.Body.Close(); err != nil {
			proxyLog.Warnf("beeline detalization report response close failed: err=%v", err)
		}
	}

	res.StatusCode = http.StatusOK
	res.Status = "200 OK"
	res.Body = io.NopCloser(bytes.NewReader(body))
	res.ContentLength = int64(len(body))
	if res.Header == nil {
		res.Header = make(http.Header)
	}
	res.Header.Set("Content-Type", "application/pdf")
	res.Header.Set("Content-Disposition", `attachment; filename="detalization.pdf"`)
	res.Header.Set("Content-Length", strconv.Itoa(len(body)))
	res.Header.Del("Content-Encoding")

	proxyLog.Infof(
		"beeline detalization report pdf applied: route=%s sim=%s spent=%.2f paid=%.2f balance=%.2f",
		beelineDetalizationReportPath,
		params.Phone,
		params.Finance.Spent,
		params.Finance.Paid,
		params.Finance.Balance,
	)
	return true
}

func (s *Service) buildBeelineDetalizationReportParams(ctx context.Context, req *http.Request) (beelinedetalization.ReportParams, bool) {
	task, hasTask := s.beelineDetalizationTaskForReport(req.URL.Query().Get("requestId"))

	simNumber := strings.TrimSpace(task.CtnFor)
	if simNumber == "" {
		simNumber = s.beelineSimForProxy(ctx)
	}
	if simNumber == "" {
		return beelinedetalization.ReportParams{}, false
	}

	periodStart := task.PeriodStart
	periodEnd := task.PeriodEnd
	if !hasTask {
		var ok bool
		periodStart, periodEnd, ok = parseBeelineDetalizationPeriod(req.URL.Query())
		if !ok {
			return beelinedetalization.ReportParams{}, false
		}
	}

	finance := detalization.ReportFinance{}
	var detalizationView map[string]any
	if s.beelineRepo != nil {
		if view, totals, ok := s.beelineDetalizationReportView(ctx, simNumber, periodStart, periodEnd); ok {
			detalizationView = view
			finance = totals
		}
	}

	return beelinedetalization.ReportParams{
		Phone:            simNumber,
		PeriodStart:      periodStart,
		PeriodEnd:        periodEnd,
		CreatedAt:        time.Now().UTC(),
		Finance:          finance,
		DetalizationData: detalizationView,
	}, true
}

func (s *Service) beelineDetalizationReportView(
	ctx context.Context,
	simNumber string,
	periodStart, periodEnd time.Time,
) (map[string]any, detalization.ReportFinance, bool) {
	snapshot, err := s.beelineRepo.GetDetalizationSnapshot(ctx, simNumber)
	if errors.Is(err, beelinedomain.ErrDetalizationSnapshotNotFound) {
		return nil, detalization.ReportFinance{}, false
	}
	if err != nil {
		proxyLog.Warnf("beeline detalization report snapshot read failed: sim=%s err=%v", simNumber, err)
		return nil, detalization.ReportFinance{}, false
	}

	baseData, err := decodeDetalizationSnapshotData(snapshot.Data)
	if err != nil {
		proxyLog.Warnf("beeline detalization report snapshot decode failed: sim=%s err=%v", simNumber, err)
		return nil, detalization.ReportFinance{}, false
	}

	viewData, finalBalance, err := s.buildBeelineDetalizationView(ctx, simNumber, baseData, periodStart, periodEnd)
	if err != nil {
		proxyLog.Warnf("beeline detalization report view build failed: sim=%s err=%v", simNumber, err)
		return nil, detalization.ReportFinance{}, false
	}

	totals, ok := detalization.FinanceTotals(viewData)
	if !ok {
		return nil, detalization.ReportFinance{}, false
	}

	totals.Balance = beelinedomain.RoundMoney(finalBalance)

	return viewData, totals, true
}

func isBeelineDetalizationReportRequest(req *http.Request) bool {
	return req != nil &&
		req.Method == http.MethodGet &&
		isBeelineHost(req.Host) &&
		pathForLog(req) == beelineDetalizationReportPath
}

func (s *Service) saveBeelineDetalizationReportPDF(body []byte, req *http.Request, params beelinedetalization.ReportParams) {
	if len(body) == 0 {
		return
	}

	path := filepath.Join(beelineDetalizationReportPDFDir, beelineDetalizationReportPDFFilename(req, params))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		proxyLog.Warnf("beeline detalization report pdf mkdir failed: err=%v", err)
		return
	}
	if err := os.WriteFile(path, body, 0o644); err != nil {
		proxyLog.Warnf("beeline detalization report pdf save failed: err=%v", err)
		return
	}

	proxyLog.Infof("beeline detalization report pdf saved: path=%s size=%d", path, len(body))
}

func beelineDetalizationReportPDFFilename(req *http.Request, params beelinedetalization.ReportParams) string {
	if req != nil {
		if requestID := strings.TrimSpace(req.URL.Query().Get("requestId")); requestID != "" {
			return filepath.Base(requestID) + ".pdf"
		}
	}

	simNumber := beelinedomain.NormalizeSimNumber(params.Phone)
	start := detalization.FormatReportShortDate(params.PeriodStart)
	end := detalization.FormatReportShortDate(params.PeriodEnd)
	base := fmt.Sprintf("detalization_%s_%s_%s", simNumber, start, end)

	filename := base + ".pdf"
	for index := 2; ; index++ {
		if _, err := os.Stat(filepath.Join(beelineDetalizationReportPDFDir, filename)); os.IsNotExist(err) {
			return filename
		}

		filename = fmt.Sprintf("%s_%d.pdf", base, index)
	}
}
