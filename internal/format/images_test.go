package format

import (
	"testing"
)

func TestExtractImageURLsFromText(t *testing.T) {
	text := `Here is the image you requested:
![A futuristic city with flying cars](https://lh3.googleusercontent.com/a/ABC123xyz=s1024)

And another image:
https://example.com/sunset.png
`

	images := ExtractImageURLsFromText(text)
	if len(images) < 2 {
		t.Fatalf("Expected at least 2 images, got %d", len(images))
	}

	if images[0].URL != "https://lh3.googleusercontent.com/a/ABC123xyz=s1024" {
		t.Errorf("Unexpected image 0 URL: %s", images[0].URL)
	}
	if images[0].RevisedPrompt != "A futuristic city with flying cars" {
		t.Errorf("Unexpected image 0 RevisedPrompt: %s", images[0].RevisedPrompt)
	}

	if images[1].URL != "https://example.com/sunset.png" {
		t.Errorf("Unexpected image 1 URL: %s", images[1].URL)
	}
}
