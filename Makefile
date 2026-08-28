# ==============================================================================
# BOB Gemini Free - Makefile
# Break Ordinary Boundaries | Powered by ABCsteps (https://abcsteps.com)
# Author: Divyanshu Singh Chouhan (@div197)
# ==============================================================================

BINARY_NAME=bob-gemini-free
VERSION=v0.1.9
LDFLAGS=-s -w -X main.Version=$(VERSION)
WAILS_LDFLAGS=-X main.desktopVersion=$(VERSION) -X main.desktopChannel=stable
WAILS=go run github.com/wailsapp/wails/v2/cmd/wails@v2.15.0

.PHONY: all build web run test test-cover dist clean desktop desktop-preview-mac desktop-preview-windows desktop-preview-linux desktop-mac desktop-windows desktop-linux

all: build test

web:
	@echo "Syncing static web studio distribution to ./web..."
	@mkdir -p web
	@sed 's/__BOB_DESKTOP_VERSION__/$(VERSION)/g' internal/server/playground.html > web/index.html
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
	@echo "Building BOB Gemini Free Native Desktop App (Requires Go, CGO & host toolchain)..."
	cd cmd/desktop && $(WAILS) build -clean -ldflags="$(WAILS_LDFLAGS)"
	@echo "Desktop build complete! Check cmd/desktop/build/bin/"

desktop-preview-mac: web
	@echo "Building the free, ad-hoc-signed BOB Gemini Free macOS beta package..."
	scripts/package-wails-preview.sh

desktop-preview-windows: web
	@echo "Building the free, unsigned BOB Gemini Free Windows beta executable..."
	scripts/package-wails-windows-preview.sh

desktop-preview-linux: web
	@echo "Building the free BOB Gemini Free Linux beta package on a native Linux host..."
	scripts/package-wails-linux-preview.sh

desktop-mac: web
	@echo "Building the BOB Gemini Free macOS app (run on macOS; universal target)..."
	cd cmd/desktop && $(WAILS) build -clean -platform darwin/universal -ldflags="$(WAILS_LDFLAGS)"

desktop-windows: web
	@echo "Building the BOB Gemini Free Windows NSIS installer (run on Windows)..."
	cd cmd/desktop && $(WAILS) build -clean -platform windows/amd64 -nsis -webview2 download -ldflags="$(WAILS_LDFLAGS)"
	@test -f cmd/desktop/build/bin/bob-gemini-free-amd64-installer.exe || (echo "NSIS installer was not created; install makensis and run this target on the intended host." && exit 1)

desktop-linux: web
	@echo "Building the BOB Gemini Free Linux app (run on Linux with WebKitGTK installed)..."
	cd cmd/desktop && $(WAILS) build -clean -platform linux/amd64 -tags webkit2_41 -ldflags="$(WAILS_LDFLAGS)"
	@test -f cmd/desktop/build/bin/bob-gemini-free-linux-amd64 || (echo "Linux native binary was not created; run this target on a native Linux host with GTK/WebKitGTK." && exit 1)
