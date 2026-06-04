package detalization

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

func compressPDF(body []byte) ([]byte, error) {
	if len(body) == 0 {
		return body, nil
	}

	mode := currentPDFCompressMode()
	if mode == pdfCompressOff {
		return body, nil
	}

	originalSize := len(body)

	switch mode {
	case pdfCompressScreen, pdfCompressEbook:
		if compressed, err := compressPDFWithGhostscript(body, mode, false); err == nil {
			logPDFCompressResult(originalSize, len(compressed), string(mode))
			return compressed, nil
		}
		fmt.Fprintf(os.Stderr, "pdf: ghostscript %s unavailable, falling back to pdfcpu optimize\n", mode)
	case pdfCompressAuto, pdfCompressStructured:
		if compressed, err := compressPDFWithGhostscript(body, mode, true); err == nil {
			logPDFCompressResult(originalSize, len(compressed), "structured")
			return compressed, nil
		}
		fmt.Fprintf(os.Stderr, "pdf: ghostscript structured compression unavailable, falling back to pdfcpu optimize\n")
	}

	compressed, err := compressPDFWithPdfcpu(body)
	if err != nil {
		return nil, err
	}

	logPDFCompressResult(originalSize, len(compressed), "optimize")
	return compressed, nil
}

func logPDFCompressResult(before, after int, method string) {
	if after >= before {
		fmt.Fprintf(os.Stderr, "pdf: compressed via %s (%d -> %d bytes, no reduction)\n", method, before, after)
		return
	}

	ratio := float64(after) / float64(before) * 100
	fmt.Fprintf(os.Stderr, "pdf: compressed via %s (%d -> %d bytes, %.1f%%)\n", method, before, after, ratio)
}

type pdfCompressMode string

const (
	pdfCompressOff        pdfCompressMode = "off"
	pdfCompressAuto       pdfCompressMode = "auto"
	pdfCompressOptimize   pdfCompressMode = "optimize"
	pdfCompressStructured pdfCompressMode = "structured"
	pdfCompressEbook      pdfCompressMode = "ebook"
	pdfCompressScreen     pdfCompressMode = "screen"
)

func currentPDFCompressMode() pdfCompressMode {
	raw := strings.ToLower(strings.TrimSpace(os.Getenv("MITM_PDF_COMPRESS")))
	switch raw {
	case "", "auto":
		if ghostscriptBinary() != "" {
			return pdfCompressStructured
		}
		return pdfCompressOptimize
	case "off", "false", "0", "none":
		return pdfCompressOff
	case "optimize", "pdfcpu":
		return pdfCompressOptimize
	case "structured", "ghostscript", "gs":
		return pdfCompressStructured
	case "ebook":
		return pdfCompressEbook
	case "screen":
		return pdfCompressScreen
	default:
		return pdfCompressOptimize
	}
}

func compressPDFWithPdfcpu(body []byte) ([]byte, error) {
	conf := model.NewDefaultConfiguration()
	conf.Optimize = true
	conf.OptimizeResourceDicts = true
	conf.OptimizeDuplicateContentStreams = true

	input := bytes.NewReader(body)
	output := &bytes.Buffer{}
	if err := api.Optimize(input, output, conf); err != nil {
		return nil, fmt.Errorf("optimize pdf: %w", err)
	}

	compressed := output.Bytes()
	if len(compressed) == 0 {
		return nil, fmt.Errorf("optimize pdf: result is empty")
	}

	return compressed, nil
}

func compressPDFWithGhostscript(body []byte, mode pdfCompressMode, preserveImages bool) ([]byte, error) {
	gsBin := ghostscriptBinary()
	if gsBin == "" {
		return nil, fmt.Errorf("ghostscript not found")
	}

	tempDir, err := makeDetalizationTempDir("beeline-detalization-compress-*")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tempDir)

	inputPath := filepath.Join(tempDir, "input.pdf")
	outputPath := filepath.Join(tempDir, "output.pdf")
	if err := os.WriteFile(inputPath, body, 0o644); err != nil {
		return nil, err
	}

	args := []string{
		gsBin,
		"-sDEVICE=pdfwrite",
		"-dCompatibilityLevel=1.4",
		"-dNOPAUSE",
		"-dQUIET",
		"-dBATCH",
		"-sOutputFile=" + outputPath,
	}
	if preserveImages {
		args = append(args,
			"-dDownsampleColorImages=false",
			"-dDownsampleGrayImages=false",
			"-dDownsampleMonoImages=false",
			"-dColorImageDownsampleThreshold=1.0",
			"-dGrayImageDownsampleThreshold=1.0",
			"-dMonoImageDownsampleThreshold=1.0",
		)
	} else {
		settings := "/ebook"
		if mode == pdfCompressScreen {
			settings = "/screen"
		}
		args = append(args, "-dPDFSETTINGS="+settings)
	}
	args = append(args, inputPath)

	cmd := exec.Command(args[0], args[1:]...)
	if output, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("ghostscript: %w: %s", err, strings.TrimSpace(string(output)))
	}

	return readPDF(outputPath)
}

func ghostscriptBinary() string {
	if custom := strings.TrimSpace(os.Getenv("MITM_GS_BIN")); custom != "" {
		if pathExists(custom) {
			return custom
		}
	}

	for _, candidate := range []string{
		"gs",
		"/usr/bin/gs",
		"/usr/local/bin/gs",
		"/opt/homebrew/bin/gs",
	} {
		if pathExists(candidate) {
			return candidate
		}
	}

	return ""
}

func pathExists(path string) bool {
	if path == "" {
		return false
	}

	if _, err := exec.LookPath(path); err == nil {
		return true
	}

	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
