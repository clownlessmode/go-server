package detalization

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
)

const dataPageBannerJPEGName = "last-banner-3x.jpg"

func writeDataPageBannerJPEG(dir string) error {
	pngData, err := dataPageAssetFS.ReadFile("templates/" + dataPageBannerFileName)
	if err != nil {
		return err
	}

	src, err := png.Decode(bytes.NewReader(pngData))
	if err != nil {
		return err
	}

	rgb := flattenImageAlpha(src)
	var encoded bytes.Buffer
	if err := jpeg.Encode(&encoded, rgb, &jpeg.Options{Quality: 92}); err != nil {
		return err
	}

	target := filepath.Join(dir, dataPageBannerJPEGName)
	if err := os.WriteFile(target, encoded.Bytes(), 0o644); err != nil {
		return err
	}

	return nil
}

func flattenImageAlpha(src image.Image) image.Image {
	bounds := src.Bounds()
	dst := image.NewRGBA(bounds)
	draw.Draw(dst, bounds, &image.Uniform{color.White}, image.Point{}, draw.Src)
	draw.Draw(dst, bounds, src, bounds.Min, draw.Over)
	return dst
}
