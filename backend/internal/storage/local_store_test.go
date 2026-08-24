package storage_test

import (
	"bytes"
	"context"
	"io"
	"testing"

	"gopuppy/internal/config"
	"gopuppy/internal/domain"
	"gopuppy/internal/storage"
)

func TestNewLocalStoreUsesExistingRoot(t *testing.T) {
	root := t.TempDir()
	store, err := storage.New(&config.Config{
		StorageDriver:    "local",
		StorageLocalRoot: root,
		JWTSecret:        "test-secret",
	})
	if err != nil {
		t.Fatalf("create local store for existing root: %v", err)
	}
	if got := store.Driver(); got != domain.DriverLocal {
		t.Fatalf("storage driver = %q, want %q", got, domain.DriverLocal)
	}

	want := []byte("existing local storage remains writable")
	if err := store.Put(context.Background(), "pets/test/photo.bin", bytes.NewReader(want), int64(len(want)), "application/octet-stream"); err != nil {
		t.Fatalf("put into existing local root: %v", err)
	}
	rc, _, err := store.Get(context.Background(), "pets/test/photo.bin")
	if err != nil {
		t.Fatalf("get from existing local root: %v", err)
	}
	defer rc.Close()
	got, err := io.ReadAll(rc)
	if err != nil {
		t.Fatalf("read stored object: %v", err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("stored object = %q, want %q", got, want)
	}
}
