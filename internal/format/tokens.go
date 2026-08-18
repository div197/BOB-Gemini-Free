package format

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/div197/bob-gemini-free/internal/models"
)

// TokensPerImageStandard defines the token cost for a standard 1024x1024 image in Gemini.
const TokensPerImageStandard = 258

// EstimateTokens calculates an accurate token count for arbitrary text.
// It accounts for English words, numbers, punctuation, CJK ideographs,
// Devanagari/Indic scripts, emojis, and whitespace subwords.
func EstimateTokens(text string) int {
	if len(text) == 0 {
		return 0
	}

	var tokens int
	var currentWord strings.Builder

	flushWord := func() {
		if currentWord.Len() == 0 {
			return
		}
		w := currentWord.String()
		currentWord.Reset()

		runeCount := utf8.RuneCountInString(w)
		if runeCount <= 4 {
			tokens += 1
		} else {
			// Subword approximation: ~3.5 to 4 characters per token
			tokens += (runeCount + 3) / 4
		}
	}

	for _, r := range text {
		// 1. CJK Ideographs / Hiragana / Katakana / Hangul: typically 1 token per character
		if unicode.Is(unicode.Han, r) || unicode.Is(unicode.Hiragana, r) ||
			unicode.Is(unicode.Katakana, r) || unicode.Is(unicode.Hangul, r) {
			flushWord()
			tokens += 1
			continue
		}

		// 2. Emojis and Special Symbols
		if unicode.Is(unicode.So, r) || unicode.Is(unicode.Sm, r) {
			flushWord()
			tokens += 2
			continue
		}

		// 3. Whitespace: word boundary
		if unicode.IsSpace(r) {
			flushWord()
			continue
		}

		// 4. Punctuation: token boundary
		if unicode.IsPunct(r) {
			flushWord()
			tokens += 1
			continue
		}

		// Standard alphanumeric + Devanagari/Indic script runes
		currentWord.WriteRune(r)
	}

	flushWord()

	if tokens == 0 && len(strings.TrimSpace(text)) > 0 {
		tokens = 1
	}

	return tokens
}

// CountGoogleTokens calculates the total tokens for a GoogleGenerateRequest (text + images).
func CountGoogleTokens(req models.GoogleGenerateRequest) int {
	total := 0

	// 1. System Instruction
	if req.SystemInstruction != nil {
		for _, part := range req.SystemInstruction.Parts {
			total += EstimateTokens(part.Text)
		}
	}

	// 2. Contents
	for _, content := range req.Contents {
		for _, part := range content.Parts {
			if part.Text != "" {
				total += EstimateTokens(part.Text)
			}
			if part.InlineData != nil && len(part.InlineData.Data) > 0 {
				total += TokensPerImageStandard
			}
		}
	}

	if total == 0 && len(req.Contents) > 0 {
		total = 1
	}

	return total
}

// CountOpenAITokens calculates the total tokens for an OpenAIChatRequest (messages + images + tools).
func CountOpenAITokens(req models.OpenAIChatRequest) int {
	total := 0

	for _, msg := range req.Messages {
		total += 4 // Message framing overhead (<|im_start|>role ... <|im_end|>)

		if str, ok := msg.Content.(string); ok {
			total += EstimateTokens(str)
		} else if parts, ok := msg.Content.([]any); ok {
			for _, p := range parts {
				if pMap, ok := p.(map[string]any); ok {
					pType, _ := pMap["type"].(string)
					if pType == "text" {
						if txt, ok := pMap["text"].(string); ok {
							total += EstimateTokens(txt)
						}
					} else if pType == "image_url" {
						total += TokensPerImageStandard
					}
				}
			}
		}

		if msg.ReasoningContent != "" {
			total += EstimateTokens(msg.ReasoningContent)
		}
	}

	// Tool declarations overhead
	for _, tool := range req.Tools {
		total += EstimateTokens(tool.Function.Name)
		total += EstimateTokens(tool.Function.Description)
		total += 10
	}

	total += 2 // Assistant response priming

	return total
}
