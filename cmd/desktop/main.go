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
	updateConfirmationPath := updater.DesktopUpdateConfirmationPath(os.Args[1:])
	var recoveryErr error
	if updateConfirmationPath == "" {
		recoveryErr = updater.RecoverDesktopUpdate("")
	}

	// 1. Run the local Gateway engine in the background
	cfg, configErr := loadDesktopConfig()

	srv := server.NewWithUpdateChannel(cfg, desktopVersion, desktopChannel)
	defer srv.Close()
	app := NewApp()
	app.updateConfirmationPath = updateConfirmationPath
	if recoveryErr != nil {
		if windowErr := wails.Run(desktopOptions(app, nil, recoveryErr, srv)); windowErr != nil {
			fmt.Printf("Desktop recovery error window failed: %v\n", windowErr)
		}
		return
	}
	if configErr != nil {
		fmt.Printf("Desktop configuration failed: %v\n", configErr)
		if windowErr := wails.Run(desktopOptions(app, nil, configErr, srv)); windowErr != nil {
			fmt.Printf("Desktop configuration error window failed: %v\n", windowErr)
		}
		return
	}
	gateway, err := startDesktopGateway(cfg.Port, srv.Handler(), desktopVersion)
	if err != nil {
		fmt.Printf("Gateway startup failed: %v\n", err)
		if windowErr := wails.Run(desktopOptions(app, nil, err, srv)); windowErr != nil {
			fmt.Printf("Desktop startup error window failed: %v\n", windowErr)
		}
		return
	}
	fmt.Printf("🚀 BOB Gemini Free local engine listening on %s\n", gateway.Endpoint())
	app.gatewayURL = gateway.Endpoint()

	// 2. Launch the branded native window, pointing it to our Gateway UI
	err = wails.Run(desktopOptions(app, gateway, nil, srv))

	if err != nil {
		println("Error:", err.Error())
	}
}

func desktopOptions(app *App, gateway *desktopGateway, startupErr error, gatewayApp *server.App) *options.App {
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
			app.startAutomaticDesktopUpdateChecks(ctx)
		},
		OnShutdown: func(ctx context.Context) {
			app.stopAutomaticDesktopUpdateChecks()
			if gatewayApp != nil {
				gatewayApp.Close()
			}
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
		app.checkDesktopUpdate(app.ctx, false)
	})
	help.AddText("Open GitHub Releases", nil, func(*menu.CallbackData) {
		if app.ctx != nil {
			runtime.BrowserOpenURL(app.ctx, updater.DesktopReleaseURL)
		}
	})
	return result
}
