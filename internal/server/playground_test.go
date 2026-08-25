package server

import (
	"strings"
	"testing"
)

func TestHostedStudioDoesNotProbeLoopbackOnStartup(t *testing.T) {
	html := string(playgroundHTML)

	if strings.Contains(html, "\nprobeLocalEngine();") {
		t.Fatal("hosted studio must not probe loopback during page startup")
	}
	for _, marker := range []string{
		"if (isHostedStudio() && !hasExplicitGatewayEndpoint())",
		"const useLiveGateway = !isHostedStudio() || hasExplicitGatewayEndpoint();",
		"Hosted pages must not probe loopback during startup",
	} {
		if !strings.Contains(html, marker) {
			t.Fatalf("playground is missing hosted-studio connection guard %q", marker)
		}
	}

	modalStart := strings.Index(html, "function openGatewayModal()")
	modalEnd := strings.Index(html, "function closeGatewayModal()")
	if modalStart < 0 || modalEnd <= modalStart {
		t.Fatal("gateway modal functions are missing")
	}
	if strings.Contains(html[modalStart:modalEnd], "pingGatewayManual()") {
		t.Fatal("opening the gateway modal must not trigger a local-network probe")
	}
}

func TestGenerationCleanupReferencesAreVisibleFromFinally(t *testing.T) {
	html := string(playgroundHTML)
	functionStart := strings.Index(html, "async function sendMessage()")
	if functionStart < 0 {
		t.Fatal("sendMessage function start is missing")
	}
	functionEndOffset := strings.Index(html[functionStart:], "\n}\n\n// Progressive Web App (PWA) Offline Service Worker Registration")
	functionEnd := functionStart + functionEndOffset
	if functionEndOffset < 0 || functionEnd <= functionStart {
		t.Fatal("sendMessage function boundaries are missing")
	}
	source := html[functionStart:functionEnd]
	tryStart := strings.Index(source, "\n  try {")
	finallyStart := strings.Index(source, "\n  } finally {")
	if tryStart < 0 || finallyStart <= tryStart {
		t.Fatal("sendMessage try/finally lifecycle is missing")
	}

	for _, declaration := range []string{
		`const sendBtn = document.getElementById("send-btn");`,
		`const sendLabel = document.getElementById("send-btn-label");`,
		`const sendIcon = document.getElementById("send-btn-icon");`,
	} {
		declarationIndex := strings.Index(source, declaration)
		if declarationIndex < 0 {
			t.Fatalf("sendMessage is missing cleanup reference declaration %q", declaration)
		}
		if declarationIndex > tryStart {
			t.Fatalf("%q is declared inside try and is unavailable to finally", declaration)
		}
	}

	for _, cleanup := range []string{
		`sendBtn.classList.remove("stop-mode")`,
		`sendLabel.innerText = dict.btnSend`,
		`sendIcon.innerText = "➤"`,
		`inputEl.disabled = false`,
	} {
		if !strings.Contains(source[finallyStart:], cleanup) {
			t.Fatalf("sendMessage finally block is missing cleanup %q", cleanup)
		}
	}

	for _, lifecycleMarker := range []string{
		`partial += decoder.decode();`,
		`streamEnded = true;`,
		`streamReadError = err;`,
		`let timer = 0;`,
		`if (timer) clearTimeout(timer);`,
		`throw new Error("Stream ended before completion")`,
		`if (streamReadError.name === 'AbortError')`,
	} {
		if !strings.Contains(source, lifecycleMarker) {
			t.Fatalf("sendMessage is missing stream lifecycle marker %q", lifecycleMarker)
		}
	}
}

func TestChatScrollKeepsAStableBottomAnchor(t *testing.T) {
	html := string(playgroundHTML)
	for _, marker := range []string{
		`const CHAT_BOTTOM_THRESHOLD = 120;`,
		`function isMessagesNearBottom()`,
		`function keepMessagesAtBottom()`,
		`msgListEl.scrollTo({ top: bottom, behavior: "auto" });`,
		`min-height: 0;`,
		`overflow-anchor: none;`,
		`var(--composer-offset, 140px)`,
		`var(--composer-offset, calc(135px + env(safe-area-inset-bottom)))`,
		`function syncComposerOffset()`,
		`new ResizeObserver(() => {`,
		`if (userIsAtBottom) keepMessagesAtBottom();`,
		`aria-label="Jump to bottom"`,
	} {
		if !strings.Contains(html, marker) {
			t.Fatalf("playground is missing stable chat scroll marker %q", marker)
		}
	}
	if strings.Contains(html, `storyDiv.scrollIntoView({ behavior: "smooth", block: "start" })`) {
		t.Fatal("new responses must not scroll the nested chat viewport to the story-card top")
	}
}

func TestPlaygroundLanguageCoverageAndScriptIsolation(t *testing.T) {
	html := string(playgroundHTML)
	for _, marker := range []string{
		`function applyLanguageUI(dict)`,
		`document.documentElement.lang = currentLang === "hi" ? "hi" : "en"`,
		`id="nav-config-label"`,
		`id="label-target-model"`,
		`id="gateway-modal-title"`,
		`const ui = dict.ui || I18N.en.ui;`,
		`ui.starterCards.snake[0]`,
		"translitCache.get(`${currentIndicLang}:${low}`)",
		`currentIndicLang.startsWith('hi-') ? OFFLINE_INDIC_RULES.special[low] : null`,
	} {
		if !strings.Contains(html, marker) {
			t.Fatalf("playground is missing language-safety marker %q", marker)
		}
	}
	if strings.Contains(html, `translitCache.get(low) || OFFLINE_INDIC_RULES.special[low]`) {
		t.Fatal("sendMessage still uses a language-agnostic transliteration cache key")
	}
}

func TestNativeExternalLinksUseTheDefaultBrowserBridge(t *testing.T) {
	html := string(playgroundHTML)
	for _, marker := range []string{
		`function openExternalURL(rawURL)`,
		`window.runtime.BrowserOpenURL(parsedURL.href)`,
		`target.target === "_blank" || linkURL.origin !== window.location.origin`,
		`openExternalURL('https://github.com/div197/BOB-Gemini-Free/releases')`,
	} {
		if !strings.Contains(html, marker) {
			t.Fatalf("playground is missing native external-link marker %q", marker)
		}
	}
	if strings.Contains(html, `onclick="window.open('https://github.com/div197/BOB-Gemini-Free/releases'`) {
		t.Fatal("release button still opens GitHub inside the native WebView")
	}
}

func TestResponsiveDrawersDoNotCoverNewChatToolbar(t *testing.T) {
	html := string(playgroundHTML)
	for _, marker := range []string{
		`--studio-toolbar-height: 42px;`,
		`class="sub-bar"`,
		`title="Start a new chat canvas"`,
		`top: var(--studio-toolbar-height);`,
		`bottom: auto;`,
		`height: calc(100% - var(--studio-toolbar-height));`,
	} {
		if !strings.Contains(html, marker) {
			t.Fatalf("responsive New Chat layout is missing marker %q", marker)
		}
	}
	if strings.Contains(html, "@media (max-width: 1140px) {\n  .telemetry-bar") &&
		strings.Contains(html, "position: absolute;\n    top: 0;\n    bottom: 0;\n    height: 100%;") {
		t.Fatal("responsive drawers still cover the full toolbar height")
	}
}

func TestPlaygroundBoundsManualRetriesAndLocksRequestControls(t *testing.T) {
	html := string(playgroundHTML)
	for _, marker := range []string{
		`const MAX_MANUAL_RETRIES = 2;`,
		`const MANUAL_RETRY_COOLDOWN_MS = 4000;`,
		`manualRetriesSinceSuccess >= MAX_MANUAL_RETRIES`,
		`Retry this response with the current model`,
		`function handleGenerationSafeModelChange(selectEl)`,
		`Model changes are locked while a response is streaming`,
		`setGenerationControlsDisabled(true);`,
		`setGenerationControlsDisabled(false);`,
		`let streamProtocolError = null;`,
		`if (finishReason === "error")`,
		`Cookie pools do not bypass quotas or provider policy.`,
	} {
		if !strings.Contains(html, marker) {
			t.Fatalf("playground is missing provider-safety marker %q", marker)
		}
	}
}
