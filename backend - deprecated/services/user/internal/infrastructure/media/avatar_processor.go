package media

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"net/http"
	"strings"

	"golang.org/x/image/draw"
	_ "golang.org/x/image/webp"
	"jobconnect/user/internal/domain"
)

const (
	maxAvatarDimension = 512
)

type AvatarProcessor struct{}

func NewAvatarProcessor() *AvatarProcessor {
	return &AvatarProcessor{}
}

func (p *AvatarProcessor) Process(content []byte, declaredContentType string) ([]byte, string, int, int, error) {
	if err := domain.ValidateAvatarSize(len(content)); err != nil {
		return nil, "", 0, 0, err
	}
	actualContentType := http.DetectContentType(content)
	if strings.TrimSpace(declaredContentType) != "" {
		declared := strings.ToLower(strings.TrimSpace(declaredContentType))
		if declared != "image/jpeg" && declared != "image/png" && declared != "image/webp" {
			return nil, "", 0, 0, fmt.Errorf("unsupported avatar content_type")
		}
	}
	if err := domain.ValidateAvatarContentType(actualContentType); err != nil {
		return nil, "", 0, 0, err
	}

	img, format, err := image.Decode(bytes.NewReader(content))
	if err != nil {
		return nil, "", 0, 0, fmt.Errorf("decode avatar: %w", err)
	}
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	if width <= 0 || height <= 0 {
		return nil, "", 0, 0, fmt.Errorf("invalid avatar dimensions")
	}

	resized := img
	if width > maxAvatarDimension || height > maxAvatarDimension {
		nw, nh := fitSize(width, height, maxAvatarDimension)
		dst := image.NewRGBA(image.Rect(0, 0, nw, nh))
		draw.CatmullRom.Scale(dst, dst.Bounds(), img, bounds, draw.Over, nil)
		resized = dst
		width = nw
		height = nh
	}

	var out bytes.Buffer
	contentType := "image/png"
	switch strings.ToLower(format) {
	case "jpeg":
		if err := jpeg.Encode(&out, resized, &jpeg.Options{Quality: 85}); err != nil {
			return nil, "", 0, 0, fmt.Errorf("encode avatar jpeg: %w", err)
		}
		contentType = "image/jpeg"
	case "png":
		if err := png.Encode(&out, resized); err != nil {
			return nil, "", 0, 0, fmt.Errorf("encode avatar png: %w", err)
		}
		contentType = "image/png"
	case "webp":
		if err := png.Encode(&out, resized); err != nil {
			return nil, "", 0, 0, fmt.Errorf("encode avatar webp->png: %w", err)
		}
		contentType = "image/png"
	default:
		return nil, "", 0, 0, fmt.Errorf("unsupported avatar image format")
	}

	if err := domain.ValidateAvatarSize(out.Len()); err != nil {
		return nil, "", 0, 0, err
	}
	return out.Bytes(), contentType, width, height, nil
}

func fitSize(width, height, maxDimension int) (int, int) {
	if width <= maxDimension && height <= maxDimension {
		return width, height
	}
	if width >= height {
		nh := (height * maxDimension) / width
		if nh < 1 {
			nh = 1
		}
		return maxDimension, nh
	}
	nw := (width * maxDimension) / height
	if nw < 1 {
		nw = 1
	}
	return nw, maxDimension
}
