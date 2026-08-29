package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"slices"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/div197/bob-gemini-free/internal/browser"
	"github.com/div197/bob-gemini-free/internal/config"
	"github.com/div197/bob-gemini-free/internal/diag"
	"github.com/div197/bob-gemini-free/internal/gemini"
	"github.com/div197/bob-gemini-free/internal/models"
	"github.com/div197/bob-gemini-free/internal/server"
	"github.com/div197/bob-gemini-free/internal/service"
	"github.com/div197/bob-gemini-free/internal/updater"
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

	var savedPaths []string
	var saveErrors []string
	if home, err := os.UserHomeDir(); err == nil {
		targetCookieFile := filepath.Join(home, ".config", "bob-gemini-free", "cookie.txt")
		if err := gemini.SaveCookieFile(targetCookieFile, extracted.RawCookie); err != nil {
			saveErrors = append(saveErrors, fmt.Sprintf("%s: %v", targetCookieFile, err))
		} else {
			savedPaths = append(savedPaths, targetCookieFile)
		}
	} else {
		saveErrors = append(saveErrors, fmt.Sprintf("home directory: %v", err))
	}
	if err := gemini.SaveCookieFile("./cookie.txt", extracted.RawCookie); err != nil {
		saveErrors = append(saveErrors, fmt.Sprintf("./cookie.txt: %v", err))
	} else {
		savedPaths = append(savedPaths, "./cookie.txt")
	}
	if len(savedPaths) == 0 {
		fmt.Printf("\n[!] Login succeeded but no secure cookie file could be saved: %s\n", strings.Join(saveErrors, "; "))
		os.Exit(1)
	}

	fmt.Println()
	fmt.Println("================================================================")
	fmt.Printf("[✔] Verified %d session tokens!\n", len(extracted.Tokens))
	if extracted.SAPISID != "" {
		fmt.Println("[✔] SAPISID detected (SAPISIDHASH support available; value withheld)")
	}
	fmt.Printf("[✔] Securely saved cookie file(s) (mode 0600): %s\n", strings.Join(savedPaths, ", "))
	fmt.Println("[i] Provider model, vision, and image capabilities remain session-dependent; verify them with a real request.")
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

	var passCount, failCount int
	_ = diag.RunDiagnosticsWithProgress(targetURL, apiKey, func(idx, total int, res diag.TestResult) {
		if res.Passed {
			passCount++
			fmt.Printf("[%d/%d] [✔ PASS] %s (%v)\n", idx, total, res.Name, res.Duration.Round(time.Millisecond))
			if res.Details != "" {
				fmt.Printf("      ↳ %s\n", res.Details)
			}
		} else {
			failCount++
			fmt.Printf("[%d/%d] [✘ FAIL] %s (%v)\n", idx, total, res.Name, res.Duration.Round(time.Millisecond))
			fmt.Printf("      ↳ Error: %v\n", res.Error)
		}
	})

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
	fmt.Printf("Requested Workload:    %d workers × %d queries\n\n", concurrency, requests)
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
	fmt.Printf("  • Effective Workload:   %d workers × %d queries\n", report.Concurrency, report.TotalRequests)
	if report.TokenCountsMeasured {
		fmt.Printf("  • Token Throughput:     %.1f tokens/sec (provider-reported usage)\n", report.TokensPerSecond)
	} else {
		fmt.Println("  • Token Throughput:     unavailable (usage was not reported for every successful response)")
	}
	fmt.Println("==================================================================")
	fmt.Println()

	if report.Failed > 0 {
		os.Exit(1)
	}
	os.Exit(0)
}

func handleStatus(targetURL, apiKey string) {
	client := &http.Client{Timeout: 5 * time.Second}
	req, err := newStatusRequest(targetURL, apiKey)
	if err != nil {
		fmt.Printf("❌ Invalid gateway status URL %q: %v\n", targetURL, err)
		os.Exit(1)
	}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("❌ Failed to reach running BOB Gemini Free gateway at %s: %v\n", targetURL, err)
		os.Exit(1)
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		fmt.Printf("❌ Gateway status endpoint returned HTTP %d\n", resp.StatusCode)
		os.Exit(1)
	}

	var data struct {
		Status              string   `json:"status"`
		Version             string   `json:"version"`
		Models              []string `json:"models"`
		RequestsServed      uint64   `json:"requests_served"`
		TokensProcessed     uint64   `json:"tokens_processed"`
		EstimatedSavingsUSD string   `json:"estimated_savings_usd"`
		UptimeSeconds       int      `json:"uptime_seconds"`
		PoolSessionsTotal   int      `json:"pool_sessions_total"`
		PoolSessionsHealthy int      `json:"pool_sessions_healthy"`
	}

	if err := json.NewDecoder(io.LimitReader(resp.Body, 64<<10)).Decode(&data); err != nil {
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
	fmt.Printf("  • Cookie Pool Sessions:  %d total, %d healthy\n", data.PoolSessionsTotal, data.PoolSessionsHealthy)
	fmt.Println("==================================================================")
	fmt.Println()
	os.Exit(0)
}

func newStatusRequest(targetURL, apiKey string) (*http.Request, error) {
	req, err := http.NewRequest(http.MethodGet, strings.TrimRight(targetURL, "/")+"/", nil)
	if err != nil {
		return nil, err
	}
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}
	return req, nil
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
		fmt.Println("[✔] SAPISID detected (SAPISIDHASH support available; value withheld)")
	}
	fmt.Printf("[✔] Securely saved cookies to: %s (mode 0600)\n", targetCookieFile)
	fmt.Println("[i] Provider model, vision, and image capabilities remain session-dependent; verify them with a real request.")
	fmt.Println()
	fmt.Println("Start BOB Gemini Free with:")
	fmt.Printf("  ./bob-gemini-free --cookie-file %s\n", targetCookieFile)
	fmt.Println()
	os.Exit(0)
}

type studioLauncher func(context.Context, int, func(string, ...any)) error

type studioFallback func(string) error

func launchStudioOrFallback(ctx context.Context, port int, launch studioLauncher, fallback studioFallback) error {
	if launch == nil {
		return fmt.Errorf("desktop Studio launcher is unavailable")
	}
	if err := launch(ctx, port, log.Printf); err == nil {
		return nil
	} else {
		studioURL := fmt.Sprintf("http://localhost:%d/playground", port)
		if fallback == nil {
			return fmt.Errorf("embedded Studio launch failed: %w; browser fallback is unavailable", err)
		}
		if fallbackErr := fallback(studioURL); fallbackErr != nil {
			return fmt.Errorf("embedded Studio launch failed: %w; browser fallback failed: %v", err, fallbackErr)
		}
	}
	return nil
}

func openStudioFallback(studioURL string) error {
	var command *exec.Cmd
	switch runtime.GOOS {
	case "linux":
		command = exec.Command("xdg-open", studioURL)
	case "windows":
		command = exec.Command("rundll32", "url.dll,FileProtocolHandler", studioURL)
	case "darwin":
		command = exec.Command("open", studioURL)
	default:
		return fmt.Errorf("default browser fallback is unsupported on %s", runtime.GOOS)
	}
	if err := command.Start(); err != nil {
		return fmt.Errorf("open default browser: %w", err)
	}
	log.Printf("🚀 Opened Studio in default browser tab: %s", studioURL)
	return nil
}

func handleUpdate(currentVersion string) {
	fmt.Println("==================================================================")
	fmt.Println("    BOB Gemini Free - In-Place Auto-Updater                       ")
	fmt.Println("    Break Ordinary Boundaries | ABCsteps (https://abcsteps.com)   ")
	fmt.Println("==================================================================")
	fmt.Printf("Current Version: %s\n\n", currentVersion)

	logFn := func(format string, args ...any) {
		fmt.Printf(format+"\n", args...)
	}

	if err := updater.SelfUpdate(currentVersion, logFn); err != nil {
		fmt.Printf("\n[!] Update error: %v\n", err)
		os.Exit(1)
	}
	os.Exit(0)
}

func handleService(action string, port int, extraArgs []string) {
	fmt.Println("==================================================================")
	fmt.Println("    BOB Gemini Free - Native OS Service Daemon Manager            ")
	fmt.Println("    Break Ordinary Boundaries | ABCsteps (https://abcsteps.com)   ")
	fmt.Println("==================================================================")

	logFn := func(format string, args ...any) {
		fmt.Printf(format+"\n", args...)
	}

	if port <= 0 {
		port = 9610
	}

	var err error
	switch strings.ToLower(action) {
	case "install", "enable":
		err = service.Install(port, extraArgs, logFn)
	case "uninstall", "remove", "disable":
		err = service.Uninstall(logFn)
	case "start":
		err = service.Start(logFn)
	case "stop":
		err = service.Stop(logFn)
	case "status", "check":
		err = service.Status(port, logFn)
	default:
		fmt.Printf("Unknown service action: %q\n\n", action)
		fmt.Println("Available service commands:")
		fmt.Println("  ./bob-gemini-free service install [--port 9610]")
		fmt.Println("  ./bob-gemini-free service start")
		fmt.Println("  ./bob-gemini-free service stop")
		fmt.Println("  ./bob-gemini-free service status")
		fmt.Println("  ./bob-gemini-free service uninstall")
		os.Exit(1)
	}

	if err != nil {
		fmt.Printf("\n[!] Service operation failed: %v\n", err)
		os.Exit(1)
	}
	os.Exit(0)
}

func main() {
	currentVersion := resolveVersion()

	// Direct subcommand routing for "service" and "update"
	if len(os.Args) > 1 {
		firstArg := strings.ToLower(os.Args[1])
		if firstArg == "update" || firstArg == "--update" {
			handleUpdate(currentVersion)
		}
		if len(os.Args) > 1 && os.Args[1] == "service" {
			action := "status"
			if len(os.Args) > 2 {
				action = os.Args[2]
			}
			port := 9610
			var extraArgs []string
			for i := 2; i < len(os.Args); i++ {
				if os.Args[i] == "--port" && i+1 < len(os.Args) {
					fmt.Sscanf(os.Args[i+1], "%d", &port)
					i++ // skip the value
				} else if os.Args[i] != action {
					extraArgs = append(extraArgs, os.Args[i])
				}
			}
			handleService(action, port, extraArgs)
		}
	}

	portFlag := flag.Int("port", 0, "Server port")
	hostFlag := flag.String("host", "", "Server host address")
	configFlag := flag.String("config", "", "Config file path")
	cookieFlag := flag.String("cookie-file", "", "Cookie file path")
	proxyFlag := flag.String("proxy", "", "HTTP proxy URL")
	impersonateFlag := flag.String("impersonate", "", "TLS impersonation profile")
	loginFlag := flag.Bool("login", false, "1-Click interactive browser sign-in window (zero copy-paste)")
	headlessFlag := flag.Bool("headless", false, "Run without opening a native window")
	setupCookieFlag := flag.Bool("setup-cookie", false, "Automated Gemini cookie setup helper (paste prompt)")
	cookieStringFlag := flag.String("cookie-string", "", "Raw cookie string for non-interactive setup")
	testFlag := flag.Bool("test", false, "Run automated diagnostic test kit against a running gateway")
	statusFlag := flag.Bool("status", false, "Query live telemetry, uptime, and financial savings from a running gateway")
	testURLFlag := flag.String("test-url", "http://127.0.0.1:9610", "Target gateway URL for diagnostic tests")
	testKeyFlag := flag.String("test-key", "", "API key to use for diagnostic tests")
	benchFlag := flag.Bool("bench", false, "Run performance and throughput benchmark against a running gateway")
	benchConcurrencyFlag := flag.Int("bench-concurrency", 3, "Number of concurrent workers for benchmarking")
	benchRequestsFlag := flag.Int("bench-requests", 6, "Total number of requests for benchmarking")
	updateFlag := flag.Bool("update", false, "Check for updates and automatically update bob-gemini-free to the latest release")
	serviceFlag := flag.String("service", "", "Manage background service: install | uninstall | start | stop | status")
	versionFlag := flag.Bool("version", false, "Show version")
	flag.Parse()

	if *versionFlag {
		fmt.Printf("BOB-Gemini-Free %s (Break Ordinary Boundaries by ABCsteps - https://abcsteps.com)\n", currentVersion)
		os.Exit(0)
	}

	if *updateFlag {
		handleUpdate(currentVersion)
	}

	if *serviceFlag != "" {
		port := 9610
		if *portFlag != 0 {
			port = *portFlag
		}
		handleService(*serviceFlag, port, nil)
	}

	var configPath string
	if configFlag != nil && *configFlag != "" {
		if *configFlag == "none" {
			configPath = "" // Explicitly bypass auto-discovery
		} else {
			configPath = *configFlag
		}
	} else {
		configPath = config.Find()
	}

	cfg, err := config.Load(configPath)
	if err != nil && configPath != "" {
		log.Fatalf("failed to load config from %s: %v", configPath, err)
	}

	// Auto-fallback test key if none specified and API keys are configured
	activeTestKey := *testKeyFlag
	if activeTestKey == "" && len(cfg.APIKeys) > 0 {
		activeTestKey = cfg.APIKeys[0]
	}

	if *statusFlag {
		handleStatus(*testURLFlag, activeTestKey)
	}

	if *loginFlag {
		handleBrowserLogin()
	}

	if *testFlag {
		handleDiagnostics(*testURLFlag, activeTestKey)
	}

	if *benchFlag {
		handleBenchmark(*testURLFlag, activeTestKey, *benchConcurrencyFlag, *benchRequestsFlag)
	}

	if *setupCookieFlag || *cookieStringFlag != "" {
		handleCookieSetup(*cookieStringFlag)
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
	defer app.Close()

	modelCatalog := models.GetAllModels()
	modelKeys := make([]string, 0, len(modelCatalog))
	for k := range modelCatalog {
		modelKeys = append(modelKeys, k)
	}
	slices.Sort(modelKeys)

	cookieStatus := "none (no cookie configured; upstream access not guaranteed)"
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
	fmt.Printf("  Playground:  http://localhost:%d/playground\n", cfg.Port)
	fmt.Printf("  Models:      %s\n", strings.Join(modelKeys, ", "))
	fmt.Printf("  Cookie:      %s\n", cookieStatus)
	fmt.Printf("  Proxy:       %s\n", proxyStatus)
	fmt.Printf("  Impersonate: %s\n", impersonateStatus)
	fmt.Println()

	srv := &http.Server{
		Addr:              net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port)),
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

	go func() {
		time.Sleep(500 * time.Millisecond) // Give server time to bind
		if !*headlessFlag {
			if err := launchStudioOrFallback(context.Background(), cfg.Port, browser.LaunchStudioWindow, openStudioFallback); err != nil {
				log.Printf("Studio startup failed: %v", err)
			}
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
