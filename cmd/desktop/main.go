package main

import (
	"context"
	"embed"
	"fmt"
	"os"
	"time"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/menu"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	macoptions "github.com/wailsapp/wails/v2/pkg/options/mac"
	"github.com/wailsapp/wails/v2/pkg/runtime"

	"github.com/div197/bob-gemini-free/internal/server"
	"github.com/div197/bob-gemini-free/internal/updater"
)

// desktopVersion is injected by a release build. Ordinary developer builds
// remain explicitly non-updatable instead of pretending to be a published
// release.
var desktopVersion = "dev"
var desktopChannel = updater.DesktopChannelStable

//go:embed all:frontend
var assets embed.FS

func main() {
	if handled, err := updater.HandleDesktopUpdateCommand(os.Args[1:]); handled {
		if err != nil {
			fmt.Printf("Desktop update helper failed: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// 1. Run the local Gateway engine in the background
	cfg := loadDesktopConfig()

	srv := server.New(cfg, desktopVersion)
	app := NewApp()
	app.updateConfirmationPath = updater.DesktopUpdateConfirmationPath(os.Args[1:])
	gateway, err := startDesktopGateway(cfg.Port, srv.Handler())
	if err != nil {
		fmt.Printf("Gateway startup failed: %v\n", err)
		if windowErr := wails.Run(desktopOptions(app, nil, err)); windowErr != nil {
			fmt.Printf("Desktop startup error window failed: %v\n", windowErr)
		}
		return
	}
	fmt.Printf("🚀 BOB Gemini Free local engine listening on %s\n", gateway.Endpoint())
	app.gatewayURL = gateway.Endpoint()

	// 2. Launch the branded native window, pointing it to our Gateway UI
	err = wails.Run(desktopOptions(app, gateway, nil))

	if err != nil {
		println("Error:", err.Error())
	}
}

func desktopOptions(app *App, gateway *desktopGateway, startupErr error) *options.App {
	return &options.App{
		Title:  "BOB Gemini Free",
		Width:  1100,
		Height: 800,
		// Wails disables the native macOS zoom/maximize button unless the Mac
		// options are present and DisableZoom is explicitly false. Keep the
		// standard window resizable so users can maximize the desktop studio.
		Mac: &macoptions.Options{
			DisableZoom: false,
		},
		Menu: desktopMenu(app),
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		OnStartup: func(ctx context.Context) {
			app.startup(ctx)
			if startupErr != nil {
				runtime.EventsEmit(ctx, "gateway-error", startupErr.Error())
				return
			}
			if app.updateConfirmationPath != "" {
				if confirmErr := updater.ConfirmDesktopUpdate(app.updateConfirmationPath); confirmErr != nil {
					fmt.Printf("Desktop update confirmation failed: %v\n", confirmErr)
				}
			}
			runtime.EventsEmit(ctx, "gateway-ready", gateway.Endpoint())
		},
		OnShutdown: func(ctx context.Context) {
			if gateway == nil {
				return
			}
			shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()
			if shutdownErr := gateway.Shutdown(shutdownCtx); shutdownErr != nil {
				fmt.Printf("Gateway shutdown failed: %v\n", shutdownErr)
			}
		},
		Bind: []interface{}{app},
	}
}

func desktopMenu(app *App) *menu.Menu {
	result := menu.NewMenuFromItems(menu.AppMenu(), menu.EditMenu(), menu.WindowMenu())
	help := result.AddSubmenu("Help")
	help.AddText("Check for Updates", nil, func(*menu.CallbackData) {
		if app.ctx == nil {
			return
		}
		update, err := updater.CheckLatestDesktopForChannel(desktopVersion, desktopChannel)
		if err != nil {
			_, _ = runtime.MessageDialog(app.ctx, runtime.MessageDialogOptions{
				Type:    runtime.ErrorDialog,
				Title:   "Update check failed",
				Message: err.Error(),
				Buttons: []string{"OK"},
			})
			return
		}
		message := fmt.Sprintf("This build is %s. The latest public release is %s. No newer release is available.", update.CurrentVersion, update.LatestVersion)
		if update.HasUpdate && update.AssetAvailable {
			if !update.ManifestAvailable {
				message = fmt.Sprintf("A newer desktop release is available: %s, but it has no signed update manifest. Open the official release page for the manual download?", update.LatestVersion)
				button, dialogErr := runtime.MessageDialog(app.ctx, runtime.MessageDialogOptions{
					Type:          runtime.QuestionDialog,
					Title:         "Manual update required",
					Message:       message,
					Buttons:       []string{"Open Releases", "Cancel"},
					DefaultButton: "Open Releases",
					CancelButton:  "Cancel",
				})
				if dialogErr == nil && button == "Open Releases" {
					runtime.BrowserOpenURL(app.ctx, update.ReleaseURL)
				}
				return
			}
			message = fmt.Sprintf("A newer signed desktop release is available: %s\\n\\nDownload, verify, and install it after restarting BOB?", update.LatestVersion)
			button, dialogErr := runtime.MessageDialog(app.ctx, runtime.MessageDialogOptions{
				Type:          runtime.QuestionDialog,
				Title:         "Update available",
				Message:       message,
				Buttons:       []string{"Install Update", "Open Releases", "Cancel"},
				DefaultButton: "Install Update",
				CancelButton:  "Cancel",
			})
			if dialogErr == nil && button == "Install Update" {
				plan, stageErr := updater.StageDesktopUpdate(update)
				if stageErr != nil {
					_, _ = runtime.MessageDialog(app.ctx, runtime.MessageDialogOptions{
						Type:    runtime.ErrorDialog,
						Title:   "Update could not be staged",
						Message: stageErr.Error(),
						Buttons: []string{"OK"},
					})
					return
				}
				if startErr := updater.StartDesktopUpdate(plan); startErr != nil {
					_, _ = runtime.MessageDialog(app.ctx, runtime.MessageDialogOptions{
						Type:    runtime.ErrorDialog,
						Title:   "Update could not start",
						Message: startErr.Error(),
						Buttons: []string{"OK"},
					})
					return
				}
				runtime.Quit(app.ctx)
				return
			}
			if dialogErr == nil && button == "Open Releases" {
				runtime.BrowserOpenURL(app.ctx, update.ReleaseURL)
			}
			return
		}
		if update.HasUpdate && !update.AssetAvailable {
			message = fmt.Sprintf("Release %s exists, but it does not contain a native package for this platform yet. Open the official release page?", update.LatestVersion)
			button, dialogErr := runtime.MessageDialog(app.ctx, runtime.MessageDialogOptions{
				Type:          runtime.QuestionDialog,
				Title:         "Desktop package unavailable",
				Message:       message,
				Buttons:       []string{"Open Releases", "Cancel"},
				DefaultButton: "Open Releases",
				CancelButton:  "Cancel",
			})
			if dialogErr == nil && button == "Open Releases" {
				runtime.BrowserOpenURL(app.ctx, update.ReleaseURL)
			}
			return
		}
		_, _ = runtime.MessageDialog(app.ctx, runtime.MessageDialogOptions{
			Type:    runtime.InfoDialog,
			Title:   "BOB Gemini Free updates",
			Message: message,
			Buttons: []string{"OK"},
		})
	})
	help.AddText("Open GitHub Releases", nil, func(*menu.CallbackData) {
		if app.ctx != nil {
			runtime.BrowserOpenURL(app.ctx, updater.DesktopReleaseURL)
		}
	})
	return result
}
