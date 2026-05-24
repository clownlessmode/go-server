package detalization

import (
	"os"
	"strings"
	"testing"
)

func TestInjectPageLogo(t *testing.T) {
	for _, templateName := range []string{"templates/first-page.html", "templates/second-page.html"} {
		body, err := os.ReadFile(templateName)
		if err != nil {
			t.Fatalf("read %s: %v", templateName, err)
		}

		html := string(injectPageLogo(body))
		checks := []string{
			"page-logo",
			`<svg width="40" height="40"`,
			"top:48px",
			"right:84px",
			"width:46px",
			"top:32pt",
			"right:56pt",
			`</div><div class="page-logo">`,
			`</div><div class="pi"`,
		}
		for _, check := range checks {
			if !strings.Contains(html, check) {
				t.Fatalf("%s missing %q", templateName, check)
			}
		}

		if strings.Count(html, `<div class="page-logo">`) != 1 {
			t.Fatalf("%s expected one logo overlay", templateName)
		}

		logoIdx := strings.Index(html, `<div class="page-logo">`)
		pcIdx := strings.Index(html, `<div class="pc pc`)
		piIdx := strings.Index(html, `<div class="pi"`)
		if logoIdx < pcIdx || logoIdx > piIdx {
			t.Fatalf("%s logo must sit between .pc and .pi", templateName)
		}
	}
}

func TestInjectDataPageLogo(t *testing.T) {
	body, err := os.ReadFile("templates/data-page.html")
	if err != nil {
		t.Fatalf("read data-page: %v", err)
	}

	html := injectDataPageLogo(string(body))
	if !strings.Contains(html, `<svg width="40" height="40"`) {
		t.Fatalf("data page logo svg missing")
	}
	if strings.Contains(html, "background: #313946") {
		t.Fatalf("black square placeholder still present")
	}
}
