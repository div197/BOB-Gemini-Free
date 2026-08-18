package multimodal

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png"
	"math"

	"golang.org/x/image/draw"
)

const (
	MaxImageByteSize   = 1024 * 1024 // 1 MB
	MaxImageB64Size    = 1300000     // ~1 MB base64
	MaxImageDimension  = 1024
	DefaultJPEGQuality = 75
)

func CompressImageBytesIfNeeded(imgData []byte, mime string, maxBytes int) ([]byte, string, error) {
	if maxBytes <= 0 {
		maxBytes = MaxImageByteSize
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

	var dstImg image.Image = img
	if width > MaxImageDimension || height > MaxImageDimension {
		ratio := math.Min(float64(MaxImageDimension)/float64(width), float64(MaxImageDimension)/float64(height))
		newWidth := int(float64(width) * ratio)
		newHeight := int(float64(height) * ratio)
		dst := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))
		draw.CatmullRom.Scale(dst, dst.Bounds(), img, bounds, draw.Over, nil)
		dstImg = dst
	}

	var buf bytes.Buffer
	err = jpeg.Encode(&buf, dstImg, &jpeg.Options{Quality: DefaultJPEGQuality})
	if err != nil {
		return imgData, mime, fmt.Errorf("failed to encode jpeg: %w", err)
	}

	return buf.Bytes(), "image/jpeg", nil
}

func CompressIfNeeded(b64 string, maxSize int) (string, error) {
	if maxSize <= 0 {
		maxSize = MaxImageB64Size
	}

	if len(b64) <= maxSize {
		return b64, nil
	}

	imgData, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64 image: %w", err)
	}

	compressedBytes, _, err := CompressImageBytesIfNeeded(imgData, "image/jpeg", maxSize)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(compressedBytes), nil
}
