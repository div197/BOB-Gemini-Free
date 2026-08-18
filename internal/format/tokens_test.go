package format

import (
	"testing"

	"github.com/div197/bob-gemini-free/internal/models"
)

func TestEstimateTokens(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		minCount int
		maxCount int
	}{
		{
			name:     "Empty String",
			input:    "",
			minCount: 0,
			maxCount: 0,
		},
		{
			name:     "Simple English",
			input:    "Hello, world! This is a test.",
			minCount: 5,
			maxCount: 15,
		},
		{
			name:     "Devanagari / Hindi",
			input:    "नमस्ते भारत, यह एक परीक्षण है।",
			minCount: 5,
			maxCount: 25,
		},
		{
			name:     "Code snippet",
			input:    "package main\n\nfunc main() {\n\tprintln(\"hello\")\n}",
			minCount: 8,
			maxCount: 25,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EstimateTokens(tt.input)
			if got < tt.minCount || got > tt.maxCount {
				t.Errorf("EstimateTokens(%q) = %d; want between %d and %d", tt.input, got, tt.minCount, tt.maxCount)
			}
		})
	}
}

func TestCountGoogleTokens(t *testing.T) {
	req := models.GoogleGenerateRequest{
		Contents: []models.GoogleContent{
			{
				Role: "user",
				Parts: []models.GooglePart{
					{Text: "Explain how quantum computing works."},
					{InlineData: &models.GoogleInlineData{MIMEType: "image/png", Data: "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg=="}},
				},
			},
		},
	}

	count := CountGoogleTokens(req)
	// Text (~5-8 tokens) + 1 Image (258 tokens) -> should be >= 260
	if count < 260 {
		t.Errorf("CountGoogleTokens() = %d, expected >= 260", count)
	}
}

func TestCountOpenAITokens(t *testing.T) {
	req := models.OpenAIChatRequest{
		Model: "gemini-3.7-flash",
		Messages: []models.OpenAIMessage{
			{Role: "system", Content: "You are a helpful assistant."},
			{Role: "user", Content: "Write a binary search algorithm in Go."},
		},
	}

	count := CountOpenAITokens(req)
	if count < 10 {
		t.Errorf("CountOpenAITokens() = %d, expected >= 10", count)
	}
}
