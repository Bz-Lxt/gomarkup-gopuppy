package storage

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"gopuppy/internal/clock"
	"gopuppy/internal/domain"
)

func TestSniffWhitelist(t *testing.T) {
	if mime, err := Sniff([]byte{0xFF, 0xD8, 0xFF, 0xE0}); err != nil || mime != "image/jpeg" {
		t.Fatalf("jpeg %s %v", mime, err)
	}
	png := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	if mime, err := Sniff(png); err != nil || mime != "image/png" {
		t.Fatalf("png %s %v", mime, err)
	}
	webp := append([]byte("RIFF"), make([]byte, 4)...)
	webp = append(webp, []byte("WEBP")...)
	if mime, err := Sniff(webp); err != nil || mime != "image/webp" {
		t.Fatalf("webp %s %v", mime, err)
	}
	if mime, err := Sniff([]byte("%PDF-1.7")); err != nil || mime != "application/pdf" {
		t.Fatalf("pdf %s %v", mime, err)
	}
	if _, err := Sniff([]byte("MZ")); err != domain.ErrUnsupportedMedia {
		t.Fatalf("exe should reject, %v", err)
	}
}

func TestSanitizeFilename(t *testing.T) {
	if err := SanitizeFilename("../etc/passwd"); err != domain.ErrPathTraversal {
		t.Fatal("expected traversal")
	}
	if err := SanitizeFilename("/abs.pdf"); err != domain.ErrPathTraversal {
		t.Fatal("abs")
	}
	if err := SanitizeFilename("report.pdf"); err != nil {
		t.Fatal(err)
	}
}

func TestHashReaderTooLarge(t *testing.T) {
	big := bytes.Repeat([]byte("a"), MaxFileBytes+2)
	if _, _, err := HashReader(bytes.NewReader(big)); err != domain.ErrTooLarge {
		t.Fatalf("want too large, got %v", err)
	}
}

func TestBuildKeyIsolatesPet(t *testing.T) {
	at := time.Date(2026, 8, 1, 12, 0, 0, 0, clock.Beijing)
	key := BuildKey("pet-a", "PHOTO", "abc", ".jpg", at)
	if !strings.HasPrefix(key, "pets/pet-a/PHOTO/2026-08/abc.jpg") {
		t.Fatalf("key %s", key)
	}
}
