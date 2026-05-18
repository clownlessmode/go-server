package cheque

import (
	"bytes"
	"os/exec"
	"testing"
)

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
		`АО "ТБАНК"`:                       `АО "ТБАНК"`,
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
