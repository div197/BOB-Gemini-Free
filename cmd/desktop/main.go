package main

import (
	"context"
	"embed"
	"fmt"
	"time"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/runtime"

	"github.com/div197/bob-gemini-free/internal/server"
)

//go:embed all:frontend
var assets embed.FS

func main() {
	// 1. Run the local Gateway engine in the background
	cfg := loadDesktopConfig()

	srv := server.New(cfg, "v0.1.7")
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
