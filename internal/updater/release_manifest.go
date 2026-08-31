package updater

import (
	"bufio"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type manifestEntry struct {
	name   string
	digest string
}

const (
	maxReleaseManifestBytes  = 1 << 20
	maxReleaseSignatureBytes = 4096
)

// CreateSignedManifest creates the exact SHA256SUMS and detached signature
// format consumed by the updater. Existing manifest files are excluded so the
// command can be rerun safely in a release-assets directory.
func CreateSignedManifest(directory string, privateKey ed25519.PrivateKey) ([]byte, []byte, error) {
	if len(privateKey) != ed25519.PrivateKeySize {
		return nil, nil, fmt.Errorf("invalid Ed25519 private key length: got %d bytes, want %d", len(privateKey), ed25519.PrivateKeySize)
	}

	entries, err := os.ReadDir(directory)
	if err != nil {
		return nil, nil, err
	}
	manifestEntries := make([]manifestEntry, 0, len(entries))
	for _, entry := range entries {
		if entry.Name() == "SHA256SUMS" || entry.Name() == "SHA256SUMS.sig" {
			continue
		}
		if entry.IsDir() {
			return nil, nil, fmt.Errorf("release asset %s is not a regular file", entry.Name())
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return nil, nil, fmt.Errorf("release asset %s is a symlink; refusing to sign an indirect file", entry.Name())
		}
		info, err := entry.Info()
		if err != nil {
			return nil, nil, fmt.Errorf("inspect release asset %s: %w", entry.Name(), err)
		}
		if !info.Mode().IsRegular() {
			return nil, nil, fmt.Errorf("release asset %s is not a regular file", entry.Name())
		}
		path := filepath.Join(directory, entry.Name())
		file, err := os.Open(path)
		if err != nil {
			return nil, nil, fmt.Errorf("open release asset %s: %w", entry.Name(), err)
		}
		hasher := sha256.New()
		_, copyErr := io.Copy(hasher, file)
		closeErr := file.Close()
		if copyErr != nil {
			return nil, nil, fmt.Errorf("hash release asset %s: %w", entry.Name(), copyErr)
		}
		if closeErr != nil {
			return nil, nil, fmt.Errorf("close release asset %s: %w", entry.Name(), closeErr)
		}
		manifestEntries = append(manifestEntries, manifestEntry{
			name:   entry.Name(),
			digest: hex.EncodeToString(hasher.Sum(nil)),
		})
	}
	if len(manifestEntries) == 0 {
		return nil, nil, fmt.Errorf("release directory contains no signable assets")
	}

	sort.Slice(manifestEntries, func(i, j int) bool {
		return manifestEntries[i].name < manifestEntries[j].name
	})
	manifest := make([]byte, 0, len(manifestEntries)*80)
	for _, entry := range manifestEntries {
		manifest = append(manifest, []byte(entry.digest+"  "+entry.name+"\n")...)
	}
	signature := ed25519.Sign(privateKey, manifest)
	encodedSignature := []byte(base64.StdEncoding.EncodeToString(signature) + "\n")
	return manifest, encodedSignature, nil
}

// VerifySignedReleaseDirectory verifies the detached signature and then
// reconciles every regular file in directory with exactly one manifest entry.
// The two manifest control files are excluded from the asset set. This is an
// operator-facing post-signing/public-download gate; it does not contact a
// remote release server.
func VerifySignedReleaseDirectory(directory string, publicKey ed25519.PublicKey) error {
	if len(publicKey) != ed25519.PublicKeySize {
		return fmt.Errorf("invalid Ed25519 public key length: got %d bytes, want %d", len(publicKey), ed25519.PublicKeySize)
	}
	entries, err := os.ReadDir(directory)
	if err != nil {
		return fmt.Errorf("read release directory: %w", err)
	}
	manifest, err := readReleaseControlFile(filepath.Join(directory, "SHA256SUMS"), maxReleaseManifestBytes)
	if err != nil {
		return fmt.Errorf("read SHA256SUMS: %w", err)
	}
	signature, err := readReleaseControlFile(filepath.Join(directory, "SHA256SUMS.sig"), maxReleaseSignatureBytes)
	if err != nil {
		return fmt.Errorf("read SHA256SUMS.sig: %w", err)
	}
	decodedSignature, err := base64.StdEncoding.DecodeString(strings.TrimSpace(string(signature)))
	if err != nil || len(decodedSignature) != ed25519.SignatureSize || !ed25519.Verify(publicKey, manifest, decodedSignature) {
		return fmt.Errorf("SHA256SUMS detached signature is invalid")
	}

	want := make(map[string]string, len(entries))
	scanner := bufio.NewScanner(strings.NewReader(string(manifest)))
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) != 2 || strings.HasPrefix(fields[0], "#") {
			return fmt.Errorf("invalid SHA256SUMS entry: %q", scanner.Text())
		}
		name := strings.TrimPrefix(fields[1], "*")
		if name == "" || filepath.Base(name) != name || name == "SHA256SUMS" || name == "SHA256SUMS.sig" {
			return fmt.Errorf("invalid release asset name in SHA256SUMS: %q", name)
		}
		digest := strings.ToLower(fields[0])
		if len(digest) != sha256.Size*2 {
			return fmt.Errorf("invalid SHA-256 digest for %s", name)
		}
		if _, err := hex.DecodeString(digest); err != nil {
			return fmt.Errorf("invalid SHA-256 digest for %s: %w", name, err)
		}
		if previous, exists := want[name]; exists {
			if previous != digest {
				return fmt.Errorf("conflicting SHA-256 entries for %s", name)
			}
			return fmt.Errorf("duplicate SHA-256 entry for %s", name)
		}
		want[name] = digest
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("scan SHA256SUMS: %w", err)
	}
	if len(want) == 0 {
		return fmt.Errorf("SHA256SUMS contains no release assets")
	}

	seen := make(map[string]struct{}, len(entries))
	for _, entry := range entries {
		name := entry.Name()
		if name == "SHA256SUMS" || name == "SHA256SUMS.sig" {
			continue
		}
		if entry.IsDir() || entry.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("release asset %s is not a regular file", name)
		}
		info, err := entry.Info()
		if err != nil {
			return fmt.Errorf("inspect release asset %s: %w", name, err)
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("release asset %s is not a regular file", name)
		}
		digest, exists := want[name]
		if !exists {
			return fmt.Errorf("release asset %s is missing from SHA256SUMS", name)
		}
		actual, err := hashReleaseAsset(filepath.Join(directory, name))
		if err != nil {
			return err
		}
		if actual != digest {
			return fmt.Errorf("SHA-256 mismatch for release asset %s", name)
		}
		seen[name] = struct{}{}
	}
	if len(seen) != len(want) {
		for name := range want {
			if _, exists := seen[name]; !exists {
				return fmt.Errorf("SHA256SUMS contains an asset that is missing from the directory: %s", name)
			}
		}
	}
	return nil
}

func hashReleaseAsset(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("open release asset %s: %w", filepath.Base(path), err)
	}
	defer file.Close()
	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", fmt.Errorf("hash release asset %s: %w", filepath.Base(path), err)
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func readReleaseControlFile(path string, maxBytes int64) ([]byte, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
		return nil, fmt.Errorf("release control file is not a regular file: %s", filepath.Base(path))
	}
	if info.Size() > maxBytes {
		return nil, fmt.Errorf("release control file exceeds %d bytes: %s", maxBytes, filepath.Base(path))
	}
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
		return nil, fmt.Errorf("release control file exceeds %d bytes: %s", maxBytes, filepath.Base(path))
	}
	return data, nil
}
