package updater

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
)

type manifestEntry struct {
	name   string
	digest string
}

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
		if entry.IsDir() || entry.Name() == "SHA256SUMS" || entry.Name() == "SHA256SUMS.sig" {
			continue
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
