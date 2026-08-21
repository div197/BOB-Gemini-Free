package updater

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
)

func signedManifestFixture(t *testing.T, assetName string, payload []byte) ([]byte, []byte, ed25519.PublicKey) {
	t.Helper()
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}
	digest := sha256.Sum256(payload)
	manifest := []byte(hexDigest(digest[:]) + "  " + assetName + "\n")
	signature := ed25519.Sign(privateKey, manifest)
	encodedSignature := []byte(base64.StdEncoding.EncodeToString(signature))
	return manifest, encodedSignature, publicKey
}

func hexDigest(bytes []byte) string {
	const hexChars = "0123456789abcdef"
	result := make([]byte, len(bytes)*2)
	for i, b := range bytes {
		result[i*2] = hexChars[b>>4]
		result[i*2+1] = hexChars[b&0x0f]
	}
	return string(result)
}

func TestVerifyReleaseManifestSignature(t *testing.T) {
	manifest, signature, publicKey := signedManifestFixture(t, "bob-gemini-free-darwin-arm64", []byte("fixture-binary"))
	if err := verifyReleaseManifest(publicKey, manifest, signature); err != nil {
		t.Fatalf("verifyReleaseManifest: %v", err)
	}

	tampered := append([]byte(nil), manifest...)
	tampered[0] = '0'
	if err := verifyReleaseManifest(publicKey, tampered, signature); err == nil {
		t.Fatal("tampered manifest was accepted")
	}
	if err := verifyReleaseManifest(publicKey, manifest, []byte("not-a-signature")); err == nil {
		t.Fatal("invalid signature encoding was accepted")
	}
}

func TestDownloadVerifiedArtifactUsesMockedHTTP(t *testing.T) {
	const assetName = "bob-gemini-free-darwin-arm64"
	payload := []byte("fixture-binary-payload")
	manifest, signature, publicKey := signedManifestFixture(t, assetName, payload)
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(payload)
	}))
	defer server.Close()

	destination := filepath.Join(t.TempDir(), "candidate")
	digest, err := downloadVerifiedArtifact(server.Client(), server.URL, destination, assetName, manifest, signature, publicKey)
	if err != nil {
		t.Fatalf("downloadVerifiedArtifact: %v", err)
	}
	if requests.Load() != 1 {
		t.Fatalf("mock download count = %d, want 1", requests.Load())
	}
	got, err := os.ReadFile(destination)
	if err != nil {
		t.Fatalf("read candidate: %v", err)
	}
	if string(got) != string(payload) {
		t.Fatalf("candidate = %q, want %q", got, payload)
	}
	if !strings.EqualFold(digest, hexDigest(sha256Bytes(payload))) {
		t.Fatalf("returned digest = %q", digest)
	}
}

func sha256Bytes(payload []byte) []byte {
	digest := sha256.Sum256(payload)
	return digest[:]
}

func TestDownloadVerifiedArtifactRejectsManifestMismatch(t *testing.T) {
	const assetName = "bob-gemini-free-darwin-arm64"
	manifest, signature, publicKey := signedManifestFixture(t, assetName, []byte("different-payload"))
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("actual-payload"))
	}))
	defer server.Close()

	destination := filepath.Join(t.TempDir(), "candidate")
	if _, err := downloadVerifiedArtifact(server.Client(), server.URL, destination, assetName, manifest, signature, publicKey); err == nil {
		t.Fatal("download with manifest mismatch was accepted")
	}
}
