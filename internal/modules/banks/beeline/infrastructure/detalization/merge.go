package detalization

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

func mergePDFs(parts ...[]byte) ([]byte, error) {
	if len(parts) == 0 {
		return nil, fmt.Errorf("merge pdfs: no input")
	}
	if len(parts) == 1 {
		return parts[0], nil
	}

	tempDir, err := os.MkdirTemp("", "beeline-detalization-merge-*")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tempDir)

	inputs := make([]string, 0, len(parts))
	for index, part := range parts {
		if len(part) == 0 {
			return nil, fmt.Errorf("merge pdfs: page %d is empty", index+1)
		}

		path := filepath.Join(tempDir, fmt.Sprintf("page-%d.pdf", index+1))
		if err := os.WriteFile(path, part, 0o644); err != nil {
			return nil, err
		}
		inputs = append(inputs, path)
	}

	outputPath := filepath.Join(tempDir, "merged.pdf")
	if err := api.MergeCreateFile(inputs, outputPath, false, model.NewDefaultConfiguration()); err != nil {
		return nil, fmt.Errorf("merge pdfs: %w", err)
	}

	merged, err := os.ReadFile(outputPath)
	if err != nil {
		return nil, err
	}
	if len(merged) == 0 {
		return nil, fmt.Errorf("merge pdfs: result is empty")
	}

	return merged, nil
}

func readPDF(path string) ([]byte, error) {
	body, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if len(body) == 0 {
		return nil, fmt.Errorf("pdf is empty: %s", path)
	}

	return body, nil
}

func keepFirstPDFPage(body []byte) ([]byte, error) {
	tempDir, err := os.MkdirTemp("", "beeline-detalization-page-*")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tempDir)

	inputPath := filepath.Join(tempDir, "input.pdf")
	outputPath := filepath.Join(tempDir, "page-1.pdf")
	if err := os.WriteFile(inputPath, body, 0o644); err != nil {
		return nil, err
	}

	if err := api.TrimFile(inputPath, outputPath, []string{"1"}, model.NewDefaultConfiguration()); err != nil {
		return nil, fmt.Errorf("trim pdf to first page: %w", err)
	}

	return readPDF(outputPath)
}
