package format

import (
	"regexp"
	"strings"
)

// Citation represents a web source reference or citation grounded in model output.
type Citation struct {
	Index   int    `json:"index"`
	Title   string `json:"title"`
	URL     string `json:"url"`
	Snippet string `json:"snippet,omitempty"`
}

var (
	reCitationLink = regexp.MustCompile(`\[([^\]]+)\]\((https?://[^\s\)]+)\)`)
)

// ExtractCitations parses web URLs and citations from the model response text.
func ExtractCitations(text string) ([]Citation, string) {
	var citations []Citation
	seen := make(map[string]bool)

	matches := reCitationLink.FindAllStringSubmatch(text, -1)
	idx := 1

	for _, m := range matches {
		if len(m) > 2 {
			title := strings.TrimSpace(m[1])
			rawURL := strings.TrimSpace(m[2])

			// Filter out image links
			if strings.HasSuffix(rawURL, ".png") || strings.HasSuffix(rawURL, ".jpg") || strings.HasSuffix(rawURL, ".jpeg") || strings.HasSuffix(rawURL, ".webp") {
				continue
			}

			if rawURL != "" && !seen[rawURL] {
				seen[rawURL] = true
				citations = append(citations, Citation{
					Index: idx,
					Title: title,
					URL:   rawURL,
				})
				idx++
			}
		}
	}

	return citations, text
}
