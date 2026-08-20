package main

import (
	"context"
	"embed"
	"fmt"
	"net/http"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"

	"github.com/div197/bob-gemini-free/internal/config"
	"github.com/div197/bob-gemini-free/internal/server"
)

//go:embed all:frontend
var assets embed.FS

func main() {
	// 1. Run the local Gateway engine in the background
	cfg := config.Default()
	cfg.Port = 9610
	cfg.APIKeys = nil // Disable global config API keys for local desktop app

	srv := server.New(cfg, "v0.1.7")
	go func() {
		fmt.Println("🚀 Wails Internal Gateway listening on :9610")
		if err := http.ListenAndServe(":9610", srv.Handler()); err != nil {
			fmt.Printf("Gateway crashed: %v\n", err)
		}
	}()

	// 2. Launch the Wails Native Window, pointing it to our Gateway UI
	err := wails.Run(&options.App{
		Title:  "BOB Gemini Free",
		Width:  1100,
		Height: 800,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		OnStartup: func(ctx context.Context) {},
		Bind:      []interface{}{},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
