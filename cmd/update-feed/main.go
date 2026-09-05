package main

import (
	"bytes"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/div197/bob-gemini-free/internal/updater"
)

const maxFeedInputBytes = 4 << 20

func main() {
	feedPath := flag.String("file", "updates/desktop-feed.json", "signed update-feed JSON file")
	signaturePath := flag.String("signature", "", "detached signature output (default: <file>.sig)")
	publicKey := flag.String("public-key", "", "optional base64 or hexadecimal public key to match against the signing key")
	privateKeyStdin := flag.Bool("private-key-stdin", false, "read the private key from stdin instead of BOB_GEMINI_FREE_UPDATE_PRIVATE_KEY")
	flag.Parse()

	if strings.TrimSpace(*signaturePath) == "" {
		*signaturePath = *feedPath + ".sig"
	}
	feed, err := readBoundedFile(*feedPath, maxFeedInputBytes)
	if err != nil {
		fail(fmt.Errorf("read update feed: %w", err))
	}
	privateKeyInput := os.Getenv("BOB_GEMINI_FREE_UPDATE_PRIVATE_KEY")
	if *privateKeyStdin {
		privateKeyInput, err = readPrivateKeyStdin()
		if err != nil {
			fail(err)
		}
	}
	privateKey, err := decodePrivateKey(privateKeyInput)
	if err != nil {
		fail(err)
	}
	if err := verifyPublicKeyMatch(privateKey, *publicKey); err != nil {
		fail(err)
	}
	signature, err := updater.SignUpdateFeed(feed, privateKey)
	if err != nil {
		fail(err)
	}
	if err := writeAtomic(*signaturePath, signature); err != nil {
		fail(fmt.Errorf("write update feed signature: %w", err))
	}
	fmt.Printf("signed update feed %s -> %s\n", *feedPath, *signaturePath)
}

func readBoundedFile(path string, maxBytes int64) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, maxBytes+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > maxBytes {
		return nil, fmt.Errorf("file exceeds %d bytes", maxBytes)
	}
	return data, nil
}

func readPrivateKeyStdin() (string, error) {
	const maxPrivateKeyInputBytes = 4096
	data, err := io.ReadAll(io.LimitReader(os.Stdin, maxPrivateKeyInputBytes+1))
	if err != nil {
		return "", fmt.Errorf("read private key from stdin: %w", err)
	}
	if len(data) > maxPrivateKeyInputBytes {
		return "", fmt.Errorf("private key stdin input exceeds %d bytes", maxPrivateKeyInputBytes)
	}
	return string(data), nil
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
	decoded, err := decodeKeyBytes(raw)
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
	decoded, err := decodeKeyBytes(raw)
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

func decodeKeyBytes(raw string) ([]byte, error) {
	if len(raw)%2 == 0 {
		if decoded, err := hex.DecodeString(raw); err == nil {
			return decoded, nil
		}
	}
	decoded, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		return nil, err
	}
	return decoded, nil
}

func writeAtomic(path string, data []byte) error {
	directory := filepath.Dir(path)
	temporary, err := os.CreateTemp(directory, ".desktop-feed-signature-*")
	if err != nil {
		return err
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	if err := temporary.Chmod(0644); err != nil {
		_ = temporary.Close()
		return err
	}
	if _, err := temporary.Write(data); err != nil {
		_ = temporary.Close()
		return err
	}
	if err := temporary.Sync(); err != nil {
		_ = temporary.Close()
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	return os.Rename(temporaryPath, path)
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
