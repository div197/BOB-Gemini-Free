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

func TestConfiguredUpdatePublicKeyPrefersEmbeddedKey(t *testing.T) {
	embeddedPublic, _, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("generate embedded key: %v", err)
	}
	environmentPublic, _, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("generate environment key: %v", err)
	}
	previous := BuildUpdatePublicKey
	BuildUpdatePublicKey = base64.StdEncoding.EncodeToString(embeddedPublic)
	t.Cleanup(func() { BuildUpdatePublicKey = previous })
	t.Setenv("BOB_GEMINI_FREE_UPDATE_PUBLIC_KEY", base64.StdEncoding.EncodeToString(environmentPublic))

	got, err := configuredUpdatePublicKey()
	if err != nil {
		t.Fatalf("configuredUpdatePublicKey: %v", err)
	}
	if string(got) != string(embeddedPublic) {
		t.Fatal("runtime environment unexpectedly overrode embedded update trust anchor")
	}
}

func TestConfiguredUpdatePublicKeyUsesEnvironmentFallback(t *testing.T) {
	previous := BuildUpdatePublicKey
	BuildUpdatePublicKey = ""
	t.Cleanup(func() { BuildUpdatePublicKey = previous })
	publicKey, _, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	t.Setenv("BOB_GEMINI_FREE_UPDATE_PUBLIC_KEY", base64.StdEncoding.EncodeToString(publicKey))

	got, err := configuredUpdatePublicKey()
	if err != nil {
		t.Fatalf("configuredUpdatePublicKey: %v", err)
	}
	if string(got) != string(publicKey) {
		t.Fatal("environment update trust anchor was not loaded")
	}
}

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

func TestCreateSignedManifestRejectsSymlinkAsset(t *testing.T) {
	directory := t.TempDir()
	target := filepath.Join(t.TempDir(), "outside-asset")
	if err := os.WriteFile(target, []byte("outside"), 0600); err != nil {
		t.Fatalf("write outside asset: %v", err)
	}
	link := filepath.Join(directory, "release.bin")
	if err := os.Symlink(target, link); err != nil {
		t.Skipf("symlinks unavailable on this host: %v", err)
	}
	_, _, err := CreateSignedManifest(directory, make(ed25519.PrivateKey, ed25519.PrivateKeySize))
	if err == nil || !strings.Contains(err.Error(), "symlink") {
		t.Fatalf("symlink asset was accepted: %v", err)
	}
}
