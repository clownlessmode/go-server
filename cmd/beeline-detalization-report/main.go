package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"project/internal/app/config"
	"project/internal/app/database"
	"project/internal/app/logger"
	"project/internal/modules/banks/beeline/detalization"
	beelinedomain "project/internal/modules/banks/beeline/domain"
	beelinedetalization "project/internal/modules/banks/beeline/infrastructure/detalization"
	beelinepostgres "project/internal/modules/banks/beeline/infrastructure/postgres"
)

const (
	pdfOutputDir  = "data/reports/beeline/cheques"
	htmlOutputDir = "data/reports/beeline/htmls_pages"
)

var reportLog = logger.New("beeline-detalization-report")

func main() {
	sim := flag.String("sim", "", "beeline sim number")
	day := flag.Bool("day", false, "limit report to last day")
	week := flag.Bool("week", false, "limit report to last 7 days")
	month := flag.Bool("month", false, "use full billing month period (default)")
	flag.Parse()

	if *sim == "" {
		reportLog.Fatalf("-sim is required")
	}

	periodMode := "month"
	switch {
	case *day:
		periodMode = "day"
	case *week:
		periodMode = "week"
	case *month:
		periodMode = "month"
	}

	selected := 0
	if *day {
		selected++
	}
	if *week {
		selected++
	}
	if *month {
		selected++
	}
	if selected > 1 {
		reportLog.Fatalf("only one of -day, -week, -month can be specified")
	}

	if err := resetOutputDir(pdfOutputDir); err != nil {
		reportLog.Fatalf("clean pdf dir: %v", err)
	}
	if err := resetOutputDir(htmlOutputDir); err != nil {
		reportLog.Fatalf("clean html dir: %v", err)
	}

	cfg := config.Load()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	db, err := database.NewPostgres(ctx, cfg.Postgres)
	if err != nil {
		reportLog.Fatalf("connect postgres: %v", err)
	}
	defer db.Close()

	repo := beelinepostgres.NewRepository(db)
	simNumber := beelinedomain.NormalizeSimNumber(*sim)

	params, err := buildReportParams(ctx, repo, simNumber, periodMode)
	if err != nil {
		reportLog.Fatalf("build report params: %v", err)
	}

	svc := beelinedetalization.NewService()
	pdf, err := svc.GenerateReportPDFWithHTML(params, htmlOutputDir)
	if err != nil {
		reportLog.Fatalf("generate pdf: %v", err)
	}

	pdfPath := filepath.Join(pdfOutputDir, reportFilename(simNumber, params.PeriodStart, params.PeriodEnd))
	if err := os.WriteFile(pdfPath, pdf, 0o644); err != nil {
		reportLog.Fatalf("write pdf: %v", err)
	}

	reportLog.Successf(
		"report generated: sim=%s pdf=%s html=%s spent=%.2f paid=%.2f balance=%.2f",
		simNumber,
		pdfPath,
		htmlOutputDir,
		params.Finance.Spent,
		params.Finance.Paid,
		params.Finance.Balance,
	)
}

func resetOutputDir(dir string) error {
	if err := os.RemoveAll(dir); err != nil {
		return err
	}

	return os.MkdirAll(dir, 0o755)
}

func buildReportParams(
	ctx context.Context,
	repo beelinedomain.Repository,
	simNumber string,
	periodMode string,
) (beelinedetalization.ReportParams, error) {
	snapshot, err := repo.GetDetalizationSnapshot(ctx, simNumber)
	if err != nil {
		return beelinedetalization.ReportParams{}, fmt.Errorf("get snapshot: %w", err)
	}

	periodStart, periodEnd := detalization.ReportPeriodForMode(periodMode, snapshot.PeriodStart, snapshot.PeriodEnd)

	baseData, err := detalization.DecodeSnapshotData(snapshot.Data)
	if err != nil {
		return beelinedetalization.ReportParams{}, fmt.Errorf("decode snapshot: %w", err)
	}

	payments, err := repo.ListPaymentsInPeriod(ctx, simNumber, periodStart, periodEnd)
	if err != nil {
		return beelinedetalization.ReportParams{}, fmt.Errorf("list payments: %w", err)
	}

	hiddenIDs, err := repo.ListHiddenTransactionIDs(ctx, simNumber)
	if err != nil {
		return beelinedetalization.ReportParams{}, fmt.Errorf("list hidden transactions: %w", err)
	}

	viewData, finalBalance, err := detalization.BuildView(baseData, payments, hiddenIDs)
	if err != nil {
		return beelinedetalization.ReportParams{}, fmt.Errorf("build view: %w", err)
	}

	if periodMode != "month" {
		viewData, finalBalance, err = detalization.TrimViewToPeriod(viewData, periodStart, periodEnd)
		if err != nil {
			return beelinedetalization.ReportParams{}, fmt.Errorf("trim period: %w", err)
		}
	}

	finance, ok := detalization.FinanceTotals(viewData)
	if !ok {
		return beelinedetalization.ReportParams{}, fmt.Errorf("finance totals unavailable")
	}
	finance.Balance = beelinedomain.RoundMoney(finalBalance)

	return beelinedetalization.ReportParams{
		Phone:            simNumber,
		PeriodStart:      periodStart,
		PeriodEnd:        periodEnd,
		CreatedAt:        time.Now().UTC(),
		Finance:          finance,
		DetalizationData: viewData,
	}, nil
}

func reportFilename(simNumber string, periodStart, periodEnd time.Time) string {
	start := detalization.FormatReportShortDate(periodStart)
	end := detalization.FormatReportShortDate(periodEnd)

	return fmt.Sprintf("detalization_%s_%s_%s.pdf", simNumber, start, end)
}
