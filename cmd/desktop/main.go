package main

import (
	"context"
	"embed"
	"fmt"
	"time"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/menu"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/runtime"

	"github.com/div197/bob-gemini-free/internal/server"
	"github.com/div197/bob-gemini-free/internal/updater"
)

const desktopVersion = "v0.1.7"

//go:embed all:frontend
var assets embed.FS

func main() {
	// 1. Run the local Gateway engine in the background
	cfg := loadDesktopConfig()

	srv := server.New(cfg, desktopVersion)
	app := NewApp()
	gateway, err := startDesktopGateway(cfg.Port, srv.Handler())
	if err != nil {
		fmt.Printf("Gateway startup failed: %v\n", err)
		if windowErr := wails.Run(desktopOptions(app, nil, err)); windowErr != nil {
			fmt.Printf("Desktop startup error window failed: %v\n", windowErr)
		}
		return
	}
	fmt.Printf("🚀 Wails Internal Gateway listening on %s\n", gateway.Endpoint())
	app.gatewayURL = gateway.Endpoint()

	// 2. Launch the Wails Native Window, pointing it to our Gateway UI
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
		Menu:   desktopMenu(app),
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		OnStartup: func(ctx context.Context) {
			app.startup(ctx)
			if startupErr != nil {
				runtime.EventsEmit(ctx, "gateway-error", startupErr.Error())
				return
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
		update, err := updater.CheckLatestDesktop(desktopVersion)
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
			message = fmt.Sprintf("A newer desktop release is available: %s\\n\\nOpen the official release page to download it?", update.LatestVersion)
			button, dialogErr := runtime.MessageDialog(app.ctx, runtime.MessageDialogOptions{
				Type:          runtime.QuestionDialog,
				Title:         "Update available",
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
