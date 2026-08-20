package service

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"text/template"
	"time"
)

// ServiceConfig holds parameters needed to render service unit files.
type ServiceConfig struct {
	ExePath   string
	Port      int
	LogPath   string
	ExtraArgs []string
}

const macOSPlistTemplate = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.abcsteps.bob-gemini-free</string>
    <key>ProgramArguments</key>
    <array>
        <string>{{.ExePath}}</string>
        <string>--port</string>
        <string>{{.Port}}</string>{{range .ExtraArgs}}
        <string>{{.}}</string>{{end}}
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
    <key>StandardOutPath</key>
    <string>{{.LogPath}}</string>
    <key>StandardErrorPath</key>
    <string>{{.LogPath}}</string>
</dict>
</plist>
`

const linuxSystemdTemplate = `[Unit]
Description=BOB Gemini Free AI Gateway Daemon
After=network.target

[Service]
ExecStart={{.ExePath}} --port {{.Port}} {{range .ExtraArgs}}'{{.}}' {{end}}
Restart=always
RestartSec=3

[Install]
WantedBy=default.target
`

const windowsBatchTemplate = `@echo off
start "" "{{.ExePath}}" --port {{.Port}} {{range .ExtraArgs}}"{{.}}" {{end}}
`

// Install registers and starts the native OS background daemon.
func Install(port int, extraArgs []string, logFn func(string, ...any)) error {
	if logFn == nil {
		logFn = func(string, ...any) {}
	}
	if port <= 0 {
		port = 9610
	}

	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}
	exePath, err = filepath.EvalSymlinks(exePath)
	if err != nil {
		return fmt.Errorf("failed to resolve symlink: %w", err)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %w", err)
	}

	logDir := filepath.Join(home, ".config", "bob-gemini-free", "logs")
	_ = os.MkdirAll(logDir, 0755)
	logPath := filepath.Join(logDir, "daemon.log")

	cfg := ServiceConfig{
		ExePath:   exePath,
		Port:      port,
		LogPath:   logPath,
		ExtraArgs: extraArgs,
	}

	switch runtime.GOOS {
	case "darwin":
		launchAgentsDir := filepath.Join(home, "Library", "LaunchAgents")
		if err := os.MkdirAll(launchAgentsDir, 0755); err != nil {
			return fmt.Errorf("failed to create LaunchAgents directory: %w", err)
		}
		plistPath := filepath.Join(launchAgentsDir, "com.abcsteps.bob-gemini-free.plist")

		var buf bytes.Buffer
		t := template.Must(template.New("plist").Parse(macOSPlistTemplate))
		if err := t.Execute(&buf, cfg); err != nil {
			return fmt.Errorf("failed to render plist template: %w", err)
		}

		if err := os.WriteFile(plistPath, buf.Bytes(), 0644); err != nil {
			return fmt.Errorf("failed to write plist file: %w", err)
		}

		logFn("[+] Created macOS LaunchAgent: %s", plistPath)
		_ = exec.Command("launchctl", "unload", plistPath).Run()
		if err := exec.Command("launchctl", "load", plistPath).Run(); err != nil {
			return fmt.Errorf("failed to load launchd service: %w", err)
		}
		logFn("[✔] Loaded launchd service com.abcsteps.bob-gemini-free")

	case "linux":
		systemdDir := filepath.Join(home, ".config", "systemd", "user")
		if err := os.MkdirAll(systemdDir, 0755); err != nil {
			return fmt.Errorf("failed to create systemd user directory: %w", err)
		}
		unitPath := filepath.Join(systemdDir, "bob-gemini-free.service")

		var buf bytes.Buffer
		t := template.Must(template.New("systemd").Parse(linuxSystemdTemplate))
		if err := t.Execute(&buf, cfg); err != nil {
			return fmt.Errorf("failed to render systemd template: %w", err)
		}

		if err := os.WriteFile(unitPath, buf.Bytes(), 0644); err != nil {
			return fmt.Errorf("failed to write systemd unit: %w", err)
		}

		logFn("[+] Created systemd user service: %s", unitPath)
		_ = exec.Command("systemctl", "--user", "daemon-reload").Run()
		if err := exec.Command("systemctl", "--user", "enable", "--now", "bob-gemini-free").Run(); err != nil {
			return fmt.Errorf("failed to enable systemd service: %w", err)
		}
		logFn("[✔] Enabled and started systemd user service bob-gemini-free")

	case "windows":
		appData := os.Getenv("APPDATA")
		if appData == "" {
			appData = filepath.Join(home, "AppData", "Roaming")
		}
		startupDir := filepath.Join(appData, "Microsoft", "Windows", "Start Menu", "Programs", "Startup")
		_ = os.MkdirAll(startupDir, 0755)
		batPath := filepath.Join(startupDir, "bob-gemini-free.bat")

		var buf bytes.Buffer
		t := template.Must(template.New("windows").Parse(windowsBatchTemplate))
		if err := t.Execute(&buf, cfg); err != nil {
			return fmt.Errorf("failed to render startup batch template: %w", err)
		}

		if err := os.WriteFile(batPath, buf.Bytes(), 0644); err != nil {
			return fmt.Errorf("failed to write startup batch: %w", err)
		}

		logFn("[+] Created Windows Startup shortcut: %s", batPath)
		_ = exec.Command("cmd", "/c", "start", "", batPath).Run()
		logFn("[✔] Registered startup launcher.")

	default:
		return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}

	logFn("[✔] Background service installed successfully!")
	logFn("[✔] Gateway running at http://127.0.0.1:%d", port)
	return nil
}

// Uninstall stops and removes the native OS background daemon.
func Uninstall(logFn func(string, ...any)) error {
	if logFn == nil {
		logFn = func(string, ...any) {}
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %w", err)
	}

	switch runtime.GOOS {
	case "darwin":
		plistPath := filepath.Join(home, "Library", "LaunchAgents", "com.abcsteps.bob-gemini-free.plist")
		_ = exec.Command("launchctl", "unload", plistPath).Run()
		if err := os.Remove(plistPath); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to delete plist: %w", err)
		}
		logFn("[✔] Uninstalled macOS LaunchAgent: %s", plistPath)

	case "linux":
		unitPath := filepath.Join(home, ".config", "systemd", "user", "bob-gemini-free.service")
		_ = exec.Command("systemctl", "--user", "disable", "--now", "bob-gemini-free").Run()
		if err := os.Remove(unitPath); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to delete systemd unit: %w", err)
		}
		_ = exec.Command("systemctl", "--user", "daemon-reload").Run()
		logFn("[✔] Uninstalled Linux systemd user service: %s", unitPath)

	case "windows":
		appData := os.Getenv("APPDATA")
		if appData == "" {
			appData = filepath.Join(home, "AppData", "Roaming")
		}
		batPath := filepath.Join(appData, "Microsoft", "Windows", "Start Menu", "Programs", "Startup", "bob-gemini-free.bat")
		if err := os.Remove(batPath); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to remove startup shortcut: %w", err)
		}
		logFn("[✔] Removed Windows Startup launcher: %s", batPath)

	default:
		return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}

	return nil
}

// Status checks whether the background service is installed and responding.
func Status(port int, logFn func(string, ...any)) error {
	if logFn == nil {
		logFn = func(string, ...any) {}
	}
	if port <= 0 {
		port = 9610
	}

	home, _ := os.UserHomeDir()
	isInstalled := false
	serviceFile := ""

	switch runtime.GOOS {
	case "darwin":
		serviceFile = filepath.Join(home, "Library", "LaunchAgents", "com.abcsteps.bob-gemini-free.plist")
		if _, err := os.Stat(serviceFile); err == nil {
			isInstalled = true
		}
	case "linux":
		serviceFile = filepath.Join(home, ".config", "systemd", "user", "bob-gemini-free.service")
		if _, err := os.Stat(serviceFile); err == nil {
			isInstalled = true
		}
	case "windows":
		appData := os.Getenv("APPDATA")
		if appData == "" {
			appData = filepath.Join(home, "AppData", "Roaming")
		}
		serviceFile = filepath.Join(appData, "Microsoft", "Windows", "Start Menu", "Programs", "Startup", "bob-gemini-free.bat")
		if _, err := os.Stat(serviceFile); err == nil {
			isInstalled = true
		}
	}

	if isInstalled {
		logFn("[✔] Service Definition: Installed (%s)", serviceFile)
	} else {
		logFn("[!] Service Definition: Not Installed (run: bob-gemini-free service install)")
	}

	// Check HTTP ping — accept both 200 (no API keys) and 401 (API keys configured)
	// as evidence the daemon is alive and listening.
	targetURL := fmt.Sprintf("http://127.0.0.1:%d/", port)
	client := &http.Client{Timeout: 1500 * time.Millisecond}
	resp, err := client.Get(targetURL)
	if err == nil {
		defer resp.Body.Close() // Always close body to prevent TCP socket leak
		if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusUnauthorized {
			logFn("[✔] Daemon Gateway Process: RUNNING (listening on http://127.0.0.1:%d)", port)
		} else {
			logFn("[!] Daemon Gateway Process: RESPONDING but unhealthy (HTTP %d) on port %d", resp.StatusCode, port)
		}
	} else {
		logFn("[!] Daemon Gateway Process: STOPPED / NOT RESPONDING on port %d", port)
	}

	return nil
}

// Start instructs the OS daemon manager to start the service.
func Start(logFn func(string, ...any)) error {
	if logFn == nil {
		logFn = func(string, ...any) {}
	}

	home, _ := os.UserHomeDir()
	switch runtime.GOOS {
	case "darwin":
		plistPath := filepath.Join(home, "Library", "LaunchAgents", "com.abcsteps.bob-gemini-free.plist")
		if err := exec.Command("launchctl", "load", plistPath).Run(); err != nil {
			_ = exec.Command("launchctl", "start", "com.abcsteps.bob-gemini-free").Run()
		}
		logFn("[✔] Triggered start on com.abcsteps.bob-gemini-free")
	case "linux":
		if err := exec.Command("systemctl", "--user", "start", "bob-gemini-free").Run(); err != nil {
			return err
		}
		logFn("[✔] Triggered start on systemd unit bob-gemini-free")
	case "windows":
		appData := os.Getenv("APPDATA")
		if appData == "" {
			appData = filepath.Join(home, "AppData", "Roaming")
		}
		batPath := filepath.Join(appData, "Microsoft", "Windows", "Start Menu", "Programs", "Startup", "bob-gemini-free.bat")
		_ = exec.Command("cmd", "/c", "start", "", batPath).Run()
		logFn("[✔] Launched Windows startup batch.")
	}
	return nil
}

// Stop instructs the OS daemon manager to stop the service.
func Stop(logFn func(string, ...any)) error {
	if logFn == nil {
		logFn = func(string, ...any) {}
	}

	home, _ := os.UserHomeDir()
	switch runtime.GOOS {
	case "darwin":
		plistPath := filepath.Join(home, "Library", "LaunchAgents", "com.abcsteps.bob-gemini-free.plist")
		_ = exec.Command("launchctl", "unload", plistPath).Run()
		logFn("[✔] Stopped macOS LaunchAgent com.abcsteps.bob-gemini-free")
	case "linux":
		_ = exec.Command("systemctl", "--user", "stop", "bob-gemini-free").Run()
		logFn("[✔] Stopped systemd unit bob-gemini-free")
	case "windows":
		_ = exec.Command("taskkill", "/IM", "bob-gemini-free.exe", "/F").Run()
		logFn("[✔] Terminated Windows process bob-gemini-free.exe")
	}
	return nil
}
