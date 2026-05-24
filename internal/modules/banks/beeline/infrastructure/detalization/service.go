package detalization

import (
	"embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

//go:embed templates/first-page.html templates/second-page.html templates/data-page.html
var templateFS embed.FS

const (
	firstPageTemplate  = "templates/first-page.html"
	secondPageTemplate = "templates/second-page.html"
	dataPageTemplate   = "templates/data-page.html"
)

type Service struct{}

func NewService() *Service {
	return &Service{}
}

func (s *Service) GenerateReportPDF(params ReportParams) ([]byte, error) {
	return s.generateReport(params, "")
}

func (s *Service) GenerateReportPDFWithHTML(params ReportParams, htmlPagesDir string) ([]byte, error) {
	return s.generateReport(params, htmlPagesDir)
}

func (s *Service) generateReport(params ReportParams, htmlPagesDir string) ([]byte, error) {
	pages := ensureTransactionPages(buildTransactionPagePlan(params.DetalizationData))
	pdfParts := make([][]byte, 0, 1+len(pages))

	firstPagePDF, err := s.generateFirstPagePDF(params, htmlPagesDir)
	if err != nil {
		return nil, fmt.Errorf("first page: %w", err)
	}
	pdfParts = append(pdfParts, firstPagePDF)

	for _, page := range pages {
		switch page.Kind {
		case transactionPageSecond:
			secondPagePDF, err := s.generatePagePDF(
				secondPageTemplate,
				func(templateBody []byte, params ReportParams) []byte {
					return renderSecondPageHTML(templateBody, params, page)
				},
				params,
				htmlPagesDir,
				fmt.Sprintf("second-page-%d", page.PageNumber),
			)
			if err != nil {
				return nil, fmt.Errorf("second page %d: %w", page.PageNumber, err)
			}
			pdfParts = append(pdfParts, secondPagePDF)
		case transactionPageData:
			dataPagePDF, err := s.generatePagePDF(
				dataPageTemplate,
				func(templateBody []byte, params ReportParams) []byte {
					return renderDataPageHTML(templateBody, params, page)
				},
				params,
				htmlPagesDir,
				fmt.Sprintf("data-page-%d", page.PageNumber),
			)
			if err != nil {
				return nil, fmt.Errorf("data page %d: %w", page.PageNumber, err)
			}
			pdfParts = append(pdfParts, dataPagePDF)
		}
	}

	return mergePDFs(pdfParts...)
}

func (s *Service) GenerateFirstPagePDF(params ReportParams) ([]byte, error) {
	return s.GenerateReportPDF(params)
}

func (s *Service) generateFirstPagePDF(params ReportParams, htmlPagesDir string) ([]byte, error) {
	return s.generatePagePDF(
		firstPageTemplate,
		func(templateBody []byte, params ReportParams) []byte {
			return renderFirstPageHTML(templateBody, params)
		},
		params,
		htmlPagesDir,
		"first-page",
	)
}

type pageRenderer func(templateBody []byte, params ReportParams) []byte

func (s *Service) generatePagePDF(
	templateName string,
	render pageRenderer,
	params ReportParams,
	htmlPagesDir string,
	htmlFileName string,
) ([]byte, error) {
	templateBody, err := templateFS.ReadFile(templateName)
	if err != nil {
		return nil, fmt.Errorf("read detalization template %s: %w", templateName, err)
	}

	htmlBody := render(templateBody, params)
	if templateName != dataPageTemplate {
		htmlBody = injectPageLogo(htmlBody)
		htmlBody = injectPrintPageBreakFix(htmlBody)
	}

	if htmlPagesDir != "" {
		if err := os.MkdirAll(htmlPagesDir, 0o755); err != nil {
			return nil, fmt.Errorf("create html pages dir: %w", err)
		}
		if htmlFileName == "" {
			htmlFileName = strings.TrimSuffix(filepath.Base(templateName), ".html")
		}
		if templateName == dataPageTemplate {
			if err := writeDataPageAssets(htmlPagesDir); err != nil {
				return nil, fmt.Errorf("write data page assets: %w", err)
			}
		}
		if err := os.WriteFile(filepath.Join(htmlPagesDir, htmlFileName+".html"), htmlBody, 0o644); err != nil {
			return nil, fmt.Errorf("write html page %s: %w", htmlFileName, err)
		}
	}

	tempDir, err := os.MkdirTemp("", "beeline-detalization-*")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tempDir)

	baseName := htmlFileName
	if baseName == "" {
		baseName = strings.TrimSuffix(filepath.Base(templateName), ".html")
	}

	if templateName == dataPageTemplate {
		if err := writeDataPageAssets(tempDir); err != nil {
			return nil, fmt.Errorf("write data page assets: %w", err)
		}
	}

	htmlPath := filepath.Join(tempDir, baseName+".html")
	pdfPath := filepath.Join(tempDir, baseName+".pdf")
	if err := os.WriteFile(htmlPath, htmlBody, 0o644); err != nil {
		return nil, err
	}

	if err := convertHTMLToPDF(htmlPath, pdfPath); err != nil {
		return nil, err
	}

	body, err := readPDF(pdfPath)
	if err != nil {
		return nil, err
	}

	return keepFirstPDFPage(body)
}

// pdf2htmlEX templates lay out one page at 793.333×1122.667 pt (4/3 of A4).
// Chrome defaults to US Letter (792 pt tall), splits overflow onto page 2, and
// keepFirstPDFPage then drops the footer. Scale content to A4 (595×842 pt).
const printPageBreakFix = `<style>@media print{
@page{size:595pt 842pt;margin:0}
html,body{margin:0;padding:0}
#sidebar{display:none}
.pf{width:595pt!important;height:842pt!important;overflow:hidden!important;page-break-after:auto!important;page-break-inside:avoid!important}
.pc{transform:scale(0.75);transform-origin:0 0;-webkit-transform:scale(0.75);-webkit-transform-origin:0 0}
}</style>`

func injectPrintPageBreakFix(htmlBody []byte) []byte {
	html := string(htmlBody)
	if strings.Contains(html, "</head>") {
		return []byte(strings.Replace(html, "</head>", printPageBreakFix+"</head>", 1))
	}

	return append([]byte(printPageBreakFix), htmlBody...)
}

func convertHTMLToPDF(htmlPath string, pdfPath string) error {
	htmlURL := "file://" + htmlPath
	args := []string{
		"--headless=new",
		"--disable-gpu",
		"--no-sandbox",
		"--disable-dev-shm-usage",
		"--no-pdf-header-footer",
		"--run-all-compositor-stages-before-draw",
		"--virtual-time-budget=5000",
		"--print-to-pdf=" + pdfPath,
		htmlURL,
	}

	var errors []string
	for _, browser := range htmlToPDFBrowsers() {
		if _, err := exec.LookPath(browser); err != nil {
			if _, statErr := os.Stat(browser); statErr != nil {
				errors = append(errors, fmt.Sprintf("%s not found", browser))
				continue
			}
		}

		cmd := exec.Command(browser, args...)
		if output, err := cmd.CombinedOutput(); err != nil {
			errors = append(errors, fmt.Sprintf("%s failed: %v: %s", browser, err, strings.TrimSpace(string(output))))
			continue
		}

		return nil
	}

	return fmt.Errorf("convert html detalization to pdf: %s; install Chromium or Google Chrome", strings.Join(errors, "; "))
}

func htmlToPDFBrowsers() []string {
	browsers := []string{
		"chromium",
		"google-chrome",
		"google-chrome-stable",
	}

	if runtime.GOOS == "darwin" {
		browsers = append(browsers,
			"/Applications/Chromium.app/Contents/MacOS/Chromium",
			"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
		)
	}

	return browsers
}