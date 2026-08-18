package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"runtime/debug"
	"strings"
	"syscall"
	"time"

	"github.com/div197/bob-gemini-free/internal/config"
	"github.com/div197/bob-gemini-free/internal/gemini"
	"github.com/div197/bob-gemini-free/internal/models"
	"github.com/div197/bob-gemini-free/internal/server"
)

var Version = "dev"

func resolveVersion() string {
	if Version != "" && Version != "dev" {
		return Version
	}
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}
	return Version
}

func handleCookieSetup(rawInput string) {
	fmt.Println("================================================================")
	fmt.Println("    BOB Gemini Free - Automated Cookie Setup Helper             ")
	fmt.Println("    Break Ordinary Boundaries | ABCsteps (https://abcsteps.com) ")
	fmt.Println("================================================================")
	fmt.Println()

	var cookieInput string
	if rawInput != "" {
		cookieInput = rawInput
	} else {
		fmt.Println("Follow these quick steps to get your Gemini session cookie:")
		fmt.Println("  1. Open Chrome and go to: https://gemini.google.com")
		fmt.Println("  2. Open DevTools (F12) -> Application -> Cookies -> https://gemini.google.com")
		fmt.Println("  3. Copy your cookies or the raw Cookie header string.")
		fmt.Println()
		fmt.Print("Paste your cookie string here and press ENTER:\n> ")
		scanner := bufio.NewScanner(os.Stdin)
		if scanner.Scan() {
			cookieInput = scanner.Text()
		}
	}

	cookieInput = strings.TrimSpace(cookieInput)
	if cookieInput == "" {
		fmt.Println("\n[!] Error: No cookie input provided. Aborting.")
		os.Exit(1)
	}

	extracted, err := gemini.ExtractCookies(cookieInput)
	if err != nil {
		fmt.Printf("\n[!] Error parsing cookie: %v\n", err)
		os.Exit(1)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("\n[!] Error finding home directory: %v\n", err)
		os.Exit(1)
	}

	targetDir := filepath.Join(home, ".config", "bob-gemini-free")
	targetCookieFile := filepath.Join(targetDir, "cookie.txt")

	if err := gemini.SaveCookieFile(targetCookieFile, extracted.RawCookie); err != nil {
		fmt.Printf("\n[!] Error saving cookie file: %v\n", err)
		os.Exit(1)
	}

	fmt.Println()
	fmt.Printf("[✔] Verified %d session tokens!\n", len(extracted.Tokens))
	if extracted.SAPISID != "" {
		fmt.Printf("[✔] SAPISID detected: %s... (SAPISIDHASH active)\n", extracted.SAPISID[:min(6, len(extracted.SAPISID))])
	}
	fmt.Printf("[✔] Securely saved cookies to: %s (mode 0600)\n", targetCookieFile)
	fmt.Println("[✔] Gemini Pro model (gemini-3.1-pro / gemini-pro) routing activated!")
	fmt.Println()
	fmt.Println("Start BOB Gemini Free with:")
	fmt.Printf("  ./bob-gemini-free --cookie-file %s\n", targetCookieFile)
	fmt.Println()
	os.Exit(0)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func main() {
	currentVersion := resolveVersion()

	portFlag := flag.Int("port", 0, "Server port")
	configFlag := flag.String("config", "", "Config file path")
	cookieFlag := flag.String("cookie-file", "", "Cookie file path")
	proxyFlag := flag.String("proxy", "", "HTTP proxy URL")
	impersonateFlag := flag.String("impersonate", "", "TLS impersonation profile")
	setupCookieFlag := flag.Bool("setup-cookie", false, "Automated Gemini cookie setup helper")
	cookieStringFlag := flag.String("cookie-string", "", "Raw cookie string for non-interactive setup")
	versionFlag := flag.Bool("version", false, "Show version")
	flag.Parse()

	if *versionFlag {
		fmt.Printf("BOB-Gemini-Free %s (Break Ordinary Boundaries by ABCsteps - https://abcsteps.com)\n", currentVersion)
		os.Exit(0)
	}

	if *setupCookieFlag || *cookieStringFlag != "" {
		handleCookieSetup(*cookieStringFlag)
	}

	configPath := *configFlag
	if configPath == "" {
		configPath = os.Getenv("BOB_GEMINI_FREE_CONFIG")
	}
	if configPath == "" {
		configPath = config.Find()
	}

	cfg, err := config.Load(configPath)
	if err != nil && configPath != "" {
		log.Printf("Warning: failed to load config from %s: %v", configPath, err)
	}

	if *portFlag != 0 {
		cfg.Port = *portFlag
	}
	if *cookieFlag != "" {
		cfg.CookieFile = *cookieFlag
	}
	if *proxyFlag != "" {
		cfg.Proxy = *proxyFlag
	}
	if *impersonateFlag != "" {
		cfg.Impersonate = *impersonateFlag
	}

	app := server.New(cfg, currentVersion)

	modelKeys := make([]string, 0, len(models.MODELS))
	for k := range models.MODELS {
		modelKeys = append(modelKeys, k)
	}

	cookieStatus := "none (anonymous free tier)"
	if cfg.CookieFile != "" {
		cookieStatus = fmt.Sprintf("yes (%s)", cfg.CookieFile)
	}

	proxyStatus := cfg.Proxy
	if proxyStatus == "" {
		proxyStatus = "system env"
	}

	impersonateStatus := cfg.Impersonate
	if impersonateStatus == "" {
		impersonateStatus = "none (stdlib)"
	}

	fmt.Printf("BOB Gemini Free %s - Break Ordinary Boundaries\n", currentVersion)
	fmt.Printf("  Powered by:  ABCsteps (https://abcsteps.com)\n")
	fmt.Printf("  Author:      Divyanshu Singh Chouhan (@div197)\n")
	fmt.Printf("  Listening:   http://%s:%d\n", cfg.Host, cfg.Port)
	fmt.Printf("  Base URL:    http://localhost:%d/v1\n", cfg.Port)
	fmt.Printf("  Models:      %s\n", strings.Join(modelKeys, ", "))
	fmt.Printf("  Cookie:      %s\n", cookieStatus)
	fmt.Printf("  Proxy:       %s\n", proxyStatus)
	fmt.Printf("  Impersonate: %s\n", impersonateStatus)
	fmt.Println()

	srv := &http.Server{
		Addr:              fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Handler:           app.Handler(),
		ReadHeaderTimeout: 10 * time.Second,
		WriteTimeout:      0,
	}

	stopCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	<-stopCtx.Done()
	log.Println("Shutting down server gracefully...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server forced shutdown: %v", err)
	}

	log.Println("Stopped.")
}
