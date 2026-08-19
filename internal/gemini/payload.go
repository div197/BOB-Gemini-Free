package gemini

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/div197/bob-gemini-free/internal/config"
)

func uuidV4() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	h := hex.EncodeToString(b[:])
	return fmt.Sprintf("%s-%s-%s-%s-%s", h[0:8], h[8:12], h[12:16], h[16:20], h[20:32])
}

const GeminiPayloadSize = 102

// BuildBody constructs the form-encoded payload (`f.req`) expected by Gemini's frontend RPC endpoint.
// Gemini web uses a sparse JSON array (102 elements) where specific indices represent payload parameters:
// - Index 0: Prompt text and image attachment references
// - Index 17: Reasoning/thinking mode depth
// - Index 59: Request UUID
// - Index 79: Target model mode ID
func BuildBodyWithAt(prompt string, modelID, thinkMode int, fileRefs []string, extra map[int]any, cfg config.Config, at string) string {
	inner := make([]any, GeminiPayloadSize)

	if len(fileRefs) > 0 {
		refs := make([]any, len(fileRefs))
		for i, ref := range fileRefs {
			refs[i] = []any{nil, nil, ref}
		}
		inner[0] = []any{prompt, 0, nil, refs, nil, nil, 0}
	} else {
		inner[0] = []any{prompt, 0, nil, nil, nil, nil, 0}
	}

	inner[1] = []any{"en"}
	inner[2] = []any{"", "", "", nil, nil, nil, nil, nil, nil, ""}
	inner[6] = []any{0}
	inner[7] = 1
	inner[10] = 1
	inner[11] = 0
	inner[17] = []any{[]any{thinkMode}}
	inner[18] = 0
	inner[27] = 1
	inner[30] = []any{4}
	inner[41] = []any{2}
	inner[53] = 0
	inner[59] = uuidV4()
	inner[61] = []any{}
	inner[68] = 1
	inner[79] = modelID

	if extra != nil {
		for k, v := range extra {
			if k >= 0 && k < len(inner) {
				inner[k] = v
			}
		}
	}

	innerJSON, _ := json.Marshal(inner)
	outer := []any{nil, string(innerJSON)}
	outerJSON, _ := json.Marshal(outer)

	form := url.Values{}
	form.Set("f.req", string(outerJSON))
	token := at
	if token == "" {
		token = cfg.XSRFToken
	}
	if token != "" {
		form.Set("at", token)
	}

	return form.Encode()
}

func BuildBody(prompt string, modelID, thinkMode int, fileRefs []string, extra map[int]any, cfg config.Config) string {
	return BuildBodyWithAt(prompt, modelID, thinkMode, fileRefs, extra, cfg, "")
}

func BuildURL(cfg config.Config) string {
	reqid := time.Now().Unix() % 1000000
	prefix := AccountPrefix(cfg.AuthUser)
	return fmt.Sprintf(
		"https://gemini.google.com%s/_/BardChatUi/data/assistant.lamda.BardFrontendService/StreamGenerate?bl=%s&hl=en&_reqid=%d&rt=c",
		prefix,
		cfg.GeminiBL,
		reqid,
	)
}
