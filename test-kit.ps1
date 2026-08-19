# ==============================================================================
# BOB Gemini Free - Universal Diagnostic & Test Kit (Windows)
# Break Ordinary Boundaries | Powered by ABCsteps (https://abcsteps.com)
# Author: Divyanshu Singh Chouhan (@div197)
# ==============================================================================

param (
    [string]$TargetUrl = "http://127.0.0.1:9610",
    [string]$ApiKey = ""
)

if (Test-Path ".\bob-gemini-free.exe") {
    if ($ApiKey -ne "") {
        .\bob-gemini-free.exe --test --test-url $TargetUrl --test-key $ApiKey
    } else {
        .\bob-gemini-free.exe --test --test-url $TargetUrl
    }
} else {
    Write-Host "[*] Compiling BOB Gemini Free Windows binary..." -ForegroundColor Cyan
    $env:CGO_ENABLED = "0"
    go build -ldflags="-s -w" -o bob-gemini-free.exe .
    if ($ApiKey -ne "") {
        .\bob-gemini-free.exe --test --test-url $TargetUrl --test-key $ApiKey
    } else {
        .\bob-gemini-free.exe --test --test-url $TargetUrl
    }
}
