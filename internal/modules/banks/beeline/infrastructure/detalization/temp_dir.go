package detalization

import (
	"os"
	"path/filepath"
)

const detalizationTempBase = "data/tmp"

func makeDetalizationTempDir(prefix string) (string, error) {
	if err := os.MkdirAll(detalizationTempBase, 0o755); err != nil {
		return os.MkdirTemp("", prefix)
	}

	return os.MkdirTemp(detalizationTempBase, prefix)
}

func chromeUserDataDir(tempDir string) string {
	return filepath.Join(tempDir, "chrome-profile")
}
