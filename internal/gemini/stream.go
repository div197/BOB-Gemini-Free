package gemini

import (
	"fmt"
	"strings"
)

type StreamParser struct {
	prevText string
	buf      string
}

func NewStreamParser() *StreamParser {
	return &StreamParser{}
}

// ResetBuffer clears the line buffer before a retry attempt while preserving prevText across retries.
func (p *StreamParser) ResetBuffer() {
	p.buf = ""
}

func (p *StreamParser) Feed(chunk string) ([]string, error) {
	p.buf += chunk

	if strings.Contains(p.buf, "BardErrorInfo") {
		if code, ok := IsBardError(p.buf); ok {
			return nil, fmt.Errorf("Gemini upstream rejected request: BardErrorInfo [%s]", code)
		}
	}

	var deltas []string
	for strings.Contains(p.buf, "\n") {
		idx := strings.Index(p.buf, "\n")
		line := p.buf[:idx]
		p.buf = p.buf[idx+1:]

		texts := ExtractTextsFromLine(line)
		for _, t := range texts {
			// Skip if this text segment has already been fully processed or is identical to the last one.
			if t == p.prevText || strings.HasPrefix(p.prevText, t) {
				continue
			}

			// If new text starts with previously processed text, extract the delta
			if strings.HasPrefix(t, p.prevText) {
				delta := CleanText(t[len(p.prevText):], false)
				p.prevText = t
				if delta != "" {
					deltas = append(deltas, delta)
				}
				break // Successfully advanced the active candidate stream
			} else if p.prevText == "" {
				delta := CleanText(t, false)
				p.prevText = t
				if delta != "" {
					deltas = append(deltas, delta)
				}
				break
			} else if len(t) > len(p.prevText) {
				// Handle clean text boundary advancements without erroring
				delta := CleanText(t[len(p.prevText):], false)
				p.prevText = t
				if delta != "" {
					deltas = append(deltas, delta)
				}
				break
			}
		}
	}

	return deltas, nil
}
