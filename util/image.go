package util

import (
	"image"
	"io"
)

func ImageGetConfigAdjusted(file io.Reader, width, height int) (image.Config, error) {
	img, _, err := image.DecodeConfig(file)
	if err == nil && (img.Width > width || img.Height > height) {
		if img.Width > img.Height {
			img.Height = int(float64(width) / (float64(img.Width) / float64(img.Height)))
			img.Width = width
		} else {
			img.Width = int(float64(height) / (float64(img.Height) / float64(img.Width)))
			img.Height = height
		}
	}
	return img, err
}
