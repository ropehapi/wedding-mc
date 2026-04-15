package service

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLocalStorage_Upload_SavesFileAndReturnsURL(t *testing.T) {
	dir := t.TempDir()
	s := NewLocalStorage(dir, "http://localhost:8080/uploads")
	content := []byte("fake image bytes")

	url, err := s.Upload(context.Background(), "weddings/abc/photo.jpg", bytes.NewReader(content), "image/jpeg")
	if err != nil {
		t.Fatalf("Upload: %v", err)
	}

	if url != "http://localhost:8080/uploads/weddings/abc/photo.jpg" {
		t.Errorf("URL: got %q", url)
	}

	dest := filepath.Join(dir, "weddings", "abc", "photo.jpg")
	got, err := os.ReadFile(dest)
	if err != nil {
		t.Fatalf("file not found after upload: %v", err)
	}
	if string(got) != string(content) {
		t.Errorf("content: got %q, want %q", got, content)
	}
}

func TestLocalStorage_Upload_CreatesIntermediateDirs(t *testing.T) {
	dir := t.TempDir()
	s := NewLocalStorage(dir, "http://localhost")

	_, err := s.Upload(context.Background(), "a/b/c/d/file.png", bytes.NewReader([]byte("x")), "image/png")
	if err != nil {
		t.Fatalf("Upload: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "a", "b", "c", "d", "file.png")); err != nil {
		t.Errorf("file not found after upload: %v", err)
	}
}

func TestLocalStorage_Delete_RemovesFile(t *testing.T) {
	dir := t.TempDir()
	s := NewLocalStorage(dir, "http://localhost")

	_, _ = s.Upload(context.Background(), "to-delete.jpg", bytes.NewReader([]byte("x")), "image/jpeg")

	if err := s.Delete(context.Background(), "to-delete.jpg"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "to-delete.jpg")); !os.IsNotExist(err) {
		t.Error("file should have been deleted")
	}
}

func TestLocalStorage_Delete_NonExistentFile_NoError(t *testing.T) {
	dir := t.TempDir()
	s := NewLocalStorage(dir, "http://localhost")

	if err := s.Delete(context.Background(), "nonexistent.jpg"); err != nil {
		t.Errorf("Delete nonexistent should not error: %v", err)
	}
}

func TestLocalStorage_Upload_URLContainsKey(t *testing.T) {
	dir := t.TempDir()
	s := NewLocalStorage(dir, "http://cdn.example.com")
	key := "weddings/xyz-123/uuid-abc.webp"

	url, err := s.Upload(context.Background(), key, bytes.NewReader([]byte("x")), "image/webp")
	if err != nil {
		t.Fatalf("Upload: %v", err)
	}
	if !strings.HasSuffix(url, key) {
		t.Errorf("URL %q should end with key %q", url, key)
	}
}
