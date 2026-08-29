package main

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/hex"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/div197/bob-gemini-free/internal/updater"
)

func main() {
	directory := flag.String("dir", "release-assets", "directory containing release assets")
	publicKey := flag.String("public-key", "", "base64 or hexadecimal Ed25519 public key")
	flag.Parse()

	key, err := decodePublicKey(*publicKey)
	if err != nil {
		fail(err)
	}
	if err := updater.VerifySignedReleaseDirectory(*directory, key); err != nil {
		fail(err)
	}
	fmt.Printf("verified signed release directory: %s\n", *directory)
}

func decodePublicKey(raw string) (ed25519.PublicKey, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, fmt.Errorf("public key is empty")
	}
	var decoded []byte
	var err error
	if len(raw)%2 == 0 {
		decoded, err = hex.DecodeString(raw)
	}
	if err != nil || len(decoded) == 0 {
		decoded, err = base64.StdEncoding.DecodeString(raw)
	}
	if err != nil {
		return nil, fmt.Errorf("public key must be base64 or hexadecimal: %w", err)
	}
	if len(decoded) != ed25519.PublicKeySize {
		return nil, fmt.Errorf("invalid public key length: got %d bytes", len(decoded))
	}
	return ed25519.PublicKey(decoded), nil
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
