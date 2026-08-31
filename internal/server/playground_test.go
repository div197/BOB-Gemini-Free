package server

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
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

func TestGatewayAuthKeyIsSessionOnly(t *testing.T) {
	html := string(playgroundHTML)
	for _, marker := range []string{
		`let gatewayApiKey = "";`,
		`const LEGACY_GATEWAY_KEY_STORAGE_KEY = "bob_api_key";`,
		`localStorage.removeItem(LEGACY_GATEWAY_KEY_STORAGE_KEY);`,
		`keyInput.value = gatewayApiKey;`,
		`gatewayApiKey = clean;`,
		`const apiKey = gatewayApiKey;`,
		`gatewayApiKey = "";`,
	} {
		if !strings.Contains(html, marker) {
			t.Fatalf("playground is missing session-only gateway-auth marker %q", marker)
		}
	}
	for _, forbidden := range []string{
		`localStorage.setItem('bob_api_key'`,
		`localStorage.getItem('bob_api_key'`,
		`localStorage.setItem("bob_api_key"`,
		`localStorage.getItem("bob_api_key"`,
	} {
		if strings.Contains(html, forbidden) {
			t.Fatalf("gateway auth key is persisted or restored from localStorage: %q", forbidden)
		}
	}
}

func TestGatewayAuthInputDoesNotProbeUntilExplicitPing(t *testing.T) {
	html := string(playgroundHTML)
	start := strings.Index(html, "function saveApiKeyFromInput(val)")
	if start < 0 {
		t.Fatal("gateway access-key input handler boundaries are missing")
	}
	endOffset := strings.Index(html[start:], "\n}\n\nfunction clearGatewayApiKey")
	if endOffset < 0 {
		t.Fatal("gateway access-key input handler end is missing")
	}
	source := html[start : start+endOffset]
	if !strings.Contains(source, "gatewayApiKey = clean;") || !strings.Contains(source, "updateGatewayCredentialStatus();") {
		t.Fatal("gateway access-key input must update only session state and route status")
	}
	for _, forbidden := range []string{"refreshTelemetry();", "syncBackendModels();", "fetch("} {
		if strings.Contains(source, forbidden) {
			t.Fatalf("gateway access-key input must not probe the network while typing: %q", forbidden)
		}
	}
	for _, marker := range []string{
		"After entering a key, choose Test Ping to verify the endpoint.",
		"Key दर्ज करने के बाद endpoint जाँचने के लिए टेस्ट पिंग चुनें।",
		"function pingGatewayManual()",
	} {
		if !strings.Contains(html, marker) {
			t.Fatalf("gateway access-key flow is missing explicit verification marker %q", marker)
		}
	}
}

func TestTelemetryUsesHealthzBeforeProtectedStats(t *testing.T) {
	html := string(playgroundHTML)
	start := strings.Index(html, "async function refreshTelemetry()")
	if start < 0 {
		t.Fatal("telemetry function start is missing")
	}
	endOffset := strings.Index(html[start:], "\n}\n\nfunction showLocalOnboardingModal")
	if endOffset < 0 {
		t.Fatal("telemetry function boundaries are missing")
	}
	source := html[start : start+endOffset]
	health := strings.Index(source, `fetch(baseUrl + "/healthz"`)
	stats := strings.Index(source, `fetch(statsUrl`)
	if health < 0 || stats < 0 || health >= stats {
		t.Fatal("telemetry must establish health/auth state before protected stats")
	}
	for _, marker := range []string{
		`healthRes.headers.get("X-BOB-Auth-Required") === "true"`,
		`if (healthAuthRequired && !apiKey)`,
		`clearTelemetryStats();`,
		`updateGatewayStatusIndicator(true, Math.round(performance.now() - start));`,
	} {
		if !strings.Contains(source, marker) {
			t.Fatalf("telemetry auth boundary is missing marker %q", marker)
		}
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
		`throw new Error("Stream completed without usable model output")`,
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

func TestStudioSSEParserAcceptsStandardDataFieldForms(t *testing.T) {
	html := string(playgroundHTML)
	start := strings.Index(html, "function processSSELines(buffer)")
	if start < 0 {
		t.Fatal("processSSELines function start is missing")
	}
	end := strings.Index(html[start:], "\n    while (true) {")
	if end <= 0 {
		t.Fatal("processSSELines function boundaries are missing")
	}
	source := html[start : start+end]
	for _, marker := range []string{
		`function consumeSSEEvent()`,
		`sseEventData.join("\n")`,
		`const streamData = sseEventData.join("\n");`,
		`const MAX_SSE_DIAGNOSTIC_COUNT = 100;`,
		`Math.min(sseUnknownEventCount + 1, MAX_SSE_DIAGNOSTIC_COUNT)`,
		`if (streamData === "[DONE]") { streamDone = true; return true; }`,
		`const data = JSON.parse(streamData);`,
	} {
		if !strings.Contains(html, marker) {
			t.Fatalf("Studio SSE parser is missing standard data-field marker %q", marker)
		}
	}
	for _, marker := range []string{
		`const separator = line.indexOf(":");`,
		`case "data":`,
		`case "event":`,
		`Math.min(sseUnknownFieldCount + 1, MAX_SSE_DIAGNOSTIC_COUNT)`,
	} {
		if !strings.Contains(source, marker) {
			t.Fatalf("Studio SSE parser is missing standard data-field marker %q", marker)
		}
	}
	for _, forbidden := range []string{
		`const trimmed = line.trim();`,
		`trimmed.startsWith("data:")`,
		`JSON.parse(fieldValue)`,
	} {
		if strings.Contains(source, forbidden) {
			t.Fatalf("Studio SSE parser still requires non-standard spacing: %q", forbidden)
		}
	}
}

func TestStudioSSEParserFlushesAnUnterminatedFinalEvent(t *testing.T) {
	html := string(playgroundHTML)
	start := strings.Index(html, "// Flush any remaining partial buffer")
	if start < 0 {
		t.Fatal("Studio stream flush boundary is missing")
	}
	source := html[start:]
	if !strings.Contains(source, `processSSELines(partial + "\n\n");`) {
		t.Fatal("Studio stream flush must terminate the final SSE event")
	}
}

func TestReachableGatewayIsNotShownOfflineAfterHTTPOrStreamFailure(t *testing.T) {
	html := string(playgroundHTML)
	functionStart := strings.Index(html, "async function sendMessage()")
	if functionStart < 0 {
		t.Fatal("sendMessage function start is missing")
	}
	functionEndOffset := strings.Index(html[functionStart:], "\n}\n\n// Progressive Web App (PWA) Offline Service Worker Registration")
	if functionEndOffset < 0 {
		t.Fatal("sendMessage function boundary is missing")
	}
	source := html[functionStart : functionStart+functionEndOffset]
	for _, marker := range []string{
		`let gatewayResponseReceived = false;`,
		`const res = await fetch(baseUrl + "/v1/chat/completions",`,
		`gatewayResponseReceived = true;`,
		`if (!gatewayResponseReceived) updateGatewayStatusIndicator(false, null);`,
	} {
		if !strings.Contains(source, marker) {
			t.Fatalf("generation status lifecycle is missing marker %q", marker)
		}
	}
	if strings.Contains(source, "\n      updateGatewayStatusIndicator(false, null);\n      let mixedContentHint") {
		t.Fatal("generation catch still marks a reachable gateway offline for provider or HTTP errors")
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

func TestResponsiveHeaderCompactsTabletControls(t *testing.T) {
	html := string(playgroundHTML)
	for _, marker := range []string{
		`@media (max-width: 860px) {`,
		`min-width: 0;`,
		`flex-shrink: 1;`,
		`.header-controls .nav-btn-text,`,
		`#cmd-menu-label {`,
		`display: none !important;`,
		`#theme-selector {`,
		`.header-controls .lang-selector {`,
		`min-width: 82px;`,
		`.header-controls .github-pill {`,
		`min-width: 34px;`,
		`width: 46px;`,
		`flex-basis: 46px;`,
	} {
		if !strings.Contains(html, marker) {
			t.Fatalf("playground is missing tablet header compaction marker %q", marker)
		}
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

func TestArtifactRenderingHasBoundedSourceAndRegistryState(t *testing.T) {
	html := string(playgroundHTML)
	for _, marker := range []string{
		`const MAX_ARTIFACT_SOURCE_CHARS = 2 * 1024 * 1024;`,
		`const MAX_ARTIFACT_REGISTRY_ITEMS = 128;`,
		`const MAX_ARTIFACT_REGISTRY_CHARS = 8 * 1024 * 1024;`,
		`function ensureArtifactRegistryCapacity(id, nextSourceLength)`,
		`function registerArtifact(artifact)`,
		`function artifactSizeLimitNotice(language, sourceLength)`,
		`if (rawCode.length > MAX_ARTIFACT_SOURCE_CHARS)`,
		`interactive previews are limited to`,
		`activeArtifactRender = { key: artifactRenderScopeKey(scopeKey), ordinal: 0 };`,
		`function nextArtifactID()`,
		`window.__bobMermaidFailure`,
		`Mermaid preview library could not be loaded`,
		`window.__bobPyodideLoadError`,
		`Pyodide WebAssembly library could not be loaded`,
		"renderMarkdown(asstFullText, `live-${msgIdx}-content`)",
	} {
		if !strings.Contains(html, marker) {
			t.Fatalf("playground is missing bounded artifact marker %q", marker)
		}
	}
	if strings.Contains(html, `artifactsRegistry.set(artId,`) {
		t.Fatal("artifact renderer still writes directly to an unbounded registry")
	}
	if strings.Contains(html, `const artId = 'art_' + Math.random()`) {
		t.Fatal("artifact IDs must be stable across repeated streaming renders")
	}
}

func TestArtifactLaunchChipUsesOneKeyboardAction(t *testing.T) {
	html := string(playgroundHTML)
	for _, marker := range []string{
		`<div class="artifact-card-chip" role="group"`,
		`<button type="button" class="btn-launch-art"`,
	} {
		if !strings.Contains(html, marker) {
			t.Fatalf("artifact launch chip is missing accessible-action marker %q", marker)
		}
	}
	if strings.Contains(html, `<div class="artifact-card-chip" data-action="launch-art"`) {
		t.Fatal("artifact launch chip still exposes a duplicate click-only container action")
	}
	if strings.Contains(html, `<button class="btn-launch-art"`) {
		t.Fatal("artifact launch button must declare an explicit button type")
	}
}

func TestClipboardActionsHaveExplicitFailureState(t *testing.T) {
	html := string(playgroundHTML)

	if strings.Count(html, "navigator.clipboard.writeText(") != 1 {
		t.Fatalf("clipboard writes must be centralized in one guarded helper, got %d call sites", strings.Count(html, "navigator.clipboard.writeText("))
	}
	for _, marker := range []string{
		"async function copyTextToClipboard(text, successMessage, failureMessage)",
		"Clipboard access is unavailable.",
		"Could not copy to clipboard.",
		"return false;",
		"return true;",
		"return copyTextToClipboard(cmd, \"Command copied to clipboard!\")",
		"return copyTextToClipboard(element.innerText, \"Code snippet copied to clipboard!\")",
		"await copyTextToClipboard(source, 'Editor source copied to clipboard!', 'Could not copy editor source.')",
		"return copyTextToClipboard(getArtifactSource(currentActiveArtifact), \"Artifact code copied to clipboard!\")",
		"await copyTextToClipboard(code, 'Code block copied to clipboard!')",
		"await copyTextToClipboard(text, \"Message copied to clipboard!\")",
		"return copyTextToClipboard(transcript, \"Complete conversation copied to clipboard!\")",
	} {
		if !strings.Contains(html, marker) {
			t.Fatalf("clipboard failure-safety marker %q is missing", marker)
		}
	}
}

func TestDeepRefinerDoesNotSilentlyReplacePromptOnFailure(t *testing.T) {
	html := string(playgroundHTML)
	start := strings.Index(html, "async function refineCurrentPromptDeep()")
	if start < 0 {
		t.Fatal("deep refiner function is missing")
	}
	endOffset := strings.Index(html[start:], "\n}\n\nfunction printConversation()")
	if endOffset < 0 {
		t.Fatal("deep refiner function boundary is missing")
	}
	source := html[start : start+endOffset]
	for _, marker := range []string{
		"3-Stage Deep Invariant Refinement",
		`original prompt kept`,
		`const data = await res.json().catch(() => null);`,
		`finally {`,
		`if (timeoutId) clearTimeout(timeoutId);`,
	} {
		if !strings.Contains(source, marker) {
			t.Fatalf("deep refiner is missing failure-safety marker %q", marker)
		}
	}
	if strings.Contains(source, "enhanceCurrentPrompt()") || strings.Contains(source, "Deep refiner fallback") {
		t.Fatal("deep refiner must not silently substitute a local prompt enhancement after a server failure")
	}
}

func TestPromptWandDoesNotUseImplicitOfflineFallback(t *testing.T) {
	html := string(playgroundHTML)
	start := strings.Index(html, "async function enhanceCurrentPrompt()")
	if start < 0 {
		t.Fatal("prompt wand function is missing")
	}
	endOffset := strings.Index(html[start:], "\n}\n\nasync function refineCurrentPromptDeep()")
	if endOffset < 0 {
		t.Fatal("prompt wand function boundary is missing")
	}
	source := html[start : start+endOffset]
	for _, marker := range []string{
		"original prompt kept",
		"const data = await res.json().catch(() => null);",
		"finally {",
		"if (timeoutId) clearTimeout(timeoutId);",
	} {
		if !strings.Contains(source, marker) {
			t.Fatalf("prompt wand is missing failure-safety marker %q", marker)
		}
	}
	for _, forbidden := range []string{
		"Intelligent Offline Fallback Engine",
		"AI prompt enhancement offline fallback",
		"let enhanced = \"\"",
		"inputEl.value = enhanced",
	} {
		if strings.Contains(source, forbidden) {
			t.Fatalf("prompt wand still contains implicit fallback %q", forbidden)
		}
	}
}

func TestImageProviderFailureDoesNotSilentlyReplayViaOCR(t *testing.T) {
	html := string(playgroundHTML)
	start := strings.Index(html, "// Do not silently convert an image request")
	if start < 0 {
		t.Fatal("image failure policy marker is missing")
	}
	endOffset := strings.Index(html[start:], "\n      let mixedContentHint = \"\";")
	if endOffset < 0 {
		t.Fatal("image failure policy boundary is missing")
	}
	source := html[start : start+endOffset]
	for _, forbidden := range []string{
		"Auto-Healing",
		"Auto-recovering via Client-Side Local OCR",
		"retryWithModel('default', chatHistory.length - 1)",
		"Extracted via Client-Side Local OCR",
	} {
		if strings.Contains(source, forbidden) {
			t.Fatalf("image provider failure still contains silent OCR replay %q", forbidden)
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

func TestHistoryStorageFailuresAreVisibleAndRecoverable(t *testing.T) {
	html := string(playgroundHTML)
	for _, marker := range []string{
		`id="history-storage-status" class="history-storage-status" role="status" aria-live="polite" aria-atomic="true" hidden`,
		`function setHistoryStorageState(state)`,
		`function tryWriteHistoryStorage(serialized)`,
		`function tryRemoveHistoryStorage()`,
		`Conversation saved in compact mode; some attachment previews may be omitted.`,
		`Previous local conversation data was unavailable, so a fresh chat was started.`,
		`This conversation could not be saved on this device.`,
		`if (tryWriteHistoryStorage(JSON.stringify(compacted)))`,
		`setHistoryStorageState("compacted")`,
		`setHistoryStorageState("unsaved")`,
		`setHistoryStorageState("recovered")`,
		`if (tryRemoveHistoryStorage()) {`,
		`setHistoryStorageState("clear");`,
		`setHistoryStorageState("unavailable");`,
	} {
		if !strings.Contains(html, marker) {
			t.Fatalf("playground is missing visible history-storage recovery marker %q", marker)
		}
	}
	if strings.Contains(html, `localStorage.setItem(STORAGE_KEY, JSON.stringify(cleanHistory))`) ||
		strings.Contains(html, `localStorage.setItem(STORAGE_KEY, JSON.stringify(compacted))`) {
		t.Fatal("history writes must go through the guarded storage helper")
	}
}

func TestPreferencesFailClosedWhenBrowserStorageIsUnavailable(t *testing.T) {
	html := string(playgroundHTML)
	for _, marker := range []string{
		`function getLocalPreference(key, fallback = null)`,
		`function setLocalPreference(key, value)`,
		`function removeLocalPreference(key)`,
		`return fallback;`,
		`return false;`,
		`setLocalPreference('bob_preferred_tts_voice', this.value)`,
		`setLocalPreference('bob_tts_rate', this.value)`,
		`getLocalPreference(ENDPOINT_KEY, '')`,
		`getLocalPreference(PANEL_LEFT_KEY, "")`,
		`setLocalPreference(THEME_KEY, themeName)`,
		`getLocalPreference(CUSTOM_THEME_KEY, null)`,
		`setLocalPreference('bob_custom_instructions', JSON.stringify({ persona, style }))`,
		`getLocalPreference('bob_reading_zoom', '1.0')`,
	} {
		if !strings.Contains(html, marker) {
			t.Fatalf("playground is missing fail-closed preference marker %q", marker)
		}
	}
	for _, forbidden := range []string{
		`localStorage.setItem('bob_preferred_tts_voice'`,
		`localStorage.setItem('bob_tts_rate'`,
		`localStorage.getItem("bob_gemini_indic_lang")`,
		`localStorage.getItem(LANG_KEY)`,
		`localStorage.setItem(LANG_KEY`,
		`localStorage.getItem(THEME_KEY)`,
		`localStorage.setItem(THEME_KEY`,
		`localStorage.getItem(CUSTOM_THEME_KEY)`,
		`localStorage.setItem(CUSTOM_THEME_KEY`,
		`localStorage.getItem('bob_custom_instructions')`,
		`localStorage.setItem('bob_custom_instructions'`,
		`localStorage.getItem('bob_preferred_tts_voice')`,
		`localStorage.getItem('bob_tts_rate')`,
		`localStorage.getItem('bob_reading_zoom')`,
		`localStorage.setItem('bob_reading_zoom'`,
	} {
		if strings.Contains(html, forbidden) {
			t.Fatalf("preference storage bypasses the guarded helper: %q", forbidden)
		}
	}
}

func TestAttachmentParsingIsBoundedAndCancellable(t *testing.T) {
	html := string(playgroundHTML)
	for _, marker := range []string{
		`const MAX_ATTACHMENT_EXTRACTED_CHARS = 1000000;`,
		`const MAX_ATTACHMENT_PDF_PAGES = 200;`,
		`const MAX_ATTACHMENT_PARSE_CONCURRENCY = 2;`,
		`function acquireAttachmentParseSlot(signal = null)`,
		`pendingAttachmentParseSlots.indexOf(grant)`,
		`reject(attachmentAbortError());`,
		`function isAttachmentEntryActive(entry)`,
		`releaseParseSlot = await acquireAttachmentParseSlot(fileEntry.abortController ? fileEntry.abortController.signal : null);`,
		`if (!isAttachmentEntryActive(fileEntry)) return;`,
		`entry.cancelled = true;`,
		`function readAttachmentBlob(blob, entry, method)`,
		`reader.abort()`,
		`abortController: typeof AbortController === 'function' ? new AbortController() : null`,
		`function runAttachmentOCR(dataUrl, entry)`,
		`if (typeof tesseract.createWorker === 'function')`,
		`worker.terminate`,
		`let pdfDestroyPromise = null;`,
		`function destroyActivePDF()`,
		`activePDFLoadingTask && typeof activePDFLoadingTask.destroy === 'function'`,
		`pdfDestroyPromise = Promise.resolve(target.destroy());`,
		`function abortPDFParse()`,
		`attachmentSignal.addEventListener('abort', abortPDFParse`,
		`if (attachmentSignal.aborted) abortPDFParse();`,
		`attachmentSignal.removeEventListener('abort', abortPDFParse);`,
		`await destroyActivePDF();`,
		`await readAttachmentAsArrayBuffer(file, fileEntry);`,
		`await readAttachmentAsText(file.slice(0, MAX_ATTACHMENT_EXTRACTED_CHARS), fileEntry);`,
		`fileEntry.extractionTruncated = true;`,
		`window.addEventListener('drop'`,
		`extractDocumentToMarkdown(dt.files[i]);`,
		`entry.abortController.abort();`,
		`file.abortController.abort();`,
	} {
		if !strings.Contains(html, marker) {
			t.Fatalf("playground is missing bounded attachment marker %q", marker)
		}
	}
	if strings.Contains(html, "function attachImageFile(file)") {
		t.Fatal("legacy attachment reader must not remain alongside the universal extractor")
	}
	if strings.Contains(html, "dropZone.addEventListener(\"drop\"") {
		t.Fatal("workspace drop handler must not duplicate the universal window drop handler")
	}
}

func TestAttachmentControlsDoNotEmbedUntrustedIDsInInlineJavaScript(t *testing.T) {
	html := string(playgroundHTML)
	for _, marker := range []string{
		`data-action="preview-attachment"`,
		`data-action="remove-attachment"`,
		`data-file-id="${escapeHtml(file.id)}"`,
		`action === 'preview-attachment'`,
		`action === 'remove-attachment'`,
		`openDocPreviewModal(fileId)`,
		`removeAttachedFile(fileId)`,
	} {
		if !strings.Contains(html, marker) {
			t.Fatalf("attachment controls are missing safe delegated-action marker %q", marker)
		}
	}
	for _, forbidden := range []string{
		`onclick="openDocPreviewModal('${file.id}')"`,
		`onclick="removeAttachedFile('${file.id}')"`,
	} {
		if strings.Contains(html, forbidden) {
			t.Fatalf("attachment control still embeds an untrusted ID in inline JavaScript %q", forbidden)
		}
	}
}

func TestPlaygroundHasOneCodeCopyHandler(t *testing.T) {
	html := string(playgroundHTML)
	if got := strings.Count(html, "function copyCode("); got != 1 {
		t.Fatalf("playground has %d copyCode declarations, want exactly one", got)
	}
}

func TestPersistedAttachmentIconsAreEscapedBeforeHistoryHTML(t *testing.T) {
	html := string(playgroundHTML)
	marker := `${escapeHtml(f.icon || '📄')}`
	if strings.Count(html, marker) < 2 {
		t.Fatalf("history attachment icon escaping marker %q must protect both rendered attachment paths", marker)
	}
	for _, forbidden := range []string{
		`<span>${f.icon || '📄'}</span>`,
		`<span>${f.icon || '📄'} </span>`,
	} {
		if strings.Contains(html, forbidden) {
			t.Fatalf("persisted attachment icon is still inserted into HTML without escaping: %q", forbidden)
		}
	}
}

func TestAttachmentImagePreviewsUseAccessibleRasterOnlyControls(t *testing.T) {
	html := string(playgroundHTML)
	for _, marker := range []string{
		`const SAFE_ATTACHMENT_IMAGE_MIME_TYPES = new Set([`,
		`function isSafeAttachmentImageDataURL(value`,
		`data-action="preview-attachment-image"`,
		`function openAttachmentImagePreview(button)`,
		`openAttachmentImagePreview(target)`,
		`window.open(src, '_blank', 'noopener,noreferrer')`,
	} {
		if !strings.Contains(html, marker) {
			t.Fatalf("attachment image preview is missing safe accessible marker %q", marker)
		}
	}
	for _, forbidden := range []string{
		`onclick="window.open(this.src)"`,
		`<img src="${escapeHtml(f.dataUrl)}" class="user-attached-thumb"`,
		`<img src="${escapeHtml(attachedImgCopy.dataUrl)}" class="user-attached-thumb"`,
	} {
		if strings.Contains(html, forbidden) {
			t.Fatalf("attachment image preview still uses unsafe or inaccessible construction %q", forbidden)
		}
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

func TestArtifactPreviewOpensBeforeIframeHydration(t *testing.T) {
	html := string(playgroundHTML)
	start := strings.Index(html, "function openArtifactCanvas(artId)")
	if start < 0 {
		t.Fatal("openArtifactCanvas function is missing")
	}
	endOffset := strings.Index(html[start:], "\n}\n\nfunction closeArtifactCanvas()")
	if endOffset < 0 {
		t.Fatal("openArtifactCanvas function boundary is missing")
	}
	source := html[start : start+endOffset]
	openIndex := strings.Index(source, `if (modal) openManagedModal(modal, "#btn-art-back");`)
	iframeIndex := strings.Index(source, `const iframe = document.getElementById("art-sandbox-iframe");`)
	if openIndex < 0 || iframeIndex < 0 {
		t.Fatal("artifact preview modal and iframe markers are missing")
	}
	if openIndex > iframeIndex {
		t.Fatal("artifact modal must be visible before assigning iframe srcdoc")
	}
}

func TestArtifactPopoutPreservesSandboxAndHandlesBrowserFailures(t *testing.T) {
	html := string(playgroundHTML)
	start := strings.Index(html, "function popOutArtifact()")
	if start < 0 {
		t.Fatal("popOutArtifact function is missing")
	}
	endOffset := strings.Index(html[start:], "\n}\n\n// Initialize Universal Marked.js Configuration")
	if endOffset < 0 {
		t.Fatal("popOutArtifact function boundary is missing")
	}
	source := html[start : start+endOffset]
	for _, marker := range []string{
		`iframe sandbox="allow-scripts allow-modals allow-forms"`,
		`const frameSource = JSON.stringify(content)`,
		`.replace(/</g, "\\u003c")`,
		`window.open(url, '_blank', 'noopener,noreferrer')`,
		`if (!opened)`,
		`The artifact window was blocked by the browser`,
		`setTimeout(() => URL.revokeObjectURL(url), 60000);`,
	} {
		if !strings.Contains(source, marker) {
			t.Fatalf("pop-out is missing safety marker %q", marker)
		}
	}
	if strings.Contains(source, `window.open(url, '_blank');`) {
		t.Fatal("pop-out still opens generated content without noopener")
	}
}

func TestArtifactPopoutScriptHasNoEmbeddedClosingTag(t *testing.T) {
	html := string(playgroundHTML)
	marker := "const frameSource = JSON.stringify(content)"
	markerIndex := strings.Index(html, marker)
	if markerIndex < 0 {
		t.Fatal("artifact pop-out source marker is missing")
	}

	scriptStart := strings.Index(html[:markerIndex], "<script>\nlet chatHistory = [];")
	scriptEnd := strings.LastIndex(html, "</script>")
	if scriptStart < 0 || scriptEnd <= scriptStart {
		t.Fatal("artifact pop-out inline script boundaries are missing")
	}
	inlineSource := html[scriptStart+len("<script>\nlet chatHistory = [];") : scriptEnd]
	if strings.Contains(inlineSource, "</script>") {
		t.Fatal("artifact pop-out inline script contains an embedded raw closing tag")
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

func TestRootCDNDependenciesHaveIntegrityPins(t *testing.T) {
	html := string(playgroundHTML)
	headEnd := strings.Index(html, "<body")
	if headEnd < 0 {
		t.Fatal("playground is missing a document body boundary")
	}
	for _, line := range strings.Split(html[:headEnd], "\n") {
		line = strings.TrimSpace(line)
		isExternalScript := strings.HasPrefix(line, "<script ") && strings.Contains(line, `src="http`)
		isExternalStylesheet := strings.HasPrefix(line, "<link ") && strings.Contains(line, `href="http`)
		if !isExternalScript && !isExternalStylesheet {
			continue
		}
		if !strings.Contains(line, `integrity="sha384-`) || !strings.Contains(line, `crossorigin="anonymous"`) {
			t.Fatalf("root CDN dependency is missing SRI/cross-origin attributes: %s", line)
		}
	}
	if !strings.Contains(html, `tesseract.js@5.1.1/dist/tesseract.min.js`) {
		t.Fatal("Tesseract.js must remain pinned to the verified v5.1.1 asset")
	}
	if strings.Contains(html, `tesseract.js@5/dist/tesseract.min.js`) {
		t.Fatal("Tesseract.js must not use a floating major-version CDN URL")
	}
}

func TestDynamicArtifactCDNBootstrapsArePinned(t *testing.T) {
	html := string(playgroundHTML)
	for _, marker := range []string{
		`mermaid@10.9.0/dist/mermaid.min.js" integrity="sha384-6F4Ibv/ylL12O35KFWTeGTHuBKDz5L6yjKsgv3QHQ8s4NTqlDXq7kMlYXGs7MHFc" crossorigin="anonymous"`,
		`pyodide/v0.26.2/full/pyodide.js" integrity="sha384-tVslJOEkg7nVRW3Y3/ReGX0NnonNrbcmt1R5qFbQXQdGa2chRkoJYHAjAsv3zoTq" crossorigin="anonymous"`,
	} {
		if !strings.Contains(html, marker) {
			t.Fatalf("dynamic artifact CDN bootstrap is missing its pinned integrity contract: %q", marker)
		}
	}
	for _, floating := range []string{
		`cdn.jsdelivr.net/npm/mermaid/dist/mermaid.min.js`,
		`cdn.jsdelivr.net/pyodide/v0.26.2/full/pyodide.js" onerror`,
	} {
		if strings.Contains(html, floating) {
			t.Fatalf("dynamic artifact CDN bootstrap remains mutable or unpinned: %q", floating)
		}
	}
}

func TestErrorRecoveryConfigActionsAvoidJavaScriptURLs(t *testing.T) {
	html := string(playgroundHTML)
	for _, marker := range []string{
		`data-action="open-gateway-modal"`,
		`action === 'open-gateway-modal'`,
		`openGatewayModal();`,
		`class="inline-action-link"`,
	} {
		if !strings.Contains(html, marker) {
			t.Fatalf("error recovery config action is missing safe delegated marker %q", marker)
		}
	}
	if strings.Contains(html, `href="javascript:void(0)"`) {
		t.Fatal("error recovery UI must not use javascript: URLs as button actions")
	}
}

func TestMarkdownLinksUseStrictProtocolWhitelist(t *testing.T) {
	html := string(playgroundHTML)
	for _, marker := range []string{
		`const SAFE_EXTERNAL_LINK_PROTOCOLS = new Set(["http:", "https:", "mailto:", "tel:"]);`,
		`function sanitizeMarkdownHref(rawHref)`,
		`return SAFE_EXTERNAL_LINK_PROTOCOLS.has(parsedURL.protocol) ? href : "#";`,
		`rawHref = sanitizeMarkdownHref(rawHref);`,
		`SAFE_EXTERNAL_LINK_PROTOCOLS.has(parsedURL.protocol)`,
	} {
		if !strings.Contains(html, marker) {
			t.Fatalf("playground is missing strict Markdown-link protocol marker %q", marker)
		}
	}
	for _, forbidden := range []string{
		`trimmedHref.startsWith('javascript:')`,
		`trimmedHref.startsWith('vbscript:')`,
		`trimmedHref.startsWith('data:text/html')`,
	} {
		if strings.Contains(html, forbidden) {
			t.Fatalf("playground still relies on blacklist-only Markdown-link filtering %q", forbidden)
		}
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
		`streamProtocolError = "Stream contained an invalid SSE event"`,
		`if (data && data.error && data.error.message)`,
		`if (finishReason === "error")`,
		`const isGatewayAuthError = !useGeminiProvider && /bob gateway access key required|invalid api key|gateway requires an api key|api key protection enabled/i.test(safeErrorMessage);`,
		`throw new Error("401 Unauthorized: BOB Gateway Access Key required");`,
		`if (useGeminiProvider)`,
		`dict.providerAuthError`,
		`session authentication|Google session|HTTP 401|HTTP 403`,
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
		`Google Gemini Developer API (optional)`,
		`https://aistudio.google.com/app/apikey`,
		`https://ai.google.dev/gemini-api/docs/rate-limits`,
		`https://ai.google.dev/gemini-api/docs/models`,
		`Use your Google key for this session`,
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
	if strings.Contains(html, `if (safeErrorMessage.includes("401"))`) {
		t.Fatal("provider HTTP 401 must not be mislabeled as gateway API-key authentication")
	}
}

func TestPlaygroundSeparatesGatewayProviderAndWebSessionCredentials(t *testing.T) {
	html := string(playgroundHTML)
	for _, marker := range []string{
		`id="gateway-credential-map"`,
		`id="gateway-credential-map-lead"`,
		`id="gateway-route-status"`,
		`id="gateway-route-status-title"`,
		`id="gateway-route-status-help"`,
		`id="gateway-route-access-state"`,
		`id="gateway-route-provider-state"`,
		`id="gateway-route-model-state"`,
		`data-state="blocked"`,
		`Credential map — these are different`,
		`most students should leave both key fields empty`,
		`Default web-session route:`,
		`Google Developer API route:`,
		`BOB endpoint access:`,
		`Cookies belong to the engine; never paste cookies into either field.`,
		`BOB Gateway Access Key (optional)`,
		`Only when the gateway owner enabled api_keys`,
		`This protects access to the BOB endpoint; it is not a Google key.`,
		`type="password" id="gateway-api-key-input"`,
		`aria-describedby="gateway-auth-help"`,
		`Google Gemini Developer API key`,
		`aria-describedby="gemini-provider-help gemini-provider-route-note"`,
		`This is not the BOB Gateway Access Key.`,
		`Use your Google key for this session`,
		`function toggleApiKeyVisibility()`,
		`input.type = input.type === "password" ? "text" : "password";`,
		`keyInput.type = "password";`,
		`providerKeyInput.type = "password";`,
		`function clearGatewayApiKey()`,
		`function clearGeminiProviderKey()`,
		`gatewayAuthRequired = null;`,
		`Keep credentials out of the retained modal DOM`,
		`gatewayAuthErrorTitle:`,
		`gatewayAuthErrorHelp:`,
		`providerAuthError:`,
	} {
		if !strings.Contains(html, marker) {
			t.Fatalf("playground is missing credential-boundary marker %q", marker)
		}
	}
	for _, forbidden := range []string{
		`API Key (Bearer Auth)`,
		`placeholder="sk-gemini or leave empty for none"`,
		`title="Show / Hide API Key"`,
	} {
		if strings.Contains(html, forbidden) {
			t.Fatalf("playground still uses ambiguous gateway credential wording %q", forbidden)
		}
	}

	for _, marker := range []string{
		`gatewayCredentialMapTitle:`,
		`gatewayAuthTitle:`,
		`gatewayRouteWebHelp:`,
		`gatewayRouteProviderHelp:`,
		`gatewayRouteAccessHelp:`,
		`gatewayRouteStatusKicker:`,
		`gatewayRouteProviderKeyMissing:`,
		`gatewayRouteModelWarning:`,
		`gatewayRouteBoundary:`,
	} {
		if strings.Count(html, marker) < 2 {
			t.Fatalf("credential-boundary translation marker %q is not present in English and Hindi dictionaries", marker)
		}
	}
}

func TestPlaygroundCredentialTranslationsUseTheCorrectDictionaryScope(t *testing.T) {
	html := string(playgroundHTML)
	for _, marker := range []string{
		`"gateway-credential-map-title": dict.gatewayCredentialMapTitle,`,
		`"gateway-credential-map-lead": dict.gatewayCredentialMapLead,`,
		`"gateway-route-web-label": dict.gatewayRouteWebLabel,`,
		`"gateway-route-web-help": dict.gatewayRouteWebHelp,`,
		`"gateway-route-provider-label": dict.gatewayRouteProviderLabel,`,
		`"gateway-route-provider-help": dict.gatewayRouteProviderHelp,`,
		`"gateway-route-access-label": dict.gatewayRouteAccessLabel,`,
		`"gateway-route-access-help": dict.gatewayRouteAccessHelp,`,
		`"gateway-auth-title": dict.gatewayAuthTitle,`,
		`"gateway-auth-optional": dict.gatewayAuthOptional,`,
		`"gateway-auth-help": dict.gatewayAuthHelp,`,
		`gatewayKeyInput.placeholder = dict.gatewayAuthPlaceholder;`,
	} {
		if !strings.Contains(html, marker) {
			t.Fatalf("credential translation mapping is missing %q", marker)
		}
	}
	for _, forbidden := range []string{
		`ui.gatewayCredentialMapTitle`,
		`ui.gatewayCredentialMapLead`,
		`ui.gatewayRouteWebLabel`,
		`ui.gatewayRouteWebHelp`,
		`ui.gatewayRouteProviderLabel`,
		`ui.gatewayRouteProviderHelp`,
		`ui.gatewayRouteAccessLabel`,
		`ui.gatewayRouteAccessHelp`,
		`ui.gatewayAuthTitle`,
		`ui.gatewayAuthOptional`,
		`ui.gatewayAuthHelp`,
		`ui.gatewayAuthPlaceholder`,
	} {
		if strings.Contains(html, forbidden) {
			t.Fatalf("credential translation still reads top-level value through ui: %q", forbidden)
		}
	}
}

func TestPlaygroundBlocksIncompatibleDeveloperRouteBeforeSend(t *testing.T) {
	html := string(playgroundHTML)
	for _, marker := range []string{
		`function selectedStudioModelName()`,
		`function developerRouteModelIssue()`,
		`function developerRouteSelectionIssue()`,
		`if (!model || !/^gemini-/i.test(model)`,
		`The selected endpoint is not trusted for a provider key.`,
		`const routeIssue = developerRouteSelectionIssue();`,
		`showToast(routeIssue, "⚠️");`,
	} {
		if !strings.Contains(html, marker) {
			t.Fatalf("playground is missing route guard marker %q", marker)
		}
	}
	sendStart := strings.Index(html, "async function sendMessage()")
	if sendStart < 0 {
		t.Fatal("sendMessage is missing")
	}
	routeGuard := strings.Index(html[sendStart:], `const routeIssue = developerRouteSelectionIssue();`)
	generationStart := strings.Index(html[sendStart:], "isGenerating = true;")
	if routeGuard < 0 || generationStart < 0 || routeGuard > generationStart {
		t.Fatal("incompatible Developer API route is not blocked before generation state or network work")
	}
}

func TestDeveloperAPIRouteToggleFailsClosedWithoutKey(t *testing.T) {
	html := string(playgroundHTML)
	start := strings.Index(html, "function toggleGeminiProviderRoute(enabled)")
	if start < 0 {
		t.Fatal("Developer API route toggle is missing")
	}
	endOffset := strings.Index(html[start:], "\n}\n\nfunction toggleGeminiProviderKeyVisibility")
	if endOffset < 0 {
		t.Fatal("Developer API route toggle boundary is missing")
	}
	source := html[start : start+endOffset]
	for _, marker := range []string{
		`if (enabled && !geminiProviderKey)`,
		`useGeminiProvider = false;`,
		`if (toggle) toggle.checked = false;`,
		`Paste your own Google AI Studio key first`,
		`if (enabled && !canSendGeminiProviderKey())`,
	} {
		if !strings.Contains(source, marker) {
			t.Fatalf("Developer API route toggle is missing fail-closed marker %q", marker)
		}
	}
	if strings.Contains(source, "useGeminiProvider = Boolean(enabled);\n  if (!geminiProviderKey)") {
		t.Fatal("Developer API route toggle enables an unusable provider state before validating the key")
	}
}

func TestDeveloperAPIRouteRequiresSafeGatewayTransport(t *testing.T) {
	html := string(playgroundHTML)
	start := strings.Index(html, "function canSendGeminiProviderKey()")
	if start < 0 {
		t.Fatal("Developer API transport guard is missing")
	}
	endOffset := strings.Index(html[start:], "\n}\n\nfunction isLoopbackGatewayURL")
	if endOffset < 0 {
		t.Fatal("Developer API transport guard boundary is missing")
	}
	source := html[start : start+endOffset]
	for _, marker := range []string{
		`const endpoint = getGatewayBaseUrl();`,
		`const endpointIsLoopback = isLoopbackGatewayURL(endpoint);`,
		`if (!isSecureGatewayForProviderKey(endpoint)) return false;`,
		`(isNativeDesktopStudio() || isLoopbackPage()) && endpointIsLoopback`,
		`return hasExplicitGatewayEndpoint();`,
	} {
		if !strings.Contains(source, marker) {
			t.Fatalf("Developer API transport guard is missing marker %q", marker)
		}
	}
	if strings.Contains(source, "if (isNativeDesktopStudio() || (isLoopbackPage() && isLoopbackGatewayURL(getGatewayBaseUrl())))") {
		t.Fatal("native context still bypasses the configured gateway transport policy")
	}

	secureStart := strings.Index(html, "function isSecureGatewayForProviderKey(rawURL)")
	if secureStart < 0 {
		t.Fatal("Developer API secure transport helper is missing")
	}
	secureEndOffset := strings.Index(html[secureStart:], "\n}\n\nfunction openGatewayModal")
	if secureEndOffset < 0 {
		t.Fatal("Developer API secure transport helper boundary is missing")
	}
	secureSource := html[secureStart : secureStart+secureEndOffset]
	for _, marker := range []string{
		`parsedURL.protocol === "https:"`,
		`parsedURL.protocol === "http:" && isLoopbackGatewayURL(parsedURL.href)`,
		`if (!parsedURL.hostname || parsedURL.username || parsedURL.password) return false;`,
	} {
		if !strings.Contains(secureSource, marker) {
			t.Fatalf("Developer API secure transport helper is missing marker %q", marker)
		}
	}
}

func TestPlaygroundAccessibilityFloorIsExplicit(t *testing.T) {
	html := string(playgroundHTML)
	for _, marker := range []string{
		`<meta name="viewport" content="width=device-width, initial-scale=1.0, viewport-fit=cover, interactive-widget=resizes-content">`,
		`button:focus-visible`,
		`@media (prefers-reduced-motion: reduce)`,
		`const MANAGED_MODAL_SELECTOR = '.dialog-backdrop.open, .cmd-modal-backdrop.open';`,
		`const modalReturnFocus = new WeakMap();`,
		`function openManagedModal(modal, initialSelector)`,
		`function closeManagedModal(modal)`,
		`function trapManagedModalFocus(e, modal)`,
		`if (openModal && trapManagedModalFocus(e, openModal)) return;`,
		`target.click();`,
		`function enhanceCommandPaletteAccessibility()`,
		`list.setAttribute("role", "listbox");`,
		`aria-activedescendant`,
		`<button type="button" class="brand-anchor"`,
		`class="inline-help-trigger" role="button" tabindex="0"`,
		`class="token-pill" role="button" tabindex="0"`,
		`<button type="button" class="active-model-chip"`,
		`<button type="button" id="prompt-token-estimate"`,
		`aria-labelledby="about-modal-title"`,
		`id="about-modal-title"`,
		`id="local-onboard-modal" class="dialog-backdrop open" role="dialog" aria-modal="true"`,
		`function closeLocalOnboardingModal()`,
		`aria-label="Chat prompt"`,
		`aria-label="Attach or paste an image"`,
		`aria-label="Gateway endpoint URL"`,
		`aria-label="Google Gemini Developer API key"`,
	} {
		if !strings.Contains(html, marker) {
			t.Fatalf("playground is missing accessibility marker %q", marker)
		}
	}

	for _, id := range []string{
		"doc-preview-modal",
		"confirm-modal",
		"custom-theme-modal",
		"about-modal",
		"gateway-modal",
		"artifact-modal",
		"instructions-modal",
		"tokenizer-modal",
		"glossary-modal",
		"cmd-modal",
	} {
		marker := `id="` + id + `"`
		start := strings.Index(html, marker)
		if start < 0 {
			t.Fatalf("modal %q is missing", id)
		}
		end := strings.Index(html[start:], ">")
		if end < 0 {
			t.Fatalf("modal %q opening tag is incomplete", id)
		}
		openingTag := html[start : start+end]
		if !strings.Contains(openingTag, `role="dialog"`) ||
			!strings.Contains(openingTag, `aria-modal="true"`) ||
			!strings.Contains(openingTag, `aria-hidden="true"`) {
			t.Fatalf("modal %q is missing explicit dialog semantics", id)
		}
	}

	if strings.Contains(html, `maximum-scale=1.0`) || strings.Contains(html, `user-scalable=no`) {
		t.Fatal("the page must not disable browser text scaling")
	}
	if strings.Contains(html, `class="modal-overlay`) || strings.Contains(html, `class="modal-content`) {
		t.Fatal("onboarding must use the shared dialog surface, not undefined modal classes")
	}

	aboutStart := strings.Index(html, "const aboutModal = document.getElementById(\"about-modal\")")
	if aboutStart < 0 {
		t.Fatal("about modal keyboard handler boundaries are missing")
	}
	aboutEnd := strings.Index(html[aboutStart:], "const gatewayModal = document.getElementById(\"gateway-modal\")")
	if aboutEnd < 0 {
		t.Fatal("about modal keyboard handler boundaries are missing")
	}
	if strings.Contains(html[aboutStart:aboutStart+aboutEnd], `e.key === 'Escape' || e.key === 'Enter'`) {
		t.Fatal("pressing Enter inside the About dialog must not dismiss it")
	}
}

func TestPlaygroundUsesNativeControlAndDrawerSemantics(t *testing.T) {
	html := string(playgroundHTML)

	buttonPattern := regexp.MustCompile(`(?s)<button\b[^>]*>`)
	for _, openingTag := range buttonPattern.FindAllString(html, -1) {
		if !strings.Contains(openingTag, `type="button"`) {
			t.Errorf("button is missing an explicit non-submit type: %s", openingTag)
		}
	}

	for _, marker := range []string{
		`<a class="skip-link" href="#user-input">Skip to prompt</a>`,
		`id="theme-selector" aria-label="Color theme"`,
		`id="lang-selector" aria-label="UI language"`,
		`id="btn-toggle-left"`,
		`id="btn-toggle-right"`,
		`aria-controls="sidebar-left"`,
		`aria-controls="sidebar-right"`,
		`aria-expanded="false"`,
		`id="sidebar-left" aria-hidden="true" aria-labelledby="left-panel-title" inert`,
		`id="sidebar-right" aria-hidden="true" aria-labelledby="right-panel-title" inert`,
		`id="instr-modal-title"`,
		`aria-labelledby="instr-modal-title"`,
		`id="cmd-modal-title"`,
		`aria-labelledby="cmd-modal-title"`,
		`panel.setAttribute("aria-hidden", isOpen ? "false" : "true")`,
		`panel.toggleAttribute("inert", !isOpen)`,
		`btn.setAttribute("aria-expanded", isOpen ? "true" : "false")`,
		`const sidebarReturnFocus = new WeakMap();`,
		`function getOpenResponsiveSidebar()`,
		`function trapResponsiveSidebarFocus(e, panel)`,
		`focusSidebarInitial(panel)`,
		`restoreSidebarFocus(panel, side)`,
		`panel.setAttribute("role", isResponsiveDrawer ? "dialog" : "complementary")`,
		`panel.setAttribute("aria-modal", "true")`,
		`toggleSidebar(openSidebar.id === 'sidebar-left' ? 'left' : 'right')`,
		`--bg-hover: var(--bg-card-hover);`,
		`--bg-main: var(--bg-app);`,
		`max-height: min(90dvh, calc(100dvh - 32px));`,
		`min-height: 44px;`,
	} {
		if !strings.Contains(html, marker) {
			t.Fatalf("playground is missing native-control or responsive-semantics marker %q", marker)
		}
	}

	for _, marker := range []string{
		`<button type="button" class="brand-anchor"`,
		`<button type="button" class="active-model-chip"`,
		`<button type="button" class="starter-card"`,
		`<button type="button" id="prompt-token-estimate"`,
		`<button type="button" class="artifact-side-rail left"`,
		`<button type="button" class="artifact-side-rail right"`,
	} {
		if !strings.Contains(html, marker) {
			t.Fatalf("click-only core control was not promoted to a native button: %q", marker)
		}
	}

	if strings.Contains(html, `class="brand-anchor" role="button"`) ||
		strings.Contains(html, `class="active-model-chip" role="button"`) ||
		strings.Contains(html, `class="starter-card" role="button"`) ||
		strings.Contains(html, `class="token-live-chip" role="button"`) {
		t.Fatal("core click-only controls still rely on generic role=button semantics")
	}
}

func TestPlaygroundHeaderAndGenerationStatesHaveAccessibleNames(t *testing.T) {
	html := string(playgroundHTML)
	for _, marker := range []string{
		`id="btn-gateway-status" aria-label="Gateway connection status and settings"`,
		`id="btn-toggle-left" aria-label="Toggle configuration panel"`,
		`id="btn-toggle-right" aria-label="Toggle integration code panel"`,
		`id="btn-cmd-menu" aria-label="Open command palette"`,
		`class="nav-pill-btn github-pill" aria-label="Open BOB Gemini Free on GitHub"`,
		`id="btn-glossary" aria-label="Open AI and systems glossary"`,
		`id="btn-translit" aria-label="Toggle Indic phonetic typing" aria-pressed="false"`,
		`id="btn-mic" aria-label="Voice input" aria-pressed="false"`,
		`id="send-btn" aria-label="Send prompt" aria-busy="false"`,
		`statusButton.setAttribute("aria-label", online`,
		`btn.setAttribute("aria-pressed", isTransliterationActive ? "true" : "false")`,
		`micBtn.setAttribute("aria-pressed", "true")`,
		`micBtn.setAttribute("aria-label", "Stop voice input")`,
		`sendBtn.setAttribute("aria-label", dict.btnStop || "Stop generation")`,
		`sendBtn.setAttribute("aria-busy", "true")`,
		`sendBtn.setAttribute("aria-busy", "false")`,
		`.btn-user-action:focus-visible`,
	} {
		if !strings.Contains(html, marker) {
			t.Fatalf("playground is missing accessible name/state marker %q", marker)
		}
	}
}

func TestPlaygroundThemeTextContrast(t *testing.T) {
	html := string(playgroundHTML)
	blockPattern := regexp.MustCompile(`(?s)(:root|\[data-theme="([^"]+)"\]) \{(.*?)\n\}`)
	variablePattern := regexp.MustCompile(`--([a-z-]+):\s*(#[0-9a-fA-F]{6})`)
	blocks := blockPattern.FindAllStringSubmatch(html, -1)
	if len(blocks) == 0 {
		t.Fatal("playground theme token blocks are missing")
	}

	for _, block := range blocks {
		name := block[2]
		if name == "" {
			name = "bob-builder"
		}
		vars := make(map[string]string)
		for _, match := range variablePattern.FindAllStringSubmatch(block[3], -1) {
			vars[match[1]] = match[2]
		}
		for _, foreground := range []string{"text-main", "text-muted", "text-subdued"} {
			for _, background := range []string{"bg-app", "bg-card", "bg-modal", "bg-input"} {
				fg, okFG := vars[foreground]
				bg, okBG := vars[background]
				if !okFG || !okBG {
					continue
				}
				contrast, err := cssHexContrast(fg, bg)
				if err != nil {
					t.Fatalf("theme %s has invalid contrast tokens %s/%s: %v", name, fg, bg, err)
				}
				if contrast < 4.5 {
					t.Errorf("theme %s has %s on %s contrast %.2f; normal text requires at least 4.5:1", name, foreground, background, contrast)
				}
			}
		}
	}
}

func cssHexContrast(foreground, background string) (float64, error) {
	parse := func(value string) (float64, error) {
		if len(value) != 7 || value[0] != '#' {
			return 0, fmt.Errorf("invalid CSS hex color %q", value)
		}
		channels := make([]float64, 3)
		for i := range channels {
			parsed, err := strconv.ParseUint(value[1+i*2:3+i*2], 16, 8)
			if err != nil {
				return 0, err
			}
			channels[i] = float64(parsed) / 255
		}
		for i, channel := range channels {
			if channel <= 0.03928 {
				channels[i] = channel / 12.92
			} else {
				channels[i] = math.Pow((channel+0.055)/1.055, 2.4)
			}
		}
		return 0.2126*channels[0] + 0.7152*channels[1] + 0.0722*channels[2], nil
	}

	fg, err := parse(foreground)
	if err != nil {
		return 0, err
	}
	bg, err := parse(background)
	if err != nil {
		return 0, err
	}
	if fg < bg {
		fg, bg = bg, fg
	}
	return (fg + 0.05) / (bg + 0.05), nil
}
