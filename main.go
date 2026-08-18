package main

import (
	"bufio"
	"context"
	"encoding/json"
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

	"github.com/div197/bob-gemini-free/internal/browser"
	"github.com/div197/bob-gemini-free/internal/config"
	"github.com/div197/bob-gemini-free/internal/diag"
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

func handleBrowserLogin() {
	fmt.Println("================================================================")
	fmt.Println("    BOB Gemini Free - 1-Click Interactive Login Window          ")
	fmt.Println("    Break Ordinary Boundaries | ABCsteps (https://abcsteps.com) ")
	fmt.Println("================================================================")
	fmt.Println()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	logFn := func(format string, args ...any) {
		fmt.Printf(format+"\n", args...)
	}

	extracted, err := browser.LaunchLoginSession(ctx, logFn)
	if err != nil {
		fmt.Printf("\n[!] Login failed: %v\n", err)
		os.Exit(1)
	}

	home, err := os.UserHomeDir()
	if err == nil {
		targetDir := filepath.Join(home, ".config", "bob-gemini-free")
		targetCookieFile := filepath.Join(targetDir, "cookie.txt")
		_ = gemini.SaveCookieFile(targetCookieFile, extracted.RawCookie)
	}

	_ = gemini.SaveCookieFile("./cookie.txt", extracted.RawCookie)

	fmt.Println()
	fmt.Println("================================================================")
	fmt.Printf("[✔] Verified %d session tokens!\n", len(extracted.Tokens))
	if extracted.SAPISID != "" {
		fmt.Printf("[✔] SAPISID detected: %s... (SAPISIDHASH active)\n", extracted.SAPISID[:min(6, len(extracted.SAPISID))])
	}
	fmt.Printf("[✔] Securely saved cookie to ./cookie.txt and ~/.config/bob-gemini-free/cookie.txt (mode 0600)\n")
	fmt.Println("[✔] Gemini Pro model (gemini-3.1-pro / gemini-pro) & Imagen 3 activated!")
	fmt.Println("================================================================")
	fmt.Println()
	fmt.Println("Start BOB Gemini Free with:")
	fmt.Println("  ./bob-gemini-free")
	fmt.Println()
	os.Exit(0)
}

func handleDiagnostics(targetURL, apiKey string) {
	fmt.Println("==================================================================")
	fmt.Println("    BOB Gemini Free - Automated Diagnostic Test Kit               ")
	fmt.Println("    Break Ordinary Boundaries | ABCsteps (https://abcsteps.com)   ")
	fmt.Println("==================================================================")
	fmt.Printf("Target Gateway URL: %s\n\n", targetURL)

	results := diag.RunDiagnostics(targetURL, apiKey)
	var passCount, failCount int

	for i, res := range results {
		if res.Passed {
			passCount++
			fmt.Printf("[%d/%d] [✔ PASS] %s (%v)\n", i+1, len(results), res.Name, res.Duration.Round(time.Millisecond))
			if res.Details != "" {
				fmt.Printf("      ↳ %s\n", res.Details)
			}
		} else {
			failCount++
			fmt.Printf("[%d/%d] [✘ FAIL] %s (%v)\n", i+1, len(results), res.Name, res.Duration.Round(time.Millisecond))
			fmt.Printf("      ↳ Error: %v\n", res.Error)
		}
	}

	fmt.Println()
	fmt.Println("==================================================================")
	if failCount == 0 {
		fmt.Printf("    ALL %d DIAGNOSTIC CHECKS PASSED (100%% SUCCESS)               \n", passCount)
	} else {
		fmt.Printf("    DIAGNOSTICS SUMMARY: %d PASSED, %d FAILED                    \n", passCount, failCount)
	}
	fmt.Println("==================================================================")
	fmt.Println()

	if failCount > 0 {
		os.Exit(1)
	}
	os.Exit(0)
}

func handleBenchmark(targetURL, apiKey string, concurrency, requests int) {
	fmt.Println("==================================================================")
	fmt.Println("    BOB Gemini Free - Performance & Stress Benchmark Runner       ")
	fmt.Println("    Break Ordinary Boundaries | ABCsteps (https://abcsteps.com)   ")
	fmt.Println("==================================================================")
	fmt.Printf("Target Gateway URL:   %s\n", targetURL)
	fmt.Printf("Concurrency Level:    %d workers\n", concurrency)
	fmt.Printf("Total Batch Requests: %d queries\n\n", requests)
	fmt.Println("[*] Benchmarking live throughput and latencies against upstream...")

	report := diag.RunBenchmark(targetURL, apiKey, concurrency, requests)

	fmt.Println()
	fmt.Println("------------------------------------------------------------------")
	fmt.Println("                    BENCHMARK RESULTS & METRICS                   ")
	fmt.Println("------------------------------------------------------------------")
	fmt.Printf("  • Completed Requests:   %d / %d (%.1f%% Success)\n", report.Successful, report.TotalRequests, float64(report.Successful)*100/float64(report.TotalRequests))
	fmt.Printf("  • Total Elapsed Time:   %v\n", report.TotalDuration.Round(time.Millisecond))
	fmt.Printf("  • Average Latency:      %v\n", report.AverageLatency.Round(time.Millisecond))
	fmt.Printf("  • Median Latency (P50): %v\n", report.P50Latency.Round(time.Millisecond))
	fmt.Printf("  • 90th Percentile (P90):%v\n", report.P90Latency.Round(time.Millisecond))
	fmt.Printf("  • 99th Percentile (P99):%v\n", report.P99Latency.Round(time.Millisecond))
	fmt.Printf("  • Request Throughput:   %.2f req/sec\n", report.RequestsPerSec)
	fmt.Printf("  • Token Throughput:     %.1f tokens/sec\n", report.TokensPerSecond)
	fmt.Println("==================================================================")
	fmt.Println()

	if report.Failed > 0 {
		os.Exit(1)
	}
	os.Exit(0)
}

func handleStatus(targetURL string) {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(targetURL + "/")
	if err != nil {
		fmt.Printf("❌ Failed to reach running BOB Gemini Free gateway at %s: %v\n", targetURL, err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	var data struct {
		Status              string   `json:"status"`
		Version             string   `json:"version"`
		Models              []string `json:"models"`
		RequestsServed      uint64   `json:"requests_served"`
		TokensProcessed     uint64   `json:"tokens_processed"`
		EstimatedSavingsUSD string   `json:"estimated_savings_usd"`
		UptimeSeconds       int      `json:"uptime_seconds"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		fmt.Printf("❌ Failed to parse status response: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("==================================================================")
	fmt.Println("    BOB Gemini Free - Live Gateway Telemetry & Status             ")
	fmt.Println("    Break Ordinary Boundaries | ABCsteps (https://abcsteps.com)   ")
	fmt.Println("==================================================================")
	fmt.Printf("  • Gateway Status:        %s (Version %s)\n", data.Status, data.Version)
	fmt.Printf("  • Target Gateway URL:    %s\n", targetURL)
	fmt.Printf("  • Server Uptime:         %d seconds (%.1f minutes)\n", data.UptimeSeconds, float64(data.UptimeSeconds)/60.0)
	fmt.Printf("  • Requests Served:       %d requests\n", data.RequestsServed)
	fmt.Printf("  • Tokens Processed:      %d tokens\n", data.TokensProcessed)
	fmt.Printf("  • Estimated USD Savings: %s (vs commercial cloud APIs)\n", data.EstimatedSavingsUSD)
	fmt.Printf("  • Active Models Loaded:  %d models\n", len(data.Models))
	fmt.Println("==================================================================")
	fmt.Println()
	os.Exit(0)
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
	hostFlag := flag.String("host", "", "Server host address")
	configFlag := flag.String("config", "", "Config file path")
	cookieFlag := flag.String("cookie-file", "", "Cookie file path")
	proxyFlag := flag.String("proxy", "", "HTTP proxy URL")
	impersonateFlag := flag.String("impersonate", "", "TLS impersonation profile")
	loginFlag := flag.Bool("login", false, "1-Click interactive browser sign-in window (zero copy-paste)")
	setupCookieFlag := flag.Bool("setup-cookie", false, "Automated Gemini cookie setup helper (paste prompt)")
	cookieStringFlag := flag.String("cookie-string", "", "Raw cookie string for non-interactive setup")
	testFlag := flag.Bool("test", false, "Run automated diagnostic test kit against a running gateway")
	statusFlag := flag.Bool("status", false, "Query live telemetry, uptime, and financial savings from a running gateway")
	testURLFlag := flag.String("test-url", "http://127.0.0.1:8081", "Target gateway URL for diagnostic tests")
	testKeyFlag := flag.String("test-key", "", "API key to use for diagnostic tests")
	benchFlag := flag.Bool("bench", false, "Run performance and throughput benchmark against a running gateway")
	benchConcurrencyFlag := flag.Int("bench-concurrency", 3, "Number of concurrent workers for benchmarking")
	benchRequestsFlag := flag.Int("bench-requests", 6, "Total number of requests for benchmarking")
	versionFlag := flag.Bool("version", false, "Show version")
	flag.Parse()

	if *versionFlag {
		fmt.Printf("BOB-Gemini-Free %s (Break Ordinary Boundaries by ABCsteps - https://abcsteps.com)\n", currentVersion)
		os.Exit(0)
	}

	if *statusFlag {
		handleStatus(*testURLFlag)
	}

	if *loginFlag {
		handleBrowserLogin()
	}

	if *testFlag {
		handleDiagnostics(*testURLFlag, *testKeyFlag)
	}

	if *benchFlag {
		handleBenchmark(*testURLFlag, *testKeyFlag, *benchConcurrencyFlag, *benchRequestsFlag)
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

	if *hostFlag != "" {
		cfg.Host = *hostFlag
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
