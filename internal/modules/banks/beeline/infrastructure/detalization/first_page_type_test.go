package detalization

import (
	"testing"
)

func TestFirstPageTemplate(t *testing.T) {
	t.Cleanup(func() {
		SetFirstPageType("")
	})

	t.Setenv(envFirstPageType, "")
	if got := firstPageTemplate(); got != firstPageTemplateV1 {
		t.Fatalf("default template = %q, want %q", got, firstPageTemplateV1)
	}

	t.Setenv(envFirstPageType, "1")
	if got := firstPageTemplate(); got != firstPageTemplateV1 {
		t.Fatalf("template 1 = %q, want %q", got, firstPageTemplateV1)
	}

	t.Setenv(envFirstPageType, "2")
	if got := firstPageTemplate(); got != firstPageTemplateV2 {
		t.Fatalf("template 2 = %q, want %q", got, firstPageTemplateV2)
	}

	SetFirstPageType("2")
	if got := firstPageTemplate(); got != firstPageTemplateV2 {
		t.Fatalf("override template 2 = %q, want %q", got, firstPageTemplateV2)
	}

	SetFirstPageType("1")
	if got := firstPageTemplate(); got != firstPageTemplateV1 {
		t.Fatalf("override template 1 = %q, want %q", got, firstPageTemplateV1)
	}
}

func TestRenderFirstPageHTMLV2(t *testing.T) {
	templateBody, err := templateFS.ReadFile(firstPageTemplateV2)
	if err != nil {
		t.Fatalf("read template: %v", err)
	}

	runRenderFirstPageHTMLV2Test(t, templateBody)
}
