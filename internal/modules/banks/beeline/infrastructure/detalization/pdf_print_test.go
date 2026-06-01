package detalization

import (
	"os"
	"strings"
	"testing"
)

func TestPreparePDF2HtmlEXForPrint(t *testing.T) {
	templateBody, err := os.ReadFile("templates/first-page-v2.html")
	if err != nil {
		t.Fatalf("read template: %v", err)
	}

	html := string(preparePDF2HtmlEXForPrint(templateBody))
	if strings.Contains(html, "<script>") {
		t.Fatalf("expected pdf2htmlEX scripts to be removed")
	}
	if !strings.Contains(html, "pc{display:block!important") {
		t.Fatalf("expected print visibility css")
	}
	if !strings.Contains(html, "page-container") {
		t.Fatalf("expected page markup to remain")
	}
}
