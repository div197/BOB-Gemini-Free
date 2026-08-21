package main

import (
	"bytes"
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
	publicKey := flag.String("public-key", "", "optional base64 or hexadecimal public key to match against the signing key")
	flag.Parse()

	privateKey, err := decodePrivateKey(os.Getenv("BOB_GEMINI_FREE_UPDATE_PRIVATE_KEY"))
	if err != nil {
		fail(err)
	}
	if err := verifyPublicKeyMatch(privateKey, *publicKey); err != nil {
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

func verifyPublicKeyMatch(privateKey ed25519.PrivateKey, encodedPublicKey string) error {
	if strings.TrimSpace(encodedPublicKey) == "" {
		return nil
	}
	publicKey, err := decodePublicKey(encodedPublicKey)
	if err != nil {
		return err
	}
	derived, ok := privateKey.Public().(ed25519.PublicKey)
	if !ok || !bytes.Equal(derived, publicKey) {
		return fmt.Errorf("configured public key does not match the release signing private key")
	}
	return nil
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

func decodePrivateKey(raw string) (ed25519.PrivateKey, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, fmt.Errorf("BOB_GEMINI_FREE_UPDATE_PRIVATE_KEY is not configured")
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
