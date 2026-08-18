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
        Write-Host "[+] Created default configuration at $ConfigDir\config.json" -ForegroundColor Green
    }
}

if (Test-Path ".\$AppName") {
    Write-Host "[✔] Existing $AppName binary found locally." -ForegroundColor Green
    Write-Host "Start with: .\$AppName --port 8081" -ForegroundColor Yellow
} elseif (Get-Command go -ErrorAction SilentlyContinue) {
    Write-Host "[*] Go detected. Compiling $AppName from source..." -ForegroundColor Cyan
    $env:CGO_ENABLED = "0"
    go build -ldflags="-s -w" -o $AppName .
    Write-Host "[✔] Successfully built $AppName!" -ForegroundColor Green
    Write-Host ""
    Write-Host "Start the server by running:" -ForegroundColor Yellow
    Write-Host "  .\$AppName --port 8081" -ForegroundColor White
} elseif (Get-Command docker -ErrorAction SilentlyContinue) {
    Write-Host "[*] Building Docker container..." -ForegroundColor Cyan
    docker build -t bob-gemini-free .
    Write-Host "[✔] Docker container built!" -ForegroundColor Green
    Write-Host "Run with: docker run -d --name bob-gemini-free -p 8081:8081 bob-gemini-free" -ForegroundColor Yellow
} else {
    Write-Host "[*] No Go or Docker detected. Fetching pre-compiled Windows 64-bit binary..." -ForegroundColor Cyan
    $DownloadUrl = "https://github.com/div197/bob-gemini-free/releases/latest/download/bob-gemini-free-windows-amd64.exe"
    try {
        Invoke-WebRequest -Uri $DownloadUrl -OutFile $AppName
        Write-Host "[✔] Standalone binary downloaded successfully!" -ForegroundColor Green
        Write-Host "Start with: .\$AppName --port 8081" -ForegroundColor Yellow
    } catch {
        Write-Host "[!] Pre-compiled binary not yet available on GitHub Releases." -ForegroundColor Yellow
        Write-Host "[*] Please install Go (https://go.dev/dl/) or download a release binary." -ForegroundColor Cyan
    }
}

Write-Host ""
Write-Host "Base URL: http://127.0.0.1:8081/v1" -ForegroundColor Cyan
Write-Host "Visit ABCsteps: https://abcsteps.com/" -ForegroundColor Cyan
