package server

import (
	"strings"
	"testing"
)

// Captured math is model output and is re-inserted after DOMPurify has run.
// When KaTeX is not defined yet (history replays before deferred scripts load)
// or is blocked, the fallback must stay escaped text, never raw HTML.
func TestMathFallbackEscapesCapturedModelOutput(t *testing.T) {
	html := string(playgroundHTML)

	if strings.Contains(html, "let renderedMath = rawMath;") {
		t.Fatal("KaTeX fallback inserts captured model math as raw HTML")
	}
	for _, marker := range []string{
		"let renderedMath = escapeHtml(rawMath);",
		"html.replaceAll(`@@@BOB_MATH_${i}@@@`, () => renderedMath);",
	} {
		if !strings.Contains(html, marker) {
			t.Fatalf("playground math restore is missing %q", marker)
		}
	}
}
