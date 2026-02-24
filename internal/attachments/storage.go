package attachments

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// StorageConfig configures the file storage backend.
type StorageConfig struct {
	Backend   string `koanf:"backend"`    // "local" or "s3" (s3 not yet implemented)
	LocalPath string `koanf:"local_path"` // directory for local storage
	MaxSizeMB int    `koanf:"max_size_mb"`
}

// Storage defines the interface for file storage backends.
type Storage interface {
	Save(ctx context.Context, filename string, reader io.Reader) (storagePath string, checksum string, sizeBytes int64, err error)
	Read(ctx context.Context, storagePath string) (io.ReadCloser, error)
	Delete(ctx context.Context, storagePath string) error
}

// LocalStorage stores files on the local filesystem.
type LocalStorage struct {
	basePath string
}

// NewLocalStorage creates a local file storage backend.
func NewLocalStorage(basePath string) (*LocalStorage, error) {
	if err := os.MkdirAll(basePath, 0750); err != nil {
		return nil, fmt.Errorf("creating storage directory: %w", err)
	}
	return &LocalStorage{basePath: basePath}, nil
}

// Save writes a file to the local filesystem.
func (s *LocalStorage) Save(ctx context.Context, filename string, reader io.Reader) (string, string, int64, error) {
	// Create a unique subdirectory to avoid collisions
	dir := filepath.Join(s.basePath, filename[:2])
	if err := os.MkdirAll(dir, 0750); err != nil {
		return "", "", 0, fmt.Errorf("creating subdirectory: %w", err)
	}

	path := filepath.Join(dir, filename)
	f, err := os.Create(path)
	if err != nil {
		return "", "", 0, fmt.Errorf("creating file: %w", err)
	}
	defer f.Close()

	hasher := sha256.New()
	tee := io.TeeReader(reader, hasher)

	n, err := io.Copy(f, tee)
	if err != nil {
		return "", "", 0, fmt.Errorf("writing file: %w", err)
	}

	checksum := fmt.Sprintf("sha256:%x", hasher.Sum(nil))
	return path, checksum, n, nil
}

// Read opens a file from the local filesystem.
func (s *LocalStorage) Read(ctx context.Context, storagePath string) (io.ReadCloser, error) {
	f, err := os.Open(storagePath)
	if err != nil {
		return nil, fmt.Errorf("opening file: %w", err)
	}
	return f, nil
}

// Delete removes a file from the local filesystem.
func (s *LocalStorage) Delete(ctx context.Context, storagePath string) error {
	return os.Remove(storagePath)
}
