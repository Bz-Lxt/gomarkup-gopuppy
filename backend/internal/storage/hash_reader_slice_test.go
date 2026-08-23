package storage_test

import (
	"bytes"
	"testing"

	"gopuppy/internal/storage"
)

func TestHashReaderResultSurvivesLaterReads(t *testing.T) {
	first := bytes.Repeat([]byte("a"), 4096)
	_, got, err := storage.HashReader(bytes.NewReader(first))
	if err != nil {
		t.Fatal(err)
	}

	second := bytes.Repeat([]byte("b"), len(first))
	if _, _, err := storage.HashReader(bytes.NewReader(second)); err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(got, first) {
		t.Fatalf("first read changed after a later read: got prefix %q, want %q", got[:8], first[:8])
	}
}
