package gemini

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"regexp"
	"strings"
	"sync"
	"testing"

	"github.com/div197/bob-gemini-free/internal/config"
)

// goldenBoQLine creates a deterministic fixture with the shape consumed by
// Gemini's wrb.fr response parser. The padding represents ignored upstream
// metadata and keeps the fixture above the parser's defensive size guards.
func goldenBoQLine(text string) string {
	parts := []any{[]any{nil, []any{text}}}
	inner := make([]any, 8)
	inner[4] = parts
	inner[7] = strings.Repeat("fixture-metadata-", 8)
	innerBytes, err := json.Marshal(inner)
	if err != nil {
		panic(err)
	}
	outer := []any{[]any{"wrb.fr", nil, string(innerBytes)}}
	outerBytes, err := json.Marshal(outer)
	if err != nil {
		panic(err)
	}
	return string(outerBytes) + "\n"
}

func TestGoldenPayloadPreservesSparseWirePositions(t *testing.T) {
	cfg := config.Default()
	refs := []string{"/blob/one", "/blob/two"}
	body := BuildBodyWithAt("नमस्ते Gemini", 3, 0, refs, map[int]any{88: "fixture-flag"}, cfg, "fixture-at")
	values, err := url.ParseQuery(body)
	if err != nil {
		t.Fatalf("ParseQuery: %v", err)
	}
	if got := values.Get("at"); got != "fixture-at" {
		t.Fatalf("at token = %q, want fixture-at", got)
	}

	var outer []any
	if err := json.Unmarshal([]byte(values.Get("f.req")), &outer); err != nil {
		t.Fatalf("decode outer f.req: %v", err)
	}
	if len(outer) != 2 || outer[1] == nil {
		t.Fatalf("unexpected outer payload shape: %#v", outer)
	}
	innerJSON, ok := outer[1].(string)
	if !ok {
		t.Fatalf("outer payload slot 1 has type %T", outer[1])
	}
	var inner []any
	if err := json.Unmarshal([]byte(innerJSON), &inner); err != nil {
		t.Fatalf("decode sparse inner payload: %v", err)
	}
	if len(inner) != GeminiPayloadSize {
		t.Fatalf("sparse payload length = %d, want %d", len(inner), GeminiPayloadSize)
	}

	if got := inner[0].([]any)[0]; got != "नमस्ते Gemini" {
		t.Errorf("prompt slot = %#v", got)
	}
	if got := inner[1]; !reflect.DeepEqual(got, []any{"en"}) {
		t.Errorf("language slot = %#v", got)
	}
	if got := inner[6]; !reflect.DeepEqual(got, []any{float64(0)}) {
		t.Errorf("slot 6 = %#v", got)
	}
	if got := inner[17]; !reflect.DeepEqual(got, []any{[]any{float64(0)}}) {
		t.Errorf("thinking slot = %#v", got)
	}
	if got := inner[79]; got != float64(3) {
		t.Errorf("model slot = %#v, want 3", got)
	}
	if got := inner[88]; got != "fixture-flag" {
		t.Errorf("extra slot = %#v", got)
	}

	attachment := inner[0].([]any)[3].([]any)
	if len(attachment) != len(refs) {
		t.Fatalf("attachment count = %d, want %d", len(attachment), len(refs))
	}
	for i, ref := range refs {
		got := attachment[i].([]any)[2]
		if got != ref {
			t.Errorf("attachment %d = %#v, want %q", i, got, ref)
		}
	}
	if ok, _ := regexp.MatchString(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`, inner[59].(string)); !ok {
		t.Errorf("request UUID has unexpected shape: %#v", inner[59])
	}
}

func TestGoldenResponseFixturesCoverTextAndSanitization(t *testing.T) {
	fixture := "ordinary response\nsecond line\nहिंदी उत्तर नमस्ते\n```python?code_reference&code_event_index=1\nprint('hidden')\n```\nhttp://googleusercontent.com/card_content/42\nhttps://support.google.com/gemini/answer/fixture" + strings.Repeat(".", 80)
	raw := goldenBoQLine("short") + goldenBoQLine(fixture)

	texts := ExtractTextsFromLine(strings.TrimSuffix(goldenBoQLine(fixture), "\n"))
	if len(texts) != 1 || texts[0] != fixture {
		t.Fatalf("parsed fixture texts = %#v", texts)
	}
	got, err := ExtractResponseText(raw)
	if err != nil {
		t.Fatalf("ExtractResponseText: %v", err)
	}
	if !strings.Contains(got, "ordinary response") || !strings.Contains(got, "हिंदी उत्तर नमस्ते") {
		t.Fatalf("response lost text or Unicode: %q", got)
	}
	if strings.Contains(got, "code_reference") || strings.Contains(got, "card_content") {
		t.Fatalf("response retained provider artifact: %q", got)
	}
	if !strings.Contains(got, "support.google.com/gemini/answer/fixture") {
		t.Fatalf("response lost citation URL: %q", got)
	}
}

func TestGoldenParserRejectsMalformedNestedFixtures(t *testing.T) {
	malformed := []string{
		"not-json\n",
		`[["wrb.fr",null,"not-json"]]` + "\n",
		`[["other",null,"[]"]]` + "\n",
	}
	for _, fixture := range malformed {
		if got := ExtractTextsFromLine(strings.TrimSuffix(fixture, "\n")); len(got) != 0 {
			t.Errorf("malformed fixture %q produced %#v", fixture, got)
		}
	}
	if _, err := ExtractResponseText("prefix BardErrorInfo [42901] suffix"); err == nil {
		t.Fatal("Bard error fixture was not rejected")
	}
}

func TestGoldenSAPISIDHashUsesDeterministicTimestamp(t *testing.T) {
	const timestamp int64 = 1700000000
	const sapisid = "fixture-sapisid"
	input := fmt.Sprintf("%d %s https://gemini.google.com", timestamp, sapisid)
	digest := sha1.Sum([]byte(input))
	want := fmt.Sprintf("SAPISIDHASH %d_%s", timestamp, hex.EncodeToString(digest[:]))
	if got := SAPISIDHashAt(sapisid, timestamp); got != want {
		t.Fatalf("SAPISIDHASH = %q, want %q", got, want)
	}
	if SAPISIDHashAt("", timestamp) != "" {
		t.Fatal("empty SAPISID should produce an empty authorization value")
	}
}

func TestGoldenStreamParserHandlesArbitraryBoundariesAndMalformedLines(t *testing.T) {
	parser := NewStreamParser()
	frames := []string{
		goldenBoQLine("Hello"),
		goldenBoQLine("Hello world"),
	}
	var deltas []string
	for _, frame := range frames {
		for start := 0; start < len(frame); {
			end := start + 3
			if end > len(frame) {
				end = len(frame)
			}
			got, err := parser.Feed(frame[start:end])
			if err != nil {
				t.Fatalf("Feed boundary %d:%d: %v", start, end, err)
			}
			deltas = append(deltas, got...)
			start = end
		}
	}
	if !reflect.DeepEqual(deltas, []string{"Hello", " world"}) {
		t.Fatalf("deltas = %#v, want [Hello, ' world']", deltas)
	}

	if got, err := parser.Feed(frames[1]); err != nil || len(got) != 0 {
		t.Fatalf("repeated cumulative frame = %#v, err=%v", got, err)
	}
	if _, err := parser.Feed("malformed nested json\n"); err != nil {
		t.Fatalf("malformed line should be ignored, got %v", err)
	}
}

func TestGoldenStreamParserFlushesTruncatedFinalFrame(t *testing.T) {
	parser := NewStreamParser()
	frame := strings.TrimSuffix(goldenBoQLine("complete frame delivered at EOF"+strings.Repeat(".", 80)), "\n")
	if got, err := parser.Feed(frame); err != nil || len(got) != 0 {
		t.Fatalf("partial frame feed = %#v, err=%v; expected buffering", got, err)
	}
	got, err := parser.Flush()
	if err != nil {
		t.Fatalf("Flush: %v", err)
	}
	if !reflect.DeepEqual(got, []string{"complete frame delivered at EOF" + strings.Repeat(".", 80)}) {
		t.Fatalf("flushed deltas = %#v", got)
	}
}

type goldenFaultReader struct {
	data    *bytes.Reader
	fault   error
	faulted bool
}

func (r *goldenFaultReader) Read(p []byte) (int, error) {
	if !r.faulted {
		n, err := r.data.Read(p)
		if n > 0 {
			return n, nil
		}
		if err == io.EOF {
			r.faulted = true
		}
	}
	return 0, r.fault
}

type goldenHTTPResponse struct {
	response *http.Response
	err      error
}

type goldenRequester struct {
	mu        sync.Mutex
	responses []goldenHTTPResponse
	called    int
}

func (r *goldenRequester) Do(req *http.Request) (*http.Response, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if req != nil && req.URL != nil && strings.Contains(req.URL.Path, "/app") {
		return goldenOK(io.NopCloser(strings.NewReader(`"SNlM0e":"fixture-at","cfb2h":"fixture-bl"`))), nil
	}
	if r.called >= len(r.responses) {
		return nil, errors.New("unexpected fixture request")
	}
	response := r.responses[r.called]
	r.called++
	return response.response, response.err
}

func goldenOK(body io.ReadCloser) *http.Response {
	return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: body}
}

func TestGoldenStreamRetryDeduplicatesCumulativeFrames(t *testing.T) {
	first := goldenBoQLine("Hello") + goldenBoQLine("Hello world")
	second := goldenBoQLine("Hello") + goldenBoQLine("Hello world")
	requester := &goldenRequester{responses: []goldenHTTPResponse{
		{response: goldenOK(io.NopCloser(&goldenFaultReader{data: bytes.NewReader([]byte(first)), fault: errors.New("connection reset")}))},
		{response: goldenOK(io.NopCloser(strings.NewReader(second)))},
	}}
	cfg := config.Default()
	cfg.RetryAttempts = 2
	cfg.RetryDelaySec = 0
	client := &Client{
		Cfg:     cfg,
		HTTP:    requester,
		Cookies: NewCookieCache(""),
		Pool:    NewCookiePool(),
		Logf:    func(string, ...any) {},
	}
	var deltas []string
	err := client.GenerateStreamContext(context.Background(), "fixture", 1, 4, nil, nil, func(delta string) error {
		deltas = append(deltas, delta)
		return nil
	})
	if err != nil {
		t.Fatalf("GenerateStreamContext: %v", err)
	}
	if !reflect.DeepEqual(deltas, []string{"Hello", " world"}) {
		t.Fatalf("retry deltas = %#v", deltas)
	}
	requester.mu.Lock()
	called := requester.called
	requester.mu.Unlock()
	if called != 2 {
		t.Fatalf("fixture request count = %d, want 2", called)
	}
}
