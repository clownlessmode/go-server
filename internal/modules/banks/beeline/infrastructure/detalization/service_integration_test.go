//go:build integration

package detalization

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/pdfcpu/pdfcpu/pkg/api"

	detaildomain "project/internal/modules/banks/beeline/detalization"
)

func TestGenerateReportPDFIntegration(t *testing.T) {
	svc := NewService()
	start, _ := time.Parse("2006-01-02", "2026-02-17")
	end, _ := time.Parse("2006-01-02", "2026-03-16")
	created, _ := time.Parse("2006-01-02 15:04", "2026-03-16 05:01")

	pdf, err := svc.GenerateReportPDF(ReportParams{
		Phone:       "79629844593",
		PeriodStart: start,
		PeriodEnd:   end,
		CreatedAt:   created,
		Finance: detaildomain.ReportFinance{
			OpeningBalance:  0,
			Spent:             -99589.25,
			Paid:              99591,
			Balance:           1.75,
			PaymentsTransfers: -98840.94,
			OtherSpent:        -748.31,
		},
	})
	if err != nil {
		t.Fatalf("GenerateReportPDF: %v", err)
	}
	if len(pdf) < 1000 {
		t.Fatalf("pdf too small: %d bytes", len(pdf))
	}
	if pdf[0] != '%' || pdf[1] != 'P' || pdf[2] != 'D' || pdf[3] != 'F' {
		t.Fatalf("invalid pdf header: %q", pdf[:min(8, len(pdf))])
	}

	tempDir := t.TempDir()
	mergedPath := filepath.Join(tempDir, "merged.pdf")
	if err := os.WriteFile(mergedPath, pdf, 0o644); err != nil {
		t.Fatal(err)
	}
	pageCount, err := api.PageCountFile(mergedPath)
	if err != nil {
		t.Fatalf("page count: %v", err)
	}
	if pageCount != 2 {
		t.Fatalf("expected 2 pages, got %d", pageCount)
	}

	if path := os.Getenv("BEELINE_DETALIZATION_PDF_OUT"); path != "" {
		if err := os.WriteFile(path, pdf, 0o644); err != nil {
			t.Fatalf("write pdf: %v", err)
		}
	}
}
