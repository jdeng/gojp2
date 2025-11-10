package gojp2

import (
	"errors"
	"fmt"
	"image"
	"io"

	"github.com/jdeng/gojp2/jp2"
)

// Decode reads a JPEG 2000 codestream from data and returns it as a Go image.
// Single-component images decode to *image.Gray, while 3/4-component images
// decode to *image.NRGBA. Unsupported component layouts return an error.
func Decode(data []byte) (image.Image, error) {
	raw, err := jp2.Decode(data)
	if err != nil {
		return nil, err
	}

	componentCount := len(raw.Components)
	if componentCount == 0 {
		return nil, errors.New("jp2: decoded image has no components")
	}

	switch componentCount {
	case 1:
		return buildGray(raw), nil
	case 2:
		if raw.AlphaIndex < 0 {
			return nil, errors.New("jp2: two components without alpha are unsupported")
		}
		return buildGrayAlpha(raw), nil
	case 3, 4:
		return buildNRGBA(raw), nil
	default:
		return nil, fmt.Errorf("jp2: unsupported component count %d", componentCount)
	}
}

// DecodeReader reads JPEG 2000 data from r and decodes it.
func DecodeReader(r io.Reader) (image.Image, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("jp2: read source: %w", err)
	}
	return Decode(data)
}

func buildGray(img *jp2.Image) image.Image {
	rect := image.Rect(0, 0, img.Width, img.Height)
	out := image.NewGray(rect)
	copy(out.Pix, img.Components[0])
	return out
}

func buildGrayAlpha(img *jp2.Image) image.Image {
	rect := image.Rect(0, 0, img.Width, img.Height)
	out := image.NewNRGBA(rect)

	lumaIdx := 0
	if img.AlphaIndex == 0 && len(img.Components) > 1 {
		lumaIdx = 1
	}
	luma := img.Components[lumaIdx]
	alpha := img.Components[img.AlphaIndex]
	pixels := img.Width * img.Height

	for i := 0; i < pixels; i++ {
		gray := luma[i]
		base := i * 4
		out.Pix[base+0] = gray
		out.Pix[base+1] = gray
		out.Pix[base+2] = gray
		out.Pix[base+3] = alpha[i]
	}

	return out
}

func buildNRGBA(img *jp2.Image) image.Image {
	rect := image.Rect(0, 0, img.Width, img.Height)
	out := image.NewNRGBA(rect)

	n := img.Width * img.Height
	r := img.Components[0]
	g := img.Components[0]
	b := img.Components[0]
	if len(img.Components) > 1 {
		g = img.Components[1]
	}
	if len(img.Components) > 2 {
		b = img.Components[2]
	}

	var alpha []byte
	if img.AlphaIndex >= 0 && img.AlphaIndex < len(img.Components) {
		alpha = img.Components[img.AlphaIndex]
	}

	for i := 0; i < n; i++ {
		base := i * 4
		out.Pix[base+0] = r[i]
		out.Pix[base+1] = g[i]
		out.Pix[base+2] = b[i]
		if alpha != nil {
			out.Pix[base+3] = alpha[i]
		} else {
			out.Pix[base+3] = 0xFF
		}
	}

	return out
}
