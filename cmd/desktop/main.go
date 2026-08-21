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

	"github.com/div197/bob-gemini-free/internal/config"
	"github.com/div197/bob-gemini-free/internal/server"
)

//go:embed all:frontend
var assets embed.FS

func main() {
	// 1. Run the local Gateway engine in the background
	cfg := config.Default()
	cfg.APIKeys = nil // Disable global config API keys for local desktop app

	srv := server.New(cfg, "v0.1.7")
	gateway, err := startDesktopGateway(cfg.Port, srv.Handler())
	if err != nil {
		fmt.Printf("Gateway startup failed: %v\n", err)
		return
	}
	fmt.Printf("🚀 Wails Internal Gateway listening on %s\n", gateway.Endpoint())
	app := NewApp()
	app.gatewayURL = gateway.Endpoint()

	// 2. Launch the Wails Native Window, pointing it to our Gateway UI
	err = wails.Run(&options.App{
		Title:  "BOB Gemini Free",
		Width:  1100,
		Height: 800,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		OnStartup: func(ctx context.Context) {
			app.startup(ctx)
			runtime.EventsEmit(ctx, "gateway-ready", gateway.Endpoint())
		},
		OnShutdown: func(ctx context.Context) {
			shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()
			if shutdownErr := gateway.Shutdown(shutdownCtx); shutdownErr != nil {
				fmt.Printf("Gateway shutdown failed: %v\n", shutdownErr)
			}
		},
		Bind: []interface{}{app},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
