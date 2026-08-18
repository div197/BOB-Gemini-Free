package format

import (
	"testing"
)

func TestExtractCitations(t *testing.T) {
	text := "Here is some information about Go: [Go Documentation](https://golang.org/doc) and [ABCsteps](https://abcsteps.com)."
	citations, _ := ExtractCitations(text)

	if len(citations) != 2 {
		t.Fatalf("Expected 2 citations, got %d", len(citations))
	}

	if citations[0].Title != "Go Documentation" || citations[0].URL != "https://golang.org/doc" {
		t.Errorf("Unexpected citation 0: %+v", citations[0])
	}
	if citations[1].Title != "ABCsteps" || citations[1].URL != "https://abcsteps.com" {
		t.Errorf("Unexpected citation 1: %+v", citations[1])
	}
}
