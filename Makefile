# ==============================================================================
# BOB Gemini Free - Makefile
# Break Ordinary Boundaries | Powered by ABCsteps (https://abcsteps.com)
# Author: Divyanshu Singh Chouhan (@div197)
# ==============================================================================

BINARY_NAME=bob-gemini-free
VERSION=v0.1.7
LDFLAGS=-s -w -X main.Version=$(VERSION)
WAILS=go run github.com/wailsapp/wails/v2/cmd/wails@v2.15.0

.PHONY: all build web run test test-cover dist clean desktop desktop-mac desktop-windows desktop-linux

all: build test

web:
	@echo "Syncing static web studio distribution to ./web..."
	@mkdir -p web
	@cp internal/server/playground.html web/index.html
	@echo "Web distribution bundle ready in ./web"

build: web
	@echo "Building $(BINARY_NAME) $(VERSION)..."
	CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o $(BINARY_NAME) .
	@echo "Build complete: ./$(BINARY_NAME)"

run: build
	./$(BINARY_NAME) --port 9610

test:
	go test -count=1 -v ./...

test-cover:
	go test -count=1 -coverprofile=coverage.txt -covermode=atomic ./...
	go tool cover -func=coverage.txt

dist: clean
	@echo "Cross-compiling $(BINARY_NAME) $(VERSION) for all platforms..."
	@mkdir -p dist
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o dist/$(BINARY_NAME)-darwin-arm64 .
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o dist/$(BINARY_NAME)-darwin-amd64 .
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o dist/$(BINARY_NAME)-linux-amd64 .
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o dist/$(BINARY_NAME)-linux-arm64 .
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o dist/$(BINARY_NAME)-windows-amd64.exe .
	GOOS=windows GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o dist/$(BINARY_NAME)-windows-arm64.exe .
	@echo "Distribution builds ready in ./dist:"
	@ls -la dist/

clean:
	rm -f $(BINARY_NAME) $(BINARY_NAME).exe coverage.txt
	rm -rf dist/
	rm -rf cmd/desktop/build/bin/

desktop: web
	@echo "Building Wails Native Desktop App (Requires Go, CGO & host toolchain)..."
	cd cmd/desktop && $(WAILS) build -clean
	@echo "Desktop build complete! Check cmd/desktop/build/bin/"

desktop-mac: web
	@echo "Building the macOS Wails app (run on macOS; universal target)..."
	cd cmd/desktop && $(WAILS) build -clean -platform darwin/universal

desktop-windows: web
	@echo "Building the Windows Wails NSIS installer (run on Windows)..."
	cd cmd/desktop && $(WAILS) build -clean -platform windows/amd64 -nsis -webview2 download

desktop-linux: web
	@echo "Building the Linux Wails app (run on Linux with WebKitGTK installed)..."
	cd cmd/desktop && $(WAILS) build -clean -platform linux/amd64 -tags webkit2_41
