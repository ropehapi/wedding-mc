package service

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// StorageService handles file storage operations.
type StorageService interface {
	Upload(ctx context.Context, key string, r io.Reader, contentType string) (publicURL string, err error)
	Delete(ctx context.Context, key string) error
}

// LocalStorage stores files on the local filesystem.
type LocalStorage struct {
	basePath string
	baseURL  string
}

// NewLocalStorage creates a LocalStorage that saves files under basePath and
// returns public URLs with the given baseURL prefix
// (e.g. "http://localhost:8080/uploads").
func NewLocalStorage(basePath, baseURL string) *LocalStorage {
	return &LocalStorage{basePath: basePath, baseURL: baseURL}
}

// Upload saves the content from r to basePath/key and returns the public URL.
func (s *LocalStorage) Upload(_ context.Context, key string, r io.Reader, _ string) (string, error) {
	dest := filepath.Join(s.basePath, filepath.FromSlash(key))
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return "", fmt.Errorf("create dirs: %w", err)
	}
	f, err := os.Create(dest)
	if err != nil {
		return "", fmt.Errorf("create file: %w", err)
	}
	defer f.Close()
	if _, err := io.Copy(f, r); err != nil {
		return "", fmt.Errorf("write file: %w", err)
	}
	return s.baseURL + "/" + key, nil
}

// Delete removes the file at basePath/key.
// It is not an error if the file does not exist.
func (s *LocalStorage) Delete(_ context.Context, key string) error {
	dest := filepath.Join(s.basePath, filepath.FromSlash(key))
	if err := os.Remove(dest); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("delete file: %w", err)
	}
	return nil
}
