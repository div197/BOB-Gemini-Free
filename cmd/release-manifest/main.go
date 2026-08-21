package main

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/hex"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/div197/bob-gemini-free/internal/updater"
)

func main() {
	directory := flag.String("dir", "release-assets", "directory containing release assets")
	flag.Parse()

	privateKey, err := decodePrivateKey(os.Getenv("BOB_GEMINI_FREE_UPDATE_PRIVATE_KEY"))
	if err != nil {
		fail(err)
	}
	manifest, signature, err := updater.CreateSignedManifest(*directory, privateKey)
	if err != nil {
		fail(err)
	}
	if err := os.WriteFile(filepath.Join(*directory, "SHA256SUMS"), manifest, 0644); err != nil {
		fail(fmt.Errorf("write SHA256SUMS: %w", err))
	}
	if err := os.WriteFile(filepath.Join(*directory, "SHA256SUMS.sig"), signature, 0644); err != nil {
		fail(fmt.Errorf("write SHA256SUMS.sig: %w", err))
	}
	fmt.Printf("signed %d release assets in %s\n", strings.Count(string(manifest), "\n"), *directory)
}

func decodePrivateKey(raw string) (ed25519.PrivateKey, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, fmt.Errorf("BOB_GEMINI_FREE_UPDATE_PRIVATE_KEY is not configured")
	}
	decoded, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		decoded, err = hex.DecodeString(raw)
	}
	if err != nil {
		return nil, fmt.Errorf("private key must be base64 or hexadecimal: %w", err)
	}
	if len(decoded) == ed25519.SeedSize {
		return ed25519.NewKeyFromSeed(decoded), nil
	}
	if len(decoded) != ed25519.PrivateKeySize {
		return nil, fmt.Errorf("invalid private key length: got %d bytes", len(decoded))
	}
	return ed25519.PrivateKey(decoded), nil
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
