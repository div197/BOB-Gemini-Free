package main

import (
	"log"

	"github.com/div197/bob-gemini-free/internal/config"
)

// loadDesktopConfig keeps the packaged app convenient for an individual user
// without widening its network boundary. The desktop process may discover the
// user's own config/cookie files, but it must never become a remotely reachable
// gateway or accept a configured API-key surface intended for server mode.
func loadDesktopConfig() config.Config {
	configPath := config.Find()
	cfg, err := config.Load(configPath)
	if err != nil {
		if configPath != "" {
			log.Printf("desktop config ignored from %s: %v", configPath, err)
		}
		// Preserve environment-variable and per-user cookie discovery even when
		// an optional config file is malformed.
		cfg, _ = config.Load("")
	}

	cfg.Host = "127.0.0.1"
	cfg.APIKeys = nil
	cfg.AllowedOrigins = nil
	return cfg
}
