package storage_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"gopuppy/internal/storage"
)

func TestHashReaderDataSurvivesContentSniff(t *testing.T) {
	original := make([]byte, 1024)
	copy(original, []byte{0xFF, 0xD8, 0xFF, 0xE0})
	for i := 4; i < len(original); i++ {
		original[i] = byte(i%251 + 1)
	}

	wantData := bytes.Clone(original)
	wantSum := sha256.Sum256(wantData)
	sum, data, err := storage.HashReader(bytes.NewReader(original))
	if err != nil {
		t.Fatalf("read upload: %v", err)
	}
	head := data
	if len(head) > 16 {
		head = head[:16]
	}
	if mime, err := storage.Sniff(head); err != nil || mime != "image/jpeg" {
		t.Fatalf("sniff upload: mime=%q err=%v", mime, err)
	}

	if !bytes.Equal(data, wantData) {
		t.Fatal("content sniff changed bytes returned for storage")
	}
	if sum != hex.EncodeToString(wantSum[:]) {
		t.Fatalf("reported hash %q does not describe the original upload", sum)
	}
	storedSum := sha256.Sum256(data)
	if sum != hex.EncodeToString(storedSum[:]) {
		t.Fatalf("reported hash %q does not describe bytes sent to storage", sum)
	}
}
