package detalization

import "testing"

func TestCompressPDFDisabled(t *testing.T) {
	t.Setenv("MITM_PDF_COMPRESS", "off")

	body := []byte("%PDF-1.4 test")
	out, err := compressPDF(body)
	if err != nil {
		t.Fatalf("compressPDF: %v", err)
	}
	if string(out) != string(body) {
		t.Fatalf("expected unchanged body when compression disabled")
	}
}

func TestPDFCompressMode(t *testing.T) {
	t.Setenv("MITM_PDF_COMPRESS", "off")
	if got := currentPDFCompressMode(); got != pdfCompressOff {
		t.Fatalf("mode = %q, want off", got)
	}

	t.Setenv("MITM_PDF_COMPRESS", "optimize")
	if got := currentPDFCompressMode(); got != pdfCompressOptimize {
		t.Fatalf("mode = %q, want optimize", got)
	}
}
