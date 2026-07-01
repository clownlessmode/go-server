package cheque

import (
	"bytes"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pdfcpu/pdfcpu/pkg/api"

	"project/internal/modules/banks/rocketbank/domain"
)

func TestSBPChequeTemplateReplacesBankName(t *testing.T) {
	templateBody, err := templateFS.ReadFile(outgoingTemplate().Path)
	if err != nil {
		t.Fatal(err)
	}

	tpl := outgoingTemplate()
	svg := string(templateBody)
	if !strings.Contains(svg, tpl.Bank) {
		t.Fatalf("template missing bank placeholder %q", tpl.Bank)
	}

	svg = strings.ReplaceAll(svg, tpl.Bank, escapeSVGText(`ПАО СБЕРБАНК`))
	if strings.Contains(svg, tpl.Bank) {
		t.Fatal("bank placeholder was not replaced")
	}
	if !strings.Contains(svg, "ПАО СБЕРБАНК") {
		t.Fatal("replacement bank name missing from svg")
	}
}

func TestGenerateSBPTransferChequeUsesConfiguredBank(t *testing.T) {
	if !hasChequePDFConverter() {
		t.Skip("sips or rsvg-convert is required to generate SBP cheque PDF")
	}

	service := NewService()
	item := domain.NewSBPTransferHistoryItem(domain.SBPTransferInput{
		Amount:              500,
		BalanceBefore:       1000,
		Direction:           "OUTGOING",
		Time:                "2026-06-19T12:00:00+0700",
		OperationFirstName:  "ИВАН",
		OperationMiddleName: "ИВАНОВИЧ",
		OperationLastName:   "ИВАНОВ",
		BankID:              "sberbank",
		PhoneNumber:         "+79001234567",
	})

	if err := service.GenerateSBPTransferCheque(item, domain.ClientInfo{}); err != nil {
		t.Fatal(err)
	}

	path := SBPTransferChequePath(item)
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.HasPrefix(body, []byte("%PDF-")) {
		t.Fatal("generated body is not a PDF")
	}
	_ = os.Remove(path)
}

func TestRenderSVGTemplatePDF(t *testing.T) {
	if _, err := exec.LookPath("sips"); err != nil {
		t.Skip("sips is required to convert SVG cheque templates to PDF")
	}

	templateBody, err := templateFS.ReadFile(outgoingTemplate().Path)
	if err != nil {
		t.Fatal(err)
	}

	body, err := renderSVGTemplatePDF(templateBody, map[string]string{
		"16.05.2026 18:14 ПО МСК":          "16.05.2026 18:14 ПО МСК",
		"50 ₽":                             "500 ₽",
		"АЗАТ АЛИКОВИЧ Г":                  "БУЗАТ АЛИЗАДЕ П",
		"+7 909 933-40-05":                 "+7 900 123-45-67",
		`АО &quot;ТБАНК&quot;`:             `ПАО СБЕРБАНК`,
		"МАКСИМ АЛЕКСАНДРОВИЧ Н.":          "ИВАН ИВАНОВИЧ И.",
		"+7 983 543-99-99":                 "+7 900 123-45-67",
		"40817 81035 02245 32469":          "40817 81000 00000 00000",
		"B61361514043330I0B10100011760501": "B61399534683834615I0B10100011760501",
		"M70093717871":                     "M00761679072",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.HasPrefix(body, []byte("%PDF-")) {
		t.Fatal("generated body is not a PDF")
	}
}

func TestRenderSVGTemplatePDFUsesOriginalPageDimensions(t *testing.T) {
	if !hasChequePDFConverter() {
		t.Skip("sips or rsvg-convert is required to convert SVG cheque templates to PDF")
	}

	tests := []struct {
		name       string
		tpl        chequeTemplate
		wantWidth  float64
		wantHeight float64
	}{
		{
			name:       "incoming sbp",
			tpl:        incomingTemplate(),
			wantWidth:  420,
			wantHeight: 1050,
		},
		{
			name:       "outgoing sbp",
			tpl:        outgoingTemplate(),
			wantWidth:  420,
			wantHeight: 1116,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			templateBody, err := templateFS.ReadFile(tt.tpl.Path)
			if err != nil {
				t.Fatal(err)
			}

			body, err := renderSVGTemplatePDF(templateBody, nil)
			if err != nil {
				t.Fatal(err)
			}

			pdfPath := filepath.Join(t.TempDir(), "cheque.pdf")
			if err := os.WriteFile(pdfPath, body, 0o644); err != nil {
				t.Fatal(err)
			}

			dims, err := api.PageDimsFile(pdfPath)
			if err != nil {
				t.Fatal(err)
			}
			if len(dims) != 1 {
				t.Fatalf("expected one page dimension, got %d", len(dims))
			}
			if math.Abs(dims[0].Width-tt.wantWidth) > 0.01 || math.Abs(dims[0].Height-tt.wantHeight) > 0.01 {
				t.Fatalf("page size = %.2f x %.2f, want %.2f x %.2f", dims[0].Width, dims[0].Height, tt.wantWidth, tt.wantHeight)
			}
		})
	}
}

func hasChequePDFConverter() bool {
	if _, err := exec.LookPath("sips"); err == nil {
		return true
	}
	if _, err := exec.LookPath("rsvg-convert"); err == nil {
		return true
	}
	return false
}
