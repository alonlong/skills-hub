package local

import (
	"bytes"
	"path/filepath"
	"testing"
)

func TestLocalStorageSaveAndRead(t *testing.T) {
	storage := NewService(t.TempDir())
	key := filepath.Join("team-ai", "email-helper", "1.0.0.zip")
	data := []byte("archive-bytes")

	path, err := storage.Save(t.Context(), key, data)
	if err != nil {
		t.Fatalf("save archive: %v", err)
	}
	if path == "" {
		t.Fatal("expected non-empty storage path")
	}

	readBack, err := storage.Read(t.Context(), key)
	if err != nil {
		t.Fatalf("read archive: %v", err)
	}
	if !bytes.Equal(readBack, data) {
		t.Fatal("stored bytes do not match")
	}
}
