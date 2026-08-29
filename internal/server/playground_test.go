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
		`status: "complete"`,
		`status: "stopped"`,
		`status: "error"`,
		`const safeErrorMessage = String((err && err.message) || "Generation failed")`,
		`if (m.role === "assistant" && m.status && m.status !== "complete") continue;`,
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
		`const CHAT_TOP_THRESHOLD = 24;`,
		`function isMessagesNearBottom()`,
		`function isMessagesNearTop()`,
		`function updateScrollButtons()`,
		`const canNavigate = hasScrollableMessages;`,
		`function scheduleScrollButtonUpdate()`,
		`requestAnimationFrame(updateScrollButtons)`,
		`function keepMessagesAtBottom()`,
		`msgListEl.scrollTo({ top: bottom, behavior: "auto" });`,
		`function scrollToTop()`,
		`msgListEl.scrollTo({ top: 0, behavior: "smooth" });`,
		`min-height: 0;`,
		`overflow-anchor: none;`,
		`var(--composer-offset, 140px)`,
		`var(--composer-offset, calc(135px + env(safe-area-inset-bottom)))`,
		`function syncComposerOffset()`,
		`new ResizeObserver(() => {`,
		`if (userIsAtBottom) keepMessagesAtBottom();`,
		`aria-label="Jump to top"`,
		`aria-label="Jump to bottom"`,
	} {
		if !strings.Contains(html, marker) {
			t.Fatalf("playground is missing stable chat scroll marker %q", marker)
		}
	}
	if strings.Contains(html, `storyDiv.scrollIntoView({ behavior: "smooth", block: "start" })`) {
		t.Fatal("new responses must not scroll the nested chat viewport to the story-card top")
	}
	if strings.Contains(html, `else if (/Gemini Developer API|AI Studio|quota|rate limit/i.test(err.message || ""))`) {
		t.Fatal("provider-error rendering must not be nested inside the stream render scheduler")
	}
	if !strings.Contains(html, `fallbackTimer = setTimeout(() => requestAnimationFrame(doRender), 66 - elapsed);`) {
		t.Fatal("stream render scheduler lost its bounded fallback timer")
	}
}

func TestArtifactEditorPreservesGeneratedSource(t *testing.T) {
	html := string(playgroundHTML)
	for _, marker := range []string{
		`id="art-code-editor" aria-label="Artifact source code editor"`,
		`function getArtifactSource(artifact)`,
		`function normalizeArtifactToken(token, fallbackLanguage)`,
		`token.text ?? token.code ?? token.raw ?? token.value`,
		`function prepareArtifactEditor(artifact)`,
		`function persistArtifactEditorDraft()`,
		`if (typeof artifact.originalCode !== 'string') artifact.originalCode = source;`,
		`if (typeof artifact.editorCode !== 'string') artifact.editorCode = source;`,
		`const needsEmptyEditorRecovery = !editor.value && !!artifact.editorCode;`,
		`editor.value = artifact.editorCode;`,
		`currentActiveArtifact.editorCode = editor.value;`,
		`function runArtifactEditorCode()`,
		`function resetArtifactEditorCode()`,
		`function copyArtifactEditorCode()`,
		`function toggleArtifactEditorWrap()`,
		`Cannot run an empty artifact source.`,
		`prepareArtifactEditor(currentActiveArtifact);`,
		`switchArtifactTab(art.isInteractive ? 'preview' : 'code');`,
	} {
		if !strings.Contains(html, marker) {
			t.Fatalf("playground is missing artifact editor marker %q", marker)
		}
	}
}

func TestHistoryStorageBoundsAttachmentsWithoutCorruptingPayloads(t *testing.T) {
	html := string(playgroundHTML)
	for _, marker := range []string{
		`const MAX_LOCAL_ATTACHMENT_BYTES = 32 * 1024 * 1024;`,
		`function validateLocalAttachment(file)`,
		`Attachment is too large. Maximum size is`,
		`const MAX_HISTORY_IMAGE_DATA_URL_CHARS = 750000;`,
		`const MAX_HISTORY_CONTENT_CHARS = 300000;`,
		`const MAX_HISTORY_MESSAGES = 200;`,
		`const MAX_HISTORY_TOTAL_CHARS = 4000000;`,
		`function clipHistoryText(value, limit)`,
		`function trimHistoryPairs(history, maxMessages = MAX_HISTORY_MESSAGES)`,
		`function trimStoredHistory(history)`,
		`chatHistory = trimStoredHistory(trimHistoryPairs(parsed`,
		`function sanitizeAttachmentForStorage(attachment, includeImageData = true)`,
		`const MAX_HISTORY_SERIALIZED_CHARS = MAX_HISTORY_TOTAL_CHARS + 500000;`,
		`Stored chat history exceeded the safe size bound`,
		`if (content.truncated) clean.contentTruncated = true;`,
		`dataOmitted = true`,
		`attachedFiles: Array.isArray(m.attachedFiles)`,
	} {
		if !strings.Contains(html, marker) {
			t.Fatalf("playground is missing bounded attachment-storage marker %q", marker)
		}
	}
	if strings.Contains(html, `dataUrl: item.attachedImage.dataUrl.slice(0, 100000)`) {
		t.Fatal("history storage must not persist a truncated, invalid image data URL")
	}
}

func TestArtifactSandboxKeepsOpaqueIsolationAndReportsRuntimeFailures(t *testing.T) {
	html := string(playgroundHTML)
	for _, marker := range []string{
		`sandbox="allow-scripts allow-modals allow-forms"`,
		`function artifactRuntimeBootstrap()`,
		`function createMemoryStorage()`,
		`installStorage('localStorage')`,
		`installStorage('sessionStorage')`,
		`type: 'BOB_ARTIFACT_RUNTIME_ERROR'`,
		`window.addEventListener('unhandledrejection'`,
		`function prepareHTMLArtifactSource(source)`,
		`const htmlPattern = /<html\b[^>]*>/i;`,
		`id="art-runtime-error" role="alert" hidden`,
		`function showArtifactRuntimeError(message)`,
		`e.source === iframe.contentWindow`,
	} {
		if !strings.Contains(html, marker) {
			t.Fatalf("playground is missing artifact sandbox safety marker %q", marker)
		}
	}
	if strings.Contains(html, `sandbox="allow-scripts allow-same-origin`) {
		t.Fatal("artifact sandbox must not combine scripts with same-origin access")
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
		`if (data && data.error && data.error.message)`,
		`if (finishReason === "error")`,
		`Cookie pools do not bypass quotas or provider policy.`,
	} {
		if !strings.Contains(html, marker) {
			t.Fatalf("playground is missing provider-safety marker %q", marker)
		}
	}
}

func TestPlaygroundEducatesAboutExplicitGeminiDeveloperAPIRoute(t *testing.T) {
	html := string(playgroundHTML)
	for _, marker := range []string{
		`Gemini Developer API (optional)`,
		`https://aistudio.google.com/app/apikey`,
		`https://ai.google.dev/gemini-api/docs/rate-limits`,
		`https://ai.google.dev/gemini-api/docs/models`,
		`Use for this session`,
		`is never saved by BOB`,
		`Google project quotas, free-tier limits, and any billing settings still apply`,
		`BOB does not rotate keys or automatically retry a request through a second provider`,
		`function canSendGeminiProviderKey()`,
		`configure a local or explicitly trusted gateway endpoint first`,
		`descEn: "Translating selected OpenAI/Anthropic HTTP REST endpoints`,
	} {
		if !strings.Contains(html, marker) {
			t.Fatalf("playground is missing Developer API education marker %q", marker)
		}
	}
	if strings.Contains(html, "enabling 100% free local access") || strings.Contains(html, "100% मुफ़्त स्थानीय उपयोग संभव") {
		t.Fatal("playground still presents provider access as universally free")
	}
}
