package main

import (
	"errors"
	"strings"
	"testing"

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
