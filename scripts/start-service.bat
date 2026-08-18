@echo off
REM ==============================================================================
REM BOB Gemini Free - Windows Background Runner
REM Break Ordinary Boundaries | Powered by ABCsteps (https://abcsteps.com)
REM ==============================================================================

if exist bob-gemini-free.exe (
    start "" /b bob-gemini-free.exe --port 8081
    echo [✔] BOB Gemini Free started in background on http://127.0.0.1:8081
) else (
    echo [*] Compiling binary first...
    set CGO_ENABLED=0
    go build -ldflags="-s -w" -o bob-gemini-free.exe .
    start "" /b bob-gemini-free.exe --port 8081
    echo [✔] BOB Gemini Free started in background on http://127.0.0.1:8081
)
