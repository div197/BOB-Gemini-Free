# ==============================================================================
# BOB Gemini Free - Makefile
# Break Ordinary Boundaries | Powered by ABCsteps (https://abcsteps.com)
# Author: Divyanshu Singh Chouhan (@div197)
# ==============================================================================

BINARY_NAME=bob-gemini-free
VERSION=v0.1.1
LDFLAGS=-s -w -X main.Version=$(VERSION)

.PHONY: all build run test test-cover dist clean

all: build test

build:
	@echo "Building $(BINARY_NAME) $(VERSION)..."
	CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o $(BINARY_NAME) .
	@echo "Build complete: ./$(BINARY_NAME)"

run: build
	./$(BINARY_NAME) --port 8081

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
	@echo "Distribution builds ready in ./dist:"
	@ls -la dist/

clean:
	rm -f $(BINARY_NAME) $(BINARY_NAME).exe coverage.txt
	rm -rf dist/
