package detalization

import (
	"embed"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

//go:embed templates/first-page.html templates/first-page-v2.html templates/second-page.html templates/data-page.html
var templateFS embed.FS

const (
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
		firstPageTemplate(),
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

	tempDir, err := makeDetalizationTempDir("beeline-detalization-*")
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

	htmlPath := absPath(filepath.Join(tempDir, baseName+".html"))
	pdfPath := absPath(filepath.Join(tempDir, baseName+".pdf"))
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
	htmlPath = absPath(htmlPath)
	pdfPath = absPath(pdfPath)
	pdfName := filepath.Base(pdfPath)
	htmlURL := htmlFileURL(htmlPath)
	workDir := filepath.Dir(pdfPath)

	var errors []string
	for _, browser := range htmlToPDFBrowsers() {
		if _, err := exec.LookPath(browser); err != nil {
			if _, statErr := os.Stat(browser); statErr != nil {
				errors = append(errors, fmt.Sprintf("%s not found", browser))
				continue
			}
		}

		userDataDir, err := os.MkdirTemp(workDir, "chrome-profile-*")
		if err != nil {
			return fmt.Errorf("create chrome profile dir: %w", err)
		}
		userDataDir = absPath(userDataDir)
		defer os.RemoveAll(userDataDir)

		args := []string{
			"--headless=new",
			"--disable-gpu",
			"--no-sandbox",
			"--disable-dev-shm-usage",
			"--disable-software-rasterizer",
			"--no-first-run",
			"--no-default-browser-check",
			"--user-data-dir=" + userDataDir,
			"--no-pdf-header-footer",
			"--run-all-compositor-stages-before-draw",
			"--virtual-time-budget=5000",
			"--print-to-pdf=" + pdfName,
			htmlURL,
		}

		cmd := exec.Command(browser, args...)
		cmd.Dir = workDir
		cmd.Env = os.Environ()
		output, err := cmd.CombinedOutput()
		if err != nil {
			errors = append(errors, fmt.Sprintf("%s failed: %v: %s", browser, err, strings.TrimSpace(string(output))))
			continue
		}

		if statErr := waitForPDF(workDir, pdfName, pdfPath); statErr != nil {
			errors = append(errors, fmt.Sprintf(
				"%s exited ok but pdf missing/empty at %s: %v; output: %s",
				browser,
				pdfPath,
				statErr,
				strings.TrimSpace(string(output)),
			))
			continue
		}

		return nil
	}

	return fmt.Errorf(
		"convert html detalization to pdf: %s; install google-chrome-stable (.deb, not snap): wget https://dl.google.com/linux/direct/google-chrome-stable_current_amd64.deb && apt install ./google-chrome-stable_current_amd64.deb",
		strings.Join(errors, "; "),
	)
}

func waitForPDF(workDir, pdfName, pdfPath string) error {
	candidates := []string{filepath.Join(workDir, pdfName), pdfPath}
	for range 20 {
		for _, candidate := range candidates {
			info, err := os.Stat(candidate)
			if err != nil || info.Size() == 0 {
				continue
			}
			if candidate != pdfPath {
				if renameErr := os.Rename(candidate, pdfPath); renameErr != nil {
					return renameErr
				}
			}
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}

	return fmt.Errorf("stat %s: no such file or directory", pdfPath)
}

func htmlFileURL(path string) string {
	abs, err := filepath.Abs(path)
	if err != nil {
		abs = path
	}

	return (&url.URL{Scheme: "file", Path: filepath.ToSlash(abs)}).String()
}

func htmlToPDFBrowsers() []string {
	browsers := []string{
		"/usr/bin/google-chrome-stable",
		"google-chrome-stable",
		"/usr/bin/google-chrome",
		"google-chrome",
		"/snap/bin/chromium",
		"/usr/lib/chromium/chromium",
		"chromium",
	}

	if custom := strings.TrimSpace(os.Getenv("MITM_CHROME_BIN")); custom != "" {
		browsers = append([]string{custom}, browsers...)
	}

	if runtime.GOOS == "darwin" {
		browsers = append(browsers,
			"/Applications/Chromium.app/Contents/MacOS/Chromium",
			"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
		)
	}

	return browsers
}