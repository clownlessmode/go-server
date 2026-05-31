package detalization

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func makeDetalizationTempDir(prefix string) (string, error) {
	base, err := detalizationTempBaseAbs()
	if err != nil {
		return "", err
	}

	dir, err := os.MkdirTemp(base, prefix)
	if err != nil {
		return "", err
	}

	return filepath.Abs(dir)
}

func detalizationTempBaseAbs() (string, error) {
	if custom := strings.TrimSpace(os.Getenv("MITM_DETALIZATION_TMP")); custom != "" {
		return ensureDir(filepath.Clean(custom))
	}

	// Snap Chromium cannot write under $HOME (e.g. /root/.mitm); /tmp is allowed.
	for _, base := range []string{"/tmp/mitm-detalization", "/var/tmp/mitm-detalization"} {
		if dir, err := ensureDir(base); err == nil {
			return dir, nil
		}
	}

	if home, err := os.UserHomeDir(); err == nil && home != "" {
		if dir, err := ensureDir(filepath.Join(home, ".mitm", "detalization-tmp")); err == nil {
			return dir, nil
		}
	}

	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("resolve detalization temp base: %w", err)
	}

	return ensureDir(filepath.Join(wd, "data", "tmp"))
}

func ensureDir(path string) (string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(abs, 0o755); err != nil {
		return "", err
	}

	return abs, nil
}

func absPath(path string) string {
	abs, err := filepath.Abs(path)
	if err != nil {
		return path
	}

	return abs
}
