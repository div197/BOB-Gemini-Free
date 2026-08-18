package format

import (
	"regexp"
	"strings"
)

// StreamDeltaType represents the classification of a streamed chunk.
type StreamDeltaType int

const (
	DeltaNone StreamDeltaType = iota
	DeltaThinking
	DeltaContent
)

// StreamChunk represents a parsed chunk from the model stream.
type StreamChunk struct {
	Type                StreamDeltaType
	Text                string
	TransitionToContent bool // true on the exact chunk where thinking finishes and content begins
}

type splitterState int

const (
	stateDetecting splitterState = iota
	stateInThinking
	stateInContent
)

// ThinkingStreamSplitter cleanly separates reasoning tokens from response content
// during real-time streaming across both OpenAI reasoning and Anthropic thinking protocols.
type ThinkingStreamSplitter struct {
	buf          string
	state        splitterState
	fullThinking strings.Builder
	fullContent  strings.Builder
}

// NewThinkingStreamSplitter instantiates a fresh stream splitter.
func NewThinkingStreamSplitter() *ThinkingStreamSplitter {
	return &ThinkingStreamSplitter{
		state: stateDetecting,
	}
}

// Feed processes a raw stream delta and returns typed chunks (thinking or content).
func (s *ThinkingStreamSplitter) Feed(delta string) []StreamChunk {
	s.buf += delta
	var results []StreamChunk

	for len(s.buf) > 0 {
		switch s.state {
		case stateDetecting:
			trimmed := strings.TrimLeft(s.buf, " \t\r\n")

			// Check for complete thought tag opening
			if strings.HasPrefix(trimmed, "```thought\n") || strings.HasPrefix(trimmed, "```thinking\n") ||
				strings.HasPrefix(trimmed, "```thought\r\n") || strings.HasPrefix(trimmed, "```thinking\r\n") {
				nlIdx := strings.Index(s.buf, "\n")
				s.buf = s.buf[nlIdx+1:]
				s.state = stateInThinking
				continue
			}

			// Check if buffer is a potential prefix of the opening tag (e.g. "`", "``", "```thought")
			if strings.HasPrefix("```thought\n", trimmed) || strings.HasPrefix("```thinking\n", trimmed) ||
				strings.HasPrefix("```thought\r\n", trimmed) || strings.HasPrefix("```thinking\r\n", trimmed) {
				// Wait for more tokens unless buffer is getting abnormally large
				if len(s.buf) < 30 {
					return results
				}
			}

			// Not a thinking block — transition directly to content
			s.state = stateInContent
			chunk := s.buf
			s.buf = ""
			s.fullContent.WriteString(chunk)
			results = append(results, StreamChunk{Type: DeltaContent, Text: chunk})
			return results

		case stateInThinking:
			idx := strings.Index(s.buf, "\n```")
			if idx != -1 {
				// Closing fence found!
				thinkingPart := s.buf[:idx]
				if len(thinkingPart) > 0 {
					s.fullThinking.WriteString(thinkingPart)
					results = append(results, StreamChunk{Type: DeltaThinking, Text: thinkingPart})
				}

				// Skip past closing fence and any immediate trailing whitespace/newlines
				after := s.buf[idx+len("\n```"):]
				for len(after) > 0 && (after[0] == ' ' || after[0] == '\r' || after[0] == '\n' || after[0] == '\t') {
					after = after[1:]
				}

				s.buf = ""
				s.state = stateInContent

				if len(after) > 0 {
					s.fullContent.WriteString(after)
					results = append(results, StreamChunk{
						Type:                DeltaContent,
						Text:                after,
						TransitionToContent: true,
					})
				} else {
					results = append(results, StreamChunk{
						Type:                DeltaNone,
						Text:                "",
						TransitionToContent: true,
					})
				}
				return results
			}

			// Check if buffer ends with partial fence ("\n", "\n`", "\n``")
			safeLen := len(s.buf)
			if strings.HasSuffix(s.buf, "\n``") {
				safeLen -= 3
			} else if strings.HasSuffix(s.buf, "\n`") {
				safeLen -= 2
			} else if strings.HasSuffix(s.buf, "\n") {
				safeLen -= 1
			}

			if safeLen > 0 {
				toEmit := s.buf[:safeLen]
				s.buf = s.buf[safeLen:]
				s.fullThinking.WriteString(toEmit)
				results = append(results, StreamChunk{Type: DeltaThinking, Text: toEmit})
			}
			return results

		case stateInContent:
			toEmit := s.buf
			s.buf = ""
			s.fullContent.WriteString(toEmit)
			results = append(results, StreamChunk{Type: DeltaContent, Text: toEmit})
			return results
		}
	}

	return results
}

// Flush emits any remaining buffered text at stream completion.
func (s *ThinkingStreamSplitter) Flush() []StreamChunk {
	var results []StreamChunk
	if len(s.buf) == 0 {
		return results
	}

	remaining := s.buf
	s.buf = ""

	switch s.state {
	case stateDetecting, stateInContent:
		s.fullContent.WriteString(remaining)
		results = append(results, StreamChunk{Type: DeltaContent, Text: remaining})
	case stateInThinking:
		// Strip closing fence if present in buffer
		trimmed := strings.TrimSpace(remaining)
		if trimmed == "```" || trimmed == "\n```" {
			results = append(results, StreamChunk{Type: DeltaNone, Text: "", TransitionToContent: true})
		} else {
			s.fullThinking.WriteString(remaining)
			results = append(results, StreamChunk{Type: DeltaThinking, Text: remaining})
		}
	}

	return results
}

// GetFullThinking returns the complete accumulated reasoning text.
func (s *ThinkingStreamSplitter) GetFullThinking() string {
	return strings.TrimSpace(s.fullThinking.String())
}

// GetFullContent returns the complete accumulated clean response content.
func (s *ThinkingStreamSplitter) GetFullContent() string {
	return strings.TrimSpace(s.fullContent.String())
}

var reThinkingBlock = regexp.MustCompile(`(?s)\x60\x60\x60(?:thought|thinking)\s*\n(.*?)\n\x60\x60\x60`)

// ExtractThinking extracts reasoning/thinking traces (e.g. ```thought\n...\n```) from Gemini responses,
// separating internal thinking tokens from the clean user-facing response.
func ExtractThinking(text string) (thinking string, cleanContent string) {
	matches := reThinkingBlock.FindStringSubmatch(text)
	if len(matches) == 2 {
		thinking = strings.TrimSpace(matches[1])
		cleanContent = strings.TrimSpace(reThinkingBlock.ReplaceAllString(text, ""))
		return thinking, cleanContent
	}
	return "", text
}
