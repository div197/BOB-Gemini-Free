package multimodal

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	_ "image/gif"
	"image/jpeg"
	_ "image/png"
	"math"

	"golang.org/x/image/draw"
	_ "golang.org/x/image/webp"
)

const (
	MaxImageByteSize        = 1024 * 1024      // 1 MB
	MaxImageB64Size         = 1300000          // ~1 MB base64
	MaxImageInputBytes      = 32 * 1024 * 1024 // bounded by the gateway request limit
	MaxImageDimension       = 1024
	MaxImageSourceDimension = 8192
	MaxImagePixels          = 16 * 1024 * 1024
	DefaultJPEGQuality      = 75
)

func CompressImageBytesIfNeeded(imgData []byte, mime string, maxBytes int) ([]byte, string, error) {
	if maxBytes <= 0 {
		maxBytes = MaxImageByteSize
	}
	if err := validateImageData(imgData); err != nil {
		return imgData, mime, err
	}

	if len(imgData) <= maxBytes && mime != "" {
		return imgData, mime, nil
	}

	img, _, err := image.Decode(bytes.NewReader(imgData))
	if err != nil {
		return imgData, mime, fmt.Errorf("failed to decode image format: %w", err)
	}

	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	if width > MaxImageDimension || height > MaxImageDimension {
		ratio := math.Min(float64(MaxImageDimension)/float64(width), float64(MaxImageDimension)/float64(height))
		width = maxInt(1, int(float64(width)*ratio))
		height = maxInt(1, int(float64(height)*ratio))
	}

	// A quality-only pass is not sufficient for screenshots and noisy camera
	// images. Reduce quality first, then dimensions, until the actual encoded
	// bytes fit the caller's budget. Never return an over-budget image.
	dstImg := image.Image(img)
	if dstImg.Bounds().Dx() != width || dstImg.Bounds().Dy() != height {
		dstImg = resizeImage(dstImg, width, height)
	}
	for {
		for _, quality := range []int{DefaultJPEGQuality, 60, 45, 30} {
			var buf bytes.Buffer
			if err := jpeg.Encode(&buf, dstImg, &jpeg.Options{Quality: quality}); err != nil {
				return imgData, mime, fmt.Errorf("failed to encode jpeg: %w", err)
			}
			if buf.Len() <= maxBytes {
				return append([]byte(nil), buf.Bytes()...), "image/jpeg", nil
			}
		}

		if width == 1 && height == 1 {
			return imgData, mime, fmt.Errorf("encoded image cannot fit within %d bytes", maxBytes)
		}
		nextWidth := maxInt(1, int(float64(width)*0.75))
		nextHeight := maxInt(1, int(float64(height)*0.75))
		if nextWidth == width && width > 1 {
			nextWidth = width - 1
		}
		if nextHeight == height && height > 1 {
			nextHeight = height - 1
		}
		width, height = nextWidth, nextHeight
		dstImg = resizeImage(dstImg, width, height)
	}
}

func resizeImage(src image.Image, width, height int) image.Image {
	dst := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.CatmullRom.Scale(dst, dst.Bounds(), src, src.Bounds(), draw.Over, nil)
	return dst
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func validateImageData(imgData []byte) error {
	if len(imgData) == 0 {
		return fmt.Errorf("image data is empty")
	}
	if len(imgData) > MaxImageInputBytes {
		return fmt.Errorf("image input exceeded %d bytes", MaxImageInputBytes)
	}
	config, _, err := image.DecodeConfig(bytes.NewReader(imgData))
	if err != nil {
		return fmt.Errorf("failed to inspect image dimensions: %w", err)
	}
	if config.Width <= 0 || config.Height <= 0 {
		return fmt.Errorf("image dimensions are invalid")
	}
	if config.Width > MaxImageSourceDimension || config.Height > MaxImageSourceDimension {
		return fmt.Errorf("image dimensions exceed %d pixels", MaxImageSourceDimension)
	}
	if uint64(config.Width)*uint64(config.Height) > MaxImagePixels {
		return fmt.Errorf("image contains more than %d pixels", MaxImagePixels)
	}
	return nil
}

func CompressIfNeeded(b64 string, maxSize int) (string, error) {
	if maxSize <= 0 {
		maxSize = MaxImageB64Size
	}

	if len(b64) <= maxSize {
		if len(b64) > ((MaxImageInputBytes*4 + 2) / 3) {
			return "", fmt.Errorf("base64 image input exceeded %d bytes", (MaxImageInputBytes*4+2)/3)
		}
		return b64, nil
	}
	if len(b64) > ((MaxImageInputBytes*4 + 2) / 3) {
		return "", fmt.Errorf("base64 image input exceeded %d bytes", (MaxImageInputBytes*4+2)/3)
	}

	imgData, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64 image: %w", err)
	}

	// Convert base64 char limit to raw byte limit: base64 expands by ~4/3
	// so raw bytes ≈ base64_chars * 3/4
	maxBytes := (maxSize * 3) / 4
	if maxBytes <= 0 {
		maxBytes = MaxImageByteSize
	}

	compressedBytes, _, err := CompressImageBytesIfNeeded(imgData, "image/jpeg", maxBytes)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(compressedBytes), nil
}
