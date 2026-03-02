package media

import (
	"bytes"
	"context"
	"fmt"
	"image"

	_ "golang.org/x/image/webp"
)

type BasicAvatarModerator struct{}

func NewBasicAvatarModerator() *BasicAvatarModerator {
	return &BasicAvatarModerator{}
}

func (m *BasicAvatarModerator) Moderate(_ context.Context, content []byte, _ string) error {
	cfg, _, err := image.DecodeConfig(bytes.NewReader(content))
	if err != nil {
		return fmt.Errorf("decode avatar config: %w", err)
	}
	if cfg.Width < 64 || cfg.Height < 64 {
		return fmt.Errorf("avatar is too small; minimum size is 64x64")
	}
	aspect := float64(cfg.Width) / float64(cfg.Height)
	if aspect > 3 || aspect < 1.0/3.0 {
		return fmt.Errorf("avatar aspect ratio is not allowed")
	}
	return nil
}
