package gemini

import (
	"bytes"
	"fmt"
	"strings"
)

const maxStreamTextBytes = 32 << 20

var errStreamTextTooLarge = fmt.Errorf("upstream stream text exceeded %d bytes", maxStreamTextBytes)

type StreamParser struct {
	prevText string
	buf      []byte
	sawText  bool
}

func NewStreamParser() *StreamParser {
	return &StreamParser{
		buf: make([]byte, 0, 4096),
	}
}

// ResetBuffer clears the line buffer before a retry attempt while preserving prevText across retries.
func (p *StreamParser) ResetBuffer() {
	p.buf = p.buf[:0]
}

func (p *StreamParser) Feed(chunk string) ([]string, error) {
	if len(p.buf)+len(chunk) > maxStreamTextBytes {
		return nil, errStreamTextTooLarge
	}
	p.buf = append(p.buf, chunk...)

	if bytes.Contains(p.buf, []byte("BardErrorInfo")) && p.prevText == "" {
		if code, ok := IsBardError(string(p.buf)); ok {
			return nil, fmt.Errorf("%s", FormatBardError(code))
		}
	}

	var deltas []string
	for {
		idx := bytes.IndexByte(p.buf, '\n')
		if idx < 0 {
			break
		}

		line := string(p.buf[:idx])
		p.buf = p.buf[idx+1:]

		texts := ExtractTextsFromLine(line)
		for _, t := range texts {
			if len(t) > maxStreamTextBytes {
				return nil, errStreamTextTooLarge
			}
			if CleanText(t, true) != "" {
				p.sawText = true
			}
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

// Flush processes a final unterminated upstream line. Gemini normally emits
// newline-delimited records, but a connection can close after delivering a
// complete record without its trailing newline. Treating the buffered bytes
// as one final line prevents a valid terminal frame from being discarded.
func (p *StreamParser) Flush() ([]string, error) {
	if len(p.buf) == 0 {
		return nil, nil
	}
	return p.Feed("\n")
}

// HasText reports whether the parser saw usable model text in this stream.
// A syntactically readable response with no text is not a successful
// generation; callers need an explicit protocol error instead of fabricating a
// normal stop response.
func (p *StreamParser) HasText() bool {
	return p != nil && p.sawText
}
