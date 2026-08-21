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
	"testing"
)

func TestCreateSignedManifestIsDeterministicAndVerifiable(t *testing.T) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	directory := t.TempDir()
	if err := os.WriteFile(filepath.Join(directory, "bob-linux-amd64"), []byte("linux"), 0644); err != nil {
		t.Fatalf("write linux asset: %v", err)
	}
	if err := os.WriteFile(filepath.Join(directory, "bob-darwin-arm64"), []byte("darwin"), 0644); err != nil {
		t.Fatalf("write darwin asset: %v", err)
	}
	if err := os.WriteFile(filepath.Join(directory, "SHA256SUMS"), []byte("stale"), 0644); err != nil {
		t.Fatalf("write stale manifest: %v", err)
	}

	manifest, signature, err := CreateSignedManifest(directory, privateKey)
	if err != nil {
		t.Fatalf("CreateSignedManifest: %v", err)
	}
	if !strings.HasSuffix(string(manifest), "bob-linux-amd64\n") {
		t.Fatalf("manifest is not sorted: %q", manifest)
	}
	if strings.Contains(string(manifest), "SHA256SUMS") {
		t.Fatalf("manifest included its own control file: %q", manifest)
	}
	decodedSignature, err := base64.StdEncoding.DecodeString(strings.TrimSpace(string(signature)))
	if err != nil {
		t.Fatalf("decode signature: %v", err)
	}
	if !ed25519.Verify(publicKey, manifest, decodedSignature) {
		t.Fatal("generated signature did not verify")
	}

	linuxDigest := sha256.Sum256([]byte("linux"))
	if !strings.Contains(string(manifest), hexDigest(linuxDigest[:])+"  bob-linux-amd64\n") {
		t.Fatalf("manifest missing linux digest: %q", manifest)
	}
}

func TestDownloadVerifiedArtifactLimitedRejectsOversizedResponse(t *testing.T) {
	const assetName = "fixture-binary"
	payload := []byte("payload larger than the test limit")
	manifest, signature, publicKey := signedManifestFixture(t, assetName, payload)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(payload)
	}))
	defer server.Close()
	destination := filepath.Join(t.TempDir(), "candidate")
	if _, err := downloadVerifiedArtifactLimited(server.Client(), server.URL, destination, assetName, manifest, signature, publicKey, 4); err == nil {
		t.Fatal("oversized artifact was accepted")
	}
}
