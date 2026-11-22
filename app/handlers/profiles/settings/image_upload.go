package settings

import (
	"errors"
	"image"
	"io"
)

const maxImageUploadBytes = 10 << 20

func validateImageDimensions(file io.ReadSeeker) error {
	const maxSide = 4096
	const maxPixels = 12_000_000
	cfg, _, err := image.DecodeConfig(file)
	if err != nil {
		return err
	}
	if cfg.Width <= 0 || cfg.Height <= 0 || cfg.Width > maxSide || cfg.Height > maxSide ||
		cfg.Width*cfg.Height > maxPixels {
		return errors.New("image exceeds upload dimensions")
	}
	_, err = file.Seek(0, io.SeekStart)
	return err
}
