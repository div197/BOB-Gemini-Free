package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/div197/bob-gemini-free/internal/updater"
)

func TestDesktopOptionsKeepNativeWindowMaximizable(t *testing.T) {
	opts := desktopOptions(NewApp(), nil, nil, nil)
	if opts.DisableResize {
		t.Fatal("desktop window unexpectedly disables resizing")
	}
	if opts.Mac == nil {
		t.Fatal("macOS options are required to keep the native zoom button enabled")
	}
	if opts.Mac.DisableZoom {
		t.Fatal("desktop window unexpectedly disables the native macOS zoom button")
	}
}

func TestDesktopBootstrapKeepsTheNativeBridgeAfterGatewayLoad(t *testing.T) {
	source, err := os.ReadFile("frontend/index.html")
	if err != nil {
		t.Fatalf("read desktop bootstrap: %v", err)
	}
	html := string(source)
	for _, marker := range []string{
		`import { BrowserOpenURL, EventsOn } from "/wailsjs/runtime/runtime.js";`,
		`data-gateway-frame`,
		`window.addEventListener("message"`,
		`event.source !== frame.contentWindow || event.origin !== gatewayOrigin`,
		`message.type !== "BOB_OPEN_EXTERNAL_URL"`,
		`BrowserOpenURL(url.href)`,
		`desktop_shell=1`,
	} {
		if !strings.Contains(html, marker) {
			t.Fatalf("desktop bootstrap is missing marker %q", marker)
		}
	}
	if strings.Contains(html, "window.location.replace(`${gateway.origin}/playground") {
		t.Fatal("desktop bootstrap navigates away from the Wails shell and loses native bridges")
	}
}

func TestDesktopUpdateMessagesUseNativeLineBreaks(t *testing.T) {
	messages := []string{
		desktopUpdateAvailableMessage("v0.2.0"),
		desktopManualUpdateMessage("v0.2.0-preview.2"),
		desktopPackageUnavailableMessage("v0.2.0"),
		desktopNoUpdateMessage(&updater.DesktopCheckResult{CurrentVersion: "v0.2.0", LatestVersion: "v0.2.0"}),
		desktopUpdateErrorMessage("BOB could not prepare this update", errors.New("read-only")),
	}
	for _, message := range messages {
		if strings.Contains(message, `\n`) {
			t.Fatalf("update dialog contains a literal escaped newline: %q", message)
		}
	}
	if !strings.Contains(desktopUpdateAvailableMessage("v0.2.0"), "\n\n") {
		t.Fatal("update dialog message is missing paragraph separation")
	}
}

func TestDesktopUpdateCheckErrorMessageHidesTransportDetails(t *testing.T) {
	timeout := desktopUpdateCheckErrorMessage(fmt.Errorf("GitHub request failed: %w", context.DeadlineExceeded))
	if strings.Contains(timeout, "context deadline exceeded") || strings.Contains(timeout, "Client.Timeout") {
		t.Fatalf("timeout details leaked into update dialog: %q", timeout)
	}
	if !strings.Contains(timeout, "Nothing was changed") || !strings.Contains(timeout, "try again") {
		t.Fatalf("timeout dialog is not actionable: %q", timeout)
	}

	policy := desktopUpdateCheckErrorMessage(errors.New("desktop update server returned HTTP 429"))
	if strings.Contains(policy, "HTTP 429") || strings.Contains(policy, "desktop update server") {
		t.Fatalf("release metadata details leaked into update dialog: %q", policy)
	}
}

func TestDesktopUpdateOperationErrorMessageHidesLocalPaths(t *testing.T) {
	path := "/private/var/folders/example/bob-gemini-free-update-123"
	readOnly := desktopUpdateErrorMessage("BOB could not prepare this update", fmt.Errorf("create staging directory %s: %w", path, syscall.EROFS))
	if strings.Contains(readOnly, path) || strings.Contains(readOnly, "read-only file system") {
		t.Fatalf("local filesystem details leaked into update dialog: %q", readOnly)
	}
	if !strings.Contains(readOnly, "Applications") || !strings.Contains(readOnly, "Nothing was changed") {
		t.Fatalf("read-only recovery guidance is incomplete: %q", readOnly)
	}

	generic := desktopUpdateErrorMessage("BOB could not start the verified update", fmt.Errorf("exec %s: permission denied", path))
	if strings.Contains(generic, path) || !strings.Contains(generic, "Nothing was changed") {
		t.Fatalf("generic update error was not bounded: %q", generic)
	}
}

func TestDesktopUpdateChecksSkipDevelopmentBuilds(t *testing.T) {
	originalVersion, originalChannel := desktopVersion, desktopChannel
	defer func() {
		desktopVersion, desktopChannel = originalVersion, originalChannel
	}()

	desktopVersion = "dev"
	desktopChannel = "stable"
	if desktopUpdateChecksEnabled() {
		t.Fatal("development build unexpectedly enabled automatic update checks")
	}

	desktopVersion = "v0.2.0"
	desktopChannel = "stable"
	if !desktopUpdateChecksEnabled() {
		t.Fatal("published stable build unexpectedly disabled automatic update checks")
	}

	desktopChannel = "unknown"
	if desktopUpdateChecksEnabled() {
		t.Fatal("unsupported update channel unexpectedly enabled automatic update checks")
	}
}

func TestDesktopAutoUpdateStartupDelayClampsJitter(t *testing.T) {
	original := desktopAutoUpdateJitterFn
	t.Cleanup(func() { desktopAutoUpdateJitterFn = original })

	for _, test := range []struct {
		name   string
		jitter time.Duration
		want   time.Duration
	}{
		{name: "negative", jitter: -time.Second, want: desktopAutoUpdateInitialDelay},
		{name: "bounded", jitter: 17 * time.Second, want: desktopAutoUpdateInitialDelay + 17*time.Second},
		{name: "maximum", jitter: desktopAutoUpdateMaxJitter, want: desktopAutoUpdateInitialDelay + desktopAutoUpdateMaxJitter},
		{name: "over maximum", jitter: desktopAutoUpdateMaxJitter + time.Second, want: desktopAutoUpdateInitialDelay + desktopAutoUpdateMaxJitter},
	} {
		t.Run(test.name, func(t *testing.T) {
			desktopAutoUpdateJitterFn = func() time.Duration { return test.jitter }
			if got := desktopAutoUpdateInitialWait(); got != test.want {
				t.Fatalf("initial wait = %s, want %s", got, test.want)
			}
		})
	}
}
