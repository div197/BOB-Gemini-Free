package format

import (
	"encoding/base64"
	"fmt"
	"mime"
	"regexp"
	"strings"
)

const (
	// MaxInlineImageBytes bounds the decoded representation before it reaches
	// the upload/compression layer. The request-body limit is a separate guard;
	// this one also protects direct library callers from oversized base64.
	MaxInlineImageBytes     = 20 << 20
	MaxInlineImageMIMEBytes = 128
)

// decodeInlineImageData validates and decodes a provider inline image before
// it is converted into an upload request. Invalid image data must be an error,
// not an omitted attachment that changes the meaning of the prompt.
func decodeInlineImageData(encoded, declaredMIME string) ([]byte, string, error) {
	declaredMIME = strings.TrimSpace(declaredMIME)
	if declaredMIME == "" {
		declaredMIME = "image/png"
	}
	if len(declaredMIME) > MaxInlineImageMIMEBytes || strings.ContainsAny(declaredMIME, "\r\n\x00") {
		return nil, "", fmt.Errorf("inline image MIME type is invalid")
	}
	mediaType, _, err := mime.ParseMediaType(declaredMIME)
	if err != nil || !strings.HasPrefix(strings.ToLower(mediaType), "image/") {
		return nil, "", fmt.Errorf("inline image MIME type %q is not an image", declaredMIME)
	}

	encoded = strings.TrimSpace(encoded)
	if encoded == "" {
		return nil, "", fmt.Errorf("inline image data is empty")
	}
	if base64.StdEncoding.DecodedLen(len(encoded)) > MaxInlineImageBytes {
		return nil, "", fmt.Errorf("inline image exceeds %d decoded bytes", MaxInlineImageBytes)
	}
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, "", fmt.Errorf("inline image base64 is invalid: %w", err)
	}
	if len(decoded) == 0 {
		return nil, "", fmt.Errorf("inline image data is empty")
	}
	return decoded, mediaType, nil
}

var (
	reImageMarkdown     = regexp.MustCompile(`!\[(.*?)\]\((https?://[^\s\)]+)\)`)
	reGoogleUserContent = regexp.MustCompile(`(https://lh3\.googleusercontent\.com/[^\s\)\"\']+)`)
	reGenericImageURL   = regexp.MustCompile(`(https?://[^\s\)\"\']+\.(?:png|jpg|jpeg|webp|gif))`)
)

type ExtractedImage struct {
	URL           string
	RevisedPrompt string
}

// ExtractImageURLsFromText extracts image URLs and markdown alt texts from Gemini output.
func ExtractImageURLsFromText(text string) []ExtractedImage {
	var results []ExtractedImage
	seen := make(map[string]bool)

	// 1. Check markdown image syntax: ![alt](url)
	matches := reImageMarkdown.FindAllStringSubmatch(text, -1)
	for _, m := range matches {
		if len(m) > 2 {
			alt := strings.TrimSpace(m[1])
			url := strings.TrimSpace(m[2])
			if url != "" && !seen[url] {
				seen[url] = true
				results = append(results, ExtractedImage{
					URL:           url,
					RevisedPrompt: alt,
				})
			}
		}
	}

	// 2. Check direct Google User Content URLs
	gMatches := reGoogleUserContent.FindAllStringSubmatch(text, -1)
	for _, m := range gMatches {
		if len(m) > 1 {
			url := strings.TrimSpace(m[1])
			if url != "" && !seen[url] {
				seen[url] = true
				results = append(results, ExtractedImage{
					URL: url,
				})
			}
		}
	}

	// 3. Check generic image URLs (.png, .jpg, etc.)
	genMatches := reGenericImageURL.FindAllStringSubmatch(text, -1)
	for _, m := range genMatches {
		if len(m) > 1 {
			url := strings.TrimSpace(m[1])
			if url != "" && !seen[url] {
				seen[url] = true
				results = append(results, ExtractedImage{
					URL: url,
				})
			}
		}
	}

	return results
}
