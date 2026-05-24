package detalization

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
)

//go:embed templates/beelinesans-regular.woff2 templates/beelinesans-regular.ttf templates/last-banner-3x.png
var dataPageAssetFS embed.FS

const (
	beelineSansRegularWoff2 = "beelinesans-regular.woff2"
	beelineSansRegularTTF   = "beelinesans-regular.ttf"
	dataPageBannerFileName  = "last-banner-3x.png"
)

func writeDataPageFontAssets(dir string) error {
	return writeDataPageAssets(dir)
}

func writeDataPageAssets(dir string) error {
	for _, name := range []string{
		"templates/" + beelineSansRegularWoff2,
		"templates/" + beelineSansRegularTTF,
		"templates/" + dataPageBannerFileName,
	} {
		body, err := dataPageAssetFS.ReadFile(name)
		if err != nil {
			return fmt.Errorf("read asset %s: %w", name, err)
		}

		target := filepath.Join(dir, filepath.Base(name))
		if err := os.WriteFile(target, body, 0o644); err != nil {
			return fmt.Errorf("write asset %s: %w", target, err)
		}
	}

	return nil
}
