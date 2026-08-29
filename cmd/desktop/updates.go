package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"github.com/div197/bob-gemini-free/internal/updater"
)

const (
	// A short startup delay keeps an update check from competing with the
	// gateway/bootstrap path. The long interval avoids turning a classroom
	// fleet restart into a burst of GitHub API traffic.
	desktopAutoUpdateInitialDelay = 30 * time.Second
	desktopAutoUpdateInterval     = 24 * time.Hour
)

func desktopUpdateChecksEnabled() bool {
	if !updater.IsDesktopVersionCheckable(desktopVersion) {
		return false
	}
	return desktopChannel == updater.DesktopChannelStable || desktopChannel == updater.DesktopChannelPreview
}

// startAutomaticDesktopUpdateChecks schedules low-frequency metadata checks.
// It never downloads or replaces an app without the user's explicit choice in
// the native update dialog.
func (a *App) startAutomaticDesktopUpdateChecks(ctx context.Context) {
	if a == nil || ctx == nil || !desktopUpdateChecksEnabled() {
		return
	}

	updateCtx, cancel := context.WithCancel(ctx)
	a.updateMu.Lock()
	if a.updateCancel != nil {
		a.updateCancel()
	}
	a.updateCancel = cancel
	a.updateMu.Unlock()

	go func() {
		initialTimer := time.NewTimer(desktopAutoUpdateInitialDelay)
		defer initialTimer.Stop()
		select {
		case <-updateCtx.Done():
			return
		case <-initialTimer.C:
		}

		a.checkDesktopUpdate(updateCtx, true)

		ticker := time.NewTicker(desktopAutoUpdateInterval)
		defer ticker.Stop()
		for {
			select {
			case <-updateCtx.Done():
				return
			case <-ticker.C:
				a.checkDesktopUpdate(updateCtx, true)
			}
		}
	}()
}

func (a *App) stopAutomaticDesktopUpdateChecks() {
	if a == nil {
		return
	}
	a.updateMu.Lock()
	defer a.updateMu.Unlock()
	if a.updateCancel != nil {
		a.updateCancel()
		a.updateCancel = nil
	}
}

func (a *App) checkDesktopUpdate(ctx context.Context, automatic bool) {
	if a == nil || ctx == nil || ctx.Err() != nil {
		return
	}
	if !desktopUpdateChecksEnabled() {
		if !automatic {
			a.showDesktopInfo(ctx, "Updates unavailable", "This local development build is not connected to the public desktop updater.")
		}
		return
	}

	update, err := updater.CheckLatestDesktopForChannelContext(ctx, desktopVersion, desktopChannel)
	if err != nil {
		if automatic {
			log.Printf("automatic desktop update check skipped: %v", err)
			return
		}
		a.showDesktopError(ctx, "Update check failed", desktopUpdateErrorMessage("BOB could not check for desktop updates", err))
		return
	}
	if !update.HasUpdate {
		if !automatic {
			a.showDesktopInfo(ctx, "BOB Gemini Free updates", desktopNoUpdateMessage(update))
		}
		return
	}
	if automatic && !update.AssetAvailable {
		// A CLI-only release is not actionable for the native app. The explicit
		// Help action still explains this case to a user who asks for details.
		return
	}
	a.showDesktopUpdateDialog(ctx, update, automatic)
}

func (a *App) showDesktopUpdateDialog(ctx context.Context, update *updater.DesktopCheckResult, automatic bool) {
	if a == nil || ctx == nil || update == nil || ctx.Err() != nil {
		return
	}
	if automatic && !a.claimAutomaticUpdatePrompt(update.LatestVersion) {
		return
	}

	// Prevent an automatic prompt and a simultaneous Help-menu action from
	// opening competing native dialogs.
	a.updateDialogMu.Lock()
	defer a.updateDialogMu.Unlock()

	if update.AssetAvailable && update.ManifestAvailable {
		button, dialogErr := runtime.MessageDialog(ctx, runtime.MessageDialogOptions{
			Type:          runtime.QuestionDialog,
			Title:         "Update available",
			Message:       desktopUpdateAvailableMessage(update.LatestVersion),
			Buttons:       []string{"Install Update", "Open Releases", "Cancel"},
			DefaultButton: "Install Update",
			CancelButton:  "Cancel",
		})
		if dialogErr != nil {
			return
		}
		if button == "Open Releases" {
			runtime.BrowserOpenURL(ctx, update.ReleaseURL)
			return
		}
		if button != "Install Update" {
			return
		}

		plan, stageErr := updater.StageDesktopUpdate(update)
		if stageErr != nil {
			a.showDesktopError(ctx, "Update could not be staged", desktopUpdateErrorMessage("BOB could not prepare this update", stageErr))
			return
		}
		if startErr := updater.StartDesktopUpdate(plan); startErr != nil {
			a.showDesktopError(ctx, "Update could not start", desktopUpdateErrorMessage("BOB could not start the verified update", startErr))
			return
		}
		runtime.Quit(ctx)
		return
	}

	if !update.ManifestAvailable {
		button, dialogErr := runtime.MessageDialog(ctx, runtime.MessageDialogOptions{
			Type:          runtime.QuestionDialog,
			Title:         "Manual update required",
			Message:       desktopManualUpdateMessage(update.LatestVersion),
			Buttons:       []string{"Open Releases", "Cancel"},
			DefaultButton: "Open Releases",
			CancelButton:  "Cancel",
		})
		if dialogErr == nil && button == "Open Releases" {
			runtime.BrowserOpenURL(ctx, update.ReleaseURL)
		}
		return
	}

	button, dialogErr := runtime.MessageDialog(ctx, runtime.MessageDialogOptions{
		Type:          runtime.QuestionDialog,
		Title:         "Desktop package unavailable",
		Message:       desktopPackageUnavailableMessage(update.LatestVersion),
		Buttons:       []string{"Open Releases", "Cancel"},
		DefaultButton: "Open Releases",
		CancelButton:  "Cancel",
	})
	if dialogErr == nil && button == "Open Releases" {
		runtime.BrowserOpenURL(ctx, update.ReleaseURL)
	}
}

func (a *App) showDesktopInfo(ctx context.Context, title, message string) {
	if a == nil || ctx == nil || ctx.Err() != nil {
		return
	}
	_, _ = runtime.MessageDialog(ctx, runtime.MessageDialogOptions{
		Type:    runtime.InfoDialog,
		Title:   title,
		Message: message,
		Buttons: []string{"OK"},
	})
}

func (a *App) showDesktopError(ctx context.Context, title, message string) {
	if a == nil || ctx == nil || ctx.Err() != nil {
		return
	}
	_, _ = runtime.MessageDialog(ctx, runtime.MessageDialogOptions{
		Type:    runtime.ErrorDialog,
		Title:   title,
		Message: message,
		Buttons: []string{"OK"},
	})
}

func (a *App) claimAutomaticUpdatePrompt(version string) bool {
	a.updateMu.Lock()
	defer a.updateMu.Unlock()
	if a.updatePromptedVersion == version {
		return false
	}
	a.updatePromptedVersion = version
	return true
}

func desktopUpdateAvailableMessage(version string) string {
	return fmt.Sprintf("A signed desktop update is available: %s.\n\nBOB will download and verify it before replacing this app, then restart to apply it.\n\nInstall now?", version)
}

func desktopManualUpdateMessage(version string) string {
	return fmt.Sprintf("A newer desktop release is available: %s.\n\nThis release does not publish the signed manifest required for automatic installation. Open the official release page to download it manually?", version)
}

func desktopPackageUnavailableMessage(version string) string {
	return fmt.Sprintf("Release %s exists, but it does not contain a native package for this platform yet.\n\nOpen the official release page?", version)
}

func desktopNoUpdateMessage(update *updater.DesktopCheckResult) string {
	if update == nil {
		return "No desktop update information is available."
	}
	return fmt.Sprintf("This build is %s.\n\nThe latest public release is %s.\n\nNo newer desktop release is available.", update.CurrentVersion, update.LatestVersion)
}

func desktopUpdateErrorMessage(summary string, err error) string {
	if err == nil {
		return summary + "."
	}
	return fmt.Sprintf("%s.\n\nDetails: %s", summary, err)
}
