package format

import (
	"regexp"
	"strings"
)

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
