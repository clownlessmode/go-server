package detalization

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteDataPageFontAssets(t *testing.T) {
	dir := t.TempDir()

	if err := writeDataPageFontAssets(dir); err != nil {
		t.Fatalf("write fonts: %v", err)
	}

	for _, name := range []string{beelineSansRegularWoff2, beelineSansRegularTTF, dataPageBannerFileName} {
		path := filepath.Join(dir, name)
		info, err := os.Stat(path)
		if err != nil {
			t.Fatalf("stat %s: %v", name, err)
		}
		if info.Size() == 0 {
			t.Fatalf("font %s is empty", name)
		}
	}
}
