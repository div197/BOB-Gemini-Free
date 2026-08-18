package gemini

import (
	"testing"
)

func TestCleanText(t *testing.T) {
	input := "Hello world\n```python?code_reference&code_event_index=1\nprint('hi')```\nhttp://googleusercontent.com/card_content/123\nExtra text"
	cleaned := CleanText(input, true)
	expected := "Hello world\nExtra text"
	if cleaned != expected {
		t.Errorf("CleanText got %q, want %q", cleaned, expected)
	}
}

func TestIsBardError(t *testing.T) {
	raw := "something BardErrorInfo [1024] happened"
	code, ok := IsBardError(raw)
	if !ok || code != "1024" {
		t.Errorf("IsBardError got (%q, %t), want (\"1024\", true)", code, ok)
	}
}
