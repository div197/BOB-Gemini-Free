# ==============================================================================
# BOB Gemini Free - Windows Setup Script
# Break Ordinary Boundaries | Powered by ABCsteps (https://abcsteps.com)
# Author: Divyanshu Singh Chouhan (@div197)
# ==============================================================================

Write-Host "================================================================" -ForegroundColor Cyan
Write-Host "    BOB Gemini Free - Break Ordinary Boundaries                " -ForegroundColor Green
Write-Host "    Powered by ABCsteps.com | Divyanshu Singh Chouhan (@div197) " -ForegroundColor Cyan
Write-Host "================================================================" -ForegroundColor Cyan
Write-Host ""

$AppName = "bob-gemini-free.exe"
$ConfigDir = "$HOME\.config\bob-gemini-free"

if (!(Test-Path $ConfigDir)) {
    New-Item -ItemType Directory -Force -Path $ConfigDir | Out-Null
}

if (!(Test-Path "$ConfigDir\config.json")) {
    if (Test-Path "config.example.json") {
        Copy-Item "config.example.json" "$ConfigDir\config.json"
        Write-Host "[✔] Created default configuration at $ConfigDir\config.json" -ForegroundColor Green
    }
}

$TargetBin = ".\$AppName"

# 1. Check if compiling from source
if ((Get-Command go -ErrorAction SilentlyContinue) -and (Test-Path "go.mod") -and (Select-String -Path "go.mod" -Pattern "bob-gemini-free" -Quiet -ErrorAction SilentlyContinue)) {
    Write-Host "[*] Go detected in source directory. Compiling $AppName from source..." -ForegroundColor Cyan
    $env:CGO_ENABLED = "0"
    go build -ldflags="-s -w" -o $AppName .
    Write-Host "[✔] Successfully built $AppName!" -ForegroundColor Green
} else {
    # 2. Download Pre-compiled release
    Write-Host "[*] Fetching pre-compiled Windows 64-bit binary..." -ForegroundColor Cyan
    $DownloadUrl = "https://github.com/div197/bob-gemini-free/releases/latest/download/bob-gemini-free-windows-amd64.exe"
    
    # If not in source tree, install to a dedicated folder in user profile to keep it clean
    if (!(Test-Path "main.go")) {
        $InstallDir = "$HOME\bob-gemini-free"
        if (!(Test-Path $InstallDir)) {
            New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
        }
        $TargetBin = "$InstallDir\$AppName"
    }

    try {
        Invoke-WebRequest -Uri $DownloadUrl -OutFile $TargetBin
        Write-Host "[✔] Standalone binary downloaded successfully to $TargetBin" -ForegroundColor Green
    } catch {
        Write-Host "[!] Pre-compiled binary not yet available on GitHub Releases." -ForegroundColor Yellow
        Write-Host "[*] Please install Go (https://go.dev/dl/) to build locally." -ForegroundColor Cyan
        exit 1
    }
}

Write-Host ""
Write-Host "================================================================" -ForegroundColor Green
Write-Host "    INSTALLATION COMPLETE! 🚀" -ForegroundColor Green
Write-Host "================================================================" -ForegroundColor Green
Write-Host ""
Write-Host "To launch the gateway and open the Web Studio, run:" -ForegroundColor White
Write-Host "  $TargetBin --port 9610" -ForegroundColor Cyan
Write-Host ""
Write-Host "API Base URL: http://127.0.0.1:9610/v1" -ForegroundColor DarkGray
Write-Host "UI Dashboard: http://127.0.0.1:9610/playground" -ForegroundColor DarkGray
Write-Host ""
