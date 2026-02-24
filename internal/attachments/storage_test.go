package attachments

import (
	"context"
	"io"
	"os"
	"strings"
	"testing"
)

func TestLocalStorage_SaveAndRead(t *testing.T) {
	dir := t.TempDir()
	storage, err := NewLocalStorage(dir)
	if err != nil {
		t.Fatalf("NewLocalStorage: %v", err)
	}

	content := "hello world test file"
	reader := strings.NewReader(content)

	path, checksum, size, err := storage.Save(context.Background(), "ab-test.txt", reader)
	if err != nil {
		t.Fatalf("Save: %v", err)
	}

	if size != int64(len(content)) {
		t.Errorf("expected size %d, got %d", len(content), size)
	}
	if !strings.HasPrefix(checksum, "sha256:") {
		t.Errorf("expected sha256 checksum, got %q", checksum)
	}
	if path == "" {
		t.Error("expected non-empty path")
	}

	// Read back
	rc, err := storage.Read(context.Background(), path)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	defer rc.Close()

	data, err := io.ReadAll(rc)
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if string(data) != content {
		t.Errorf("expected %q, got %q", content, string(data))
	}
}

func TestLocalStorage_Delete(t *testing.T) {
	dir := t.TempDir()
	storage, err := NewLocalStorage(dir)
	if err != nil {
		t.Fatalf("NewLocalStorage: %v", err)
	}

	content := "temp file"
	reader := strings.NewReader(content)

	path, _, _, err := storage.Save(context.Background(), "de-test.txt", reader)
	if err != nil {
		t.Fatalf("Save: %v", err)
	}

	// File should exist
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatal("file should exist after save")
	}

	// Delete
	if err := storage.Delete(context.Background(), path); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	// File should not exist
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Error("file should not exist after delete")
	}
}

func TestLocalStorage_ReadNonexistent(t *testing.T) {
	dir := t.TempDir()
	storage, err := NewLocalStorage(dir)
	if err != nil {
		t.Fatalf("NewLocalStorage: %v", err)
	}

	_, err = storage.Read(context.Background(), "/nonexistent/path/file.txt")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}
