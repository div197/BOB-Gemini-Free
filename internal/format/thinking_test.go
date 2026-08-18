package format

import (
	"testing"
)

func TestThinkingStreamSplitter_SingleChunkWithThinking(t *testing.T) {
	splitter := NewThinkingStreamSplitter()
	raw := "```thought\nAnalyzing the problem step by step.\n1+1=2.\n```\nHere is the final answer: 2."

	chunks := splitter.Feed(raw)
	flushChunks := splitter.Flush()
	allChunks := append(chunks, flushChunks...)

	var thinkingText, contentText string
	for _, ch := range allChunks {
		if ch.Type == DeltaThinking {
			thinkingText += ch.Text
		} else if ch.Type == DeltaContent {
			contentText += ch.Text
		}
	}

	if thinkingText != "Analyzing the problem step by step.\n1+1=2." {
		t.Errorf("Unexpected thinking text: %q", thinkingText)
	}
	if contentText != "Here is the final answer: 2." {
		t.Errorf("Unexpected content text: %q", contentText)
	}
	if splitter.GetFullThinking() != "Analyzing the problem step by step.\n1+1=2." {
		t.Errorf("GetFullThinking mismatch: %q", splitter.GetFullThinking())
	}
	if splitter.GetFullContent() != "Here is the final answer: 2." {
		t.Errorf("GetFullContent mismatch: %q", splitter.GetFullContent())
	}
}

func TestThinkingStreamSplitter_MultiChunkStream(t *testing.T) {
	splitter := NewThinkingStreamSplitter()
	deltas := []string{
		"```",
		"thought\n",
		"Let me ",
		"think ",
		"carefully.\n",
		"```\n",
		"Hello ",
		"world!",
	}

	var thinkingDeltas []string
	var contentDeltas []string
	var hadTransition bool

	for _, d := range deltas {
		for _, ch := range splitter.Feed(d) {
			if ch.TransitionToContent {
				hadTransition = true
			}
			if ch.Type == DeltaThinking && ch.Text != "" {
				thinkingDeltas = append(thinkingDeltas, ch.Text)
			} else if ch.Type == DeltaContent && ch.Text != "" {
				contentDeltas = append(contentDeltas, ch.Text)
			}
		}
	}
	for _, ch := range splitter.Flush() {
		if ch.TransitionToContent {
			hadTransition = true
		}
		if ch.Type == DeltaThinking && ch.Text != "" {
			thinkingDeltas = append(thinkingDeltas, ch.Text)
		} else if ch.Type == DeltaContent && ch.Text != "" {
			contentDeltas = append(contentDeltas, ch.Text)
		}
	}

	if !hadTransition {
		t.Errorf("Expected TransitionToContent flag to be set")
	}
	if splitter.GetFullThinking() != "Let me think carefully." {
		t.Errorf("GetFullThinking mismatch: %q", splitter.GetFullThinking())
	}
	if splitter.GetFullContent() != "Hello world!" {
		t.Errorf("GetFullContent mismatch: %q", splitter.GetFullContent())
	}
}

func TestThinkingStreamSplitter_NoThinking(t *testing.T) {
	splitter := NewThinkingStreamSplitter()
	deltas := []string{
		"Just ",
		"a normal ",
		"response ",
		"without reasoning.",
	}

	var contentText string
	for _, d := range deltas {
		for _, ch := range splitter.Feed(d) {
			if ch.Type == DeltaThinking {
				t.Errorf("Did not expect thinking chunks in non-thinking response")
			}
			contentText += ch.Text
		}
	}
	for _, ch := range splitter.Flush() {
		contentText += ch.Text
	}

	if contentText != "Just a normal response without reasoning." {
		t.Errorf("Content mismatch: %q", contentText)
	}
	if splitter.GetFullThinking() != "" {
		t.Errorf("Expected empty full thinking, got %q", splitter.GetFullThinking())
	}
	if splitter.GetFullContent() != "Just a normal response without reasoning." {
		t.Errorf("GetFullContent mismatch: %q", splitter.GetFullContent())
	}
}

func TestThinkingStreamSplitter_CodeBlockInContent(t *testing.T) {
	splitter := NewThinkingStreamSplitter()
	deltas := []string{
		"Here is some Python code:\n",
		"```python\n",
		"def hello():\n",
		"    return 'world'\n",
		"```\n",
		"Done!",
	}

	for _, d := range deltas {
		for _, ch := range splitter.Feed(d) {
			if ch.Type == DeltaThinking {
				t.Errorf("Code block in content must not be identified as thinking")
			}
		}
	}
	for _, ch := range splitter.Flush() {
		if ch.Type == DeltaThinking {
			t.Errorf("Code block in content must not be identified as thinking")
		}
	}

	if splitter.GetFullThinking() != "" {
		t.Errorf("Expected empty thinking, got %q", splitter.GetFullThinking())
	}
	expected := "Here is some Python code:\n```python\ndef hello():\n    return 'world'\n```\nDone!"
	if splitter.GetFullContent() != expected {
		t.Errorf("Content mismatch:\nGot: %q\nWant: %q", splitter.GetFullContent(), expected)
	}
}
