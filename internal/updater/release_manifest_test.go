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

func TestVerifySignedReleaseDirectoryReconcilesEveryAsset(t *testing.T) {
	directory := t.TempDir()
	if err := os.WriteFile(filepath.Join(directory, "bob.zip"), []byte("package"), 0600); err != nil {
		t.Fatalf("write asset: %v", err)
	}
	if err := os.WriteFile(filepath.Join(directory, "RELEASE-NOTICE.txt"), []byte("notice"), 0600); err != nil {
		t.Fatalf("write notice: %v", err)
	}
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}
	manifest, signature, err := CreateSignedManifest(directory, privateKey)
	if err != nil {
		t.Fatalf("CreateSignedManifest: %v", err)
	}
	if err := os.WriteFile(filepath.Join(directory, "SHA256SUMS"), manifest, 0600); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
	if err := os.WriteFile(filepath.Join(directory, "SHA256SUMS.sig"), signature, 0600); err != nil {
		t.Fatalf("write signature: %v", err)
	}
	if err := VerifySignedReleaseDirectory(directory, publicKey); err != nil {
		t.Fatalf("VerifySignedReleaseDirectory: %v", err)
	}

	if err := os.WriteFile(filepath.Join(directory, "unlisted.txt"), []byte("extra"), 0600); err != nil {
		t.Fatalf("write extra asset: %v", err)
	}
	if err := VerifySignedReleaseDirectory(directory, publicKey); err == nil || !strings.Contains(err.Error(), "missing from SHA256SUMS") {
		t.Fatalf("unlisted asset result = %v, want manifest mismatch", err)
	}
}

func TestVerifySignedReleaseDirectoryRejectsTamperingAndDuplicateEntries(t *testing.T) {
	directory := t.TempDir()
	assetPath := filepath.Join(directory, "bob.zip")
	if err := os.WriteFile(assetPath, []byte("package"), 0600); err != nil {
		t.Fatalf("write asset: %v", err)
	}
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}
	manifest, signature, err := CreateSignedManifest(directory, privateKey)
	if err != nil {
		t.Fatalf("CreateSignedManifest: %v", err)
	}
	if err := os.WriteFile(filepath.Join(directory, "SHA256SUMS"), manifest, 0600); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
	if err := os.WriteFile(filepath.Join(directory, "SHA256SUMS.sig"), signature, 0600); err != nil {
		t.Fatalf("write signature: %v", err)
	}
	if err := os.WriteFile(assetPath, []byte("tampered"), 0600); err != nil {
		t.Fatalf("tamper asset: %v", err)
	}
	if err := VerifySignedReleaseDirectory(directory, publicKey); err == nil || !strings.Contains(err.Error(), "SHA-256 mismatch") {
		t.Fatalf("tampered asset result = %v, want checksum mismatch", err)
	}

	if err := os.WriteFile(assetPath, []byte("package"), 0600); err != nil {
		t.Fatalf("restore asset: %v", err)
	}
	digest := sha256.Sum256([]byte("package"))
	duplicateManifest := []byte(hexDigest(digest[:]) + "  bob.zip\n" + hexDigest(digest[:]) + "  bob.zip\n")
	duplicateSignature := ed25519.Sign(privateKey, duplicateManifest)
	if err := os.WriteFile(filepath.Join(directory, "SHA256SUMS"), duplicateManifest, 0600); err != nil {
		t.Fatalf("write duplicate manifest: %v", err)
	}
	if err := os.WriteFile(filepath.Join(directory, "SHA256SUMS.sig"), []byte(base64.StdEncoding.EncodeToString(duplicateSignature)), 0600); err != nil {
		t.Fatalf("write duplicate signature: %v", err)
	}
	if err := VerifySignedReleaseDirectory(directory, publicKey); err == nil || !strings.Contains(err.Error(), "duplicate") {
		t.Fatalf("duplicate manifest result = %v, want duplicate rejection", err)
	}
}

func TestVerifySignedReleaseDirectoryBoundsControlFiles(t *testing.T) {
	directory := t.TempDir()
	if err := os.WriteFile(filepath.Join(directory, "SHA256SUMS"), []byte(strings.Repeat("x", maxReleaseManifestBytes+1)), 0600); err != nil {
		t.Fatalf("write oversized manifest: %v", err)
	}
	if err := os.WriteFile(filepath.Join(directory, "SHA256SUMS.sig"), []byte("signature"), 0600); err != nil {
		t.Fatalf("write signature: %v", err)
	}
	if err := VerifySignedReleaseDirectory(directory, make(ed25519.PublicKey, ed25519.PublicKeySize)); err == nil || !strings.Contains(err.Error(), "control file exceeds") {
		t.Fatalf("oversized manifest result = %v, want bounded read failure", err)
	}

	if err := os.WriteFile(filepath.Join(directory, "SHA256SUMS"), []byte("manifest"), 0600); err != nil {
		t.Fatalf("write small manifest: %v", err)
	}
	if err := os.WriteFile(filepath.Join(directory, "SHA256SUMS.sig"), []byte(strings.Repeat("x", maxReleaseSignatureBytes+1)), 0600); err != nil {
		t.Fatalf("write oversized signature: %v", err)
	}
	if err := VerifySignedReleaseDirectory(directory, make(ed25519.PublicKey, ed25519.PublicKeySize)); err == nil || !strings.Contains(err.Error(), "control file exceeds") {
		t.Fatalf("oversized signature result = %v, want bounded read failure", err)
	}
}

func TestVerifySignedReleaseDirectoryRejectsControlSymlinks(t *testing.T) {
	directory := t.TempDir()
	outside := filepath.Join(t.TempDir(), "outside")
	if err := os.WriteFile(outside, []byte("control"), 0600); err != nil {
		t.Fatalf("write outside control: %v", err)
	}
	if err := os.Symlink(outside, filepath.Join(directory, "SHA256SUMS")); err != nil {
		t.Skipf("symlinks unavailable on this host: %v", err)
	}
	if err := os.WriteFile(filepath.Join(directory, "SHA256SUMS.sig"), []byte("signature"), 0600); err != nil {
		t.Fatalf("write signature: %v", err)
	}
	if err := VerifySignedReleaseDirectory(directory, make(ed25519.PublicKey, ed25519.PublicKeySize)); err == nil || !strings.Contains(err.Error(), "regular file") {
		t.Fatalf("control symlink result = %v, want regular-file rejection", err)
	}
}
