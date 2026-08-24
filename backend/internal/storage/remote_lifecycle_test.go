package storage_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"gopuppy/internal/domain"
	"gopuppy/internal/storage"
)

func TestRemoteGetBodyRemainsReadable(t *testing.T) {
	const payload = "complete object contents"
	body := &observableBody{Reader: strings.NewReader(payload)}
	originalClient := http.DefaultClient
	http.DefaultClient = &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/octet-stream"}},
			Body:       body,
			Request:    req,
		}, nil
	})}
	t.Cleanup(func() { http.DefaultClient = originalClient })

	store := &storage.Remote{
		Kind:      domain.DriverOSS,
		Endpoint:  "https://objects.example.test",
		AccessKey: "access-key",
		Secret:    "secret",
	}
	rc, contentType, err := store.Get(context.Background(), "pets/pet-1/photo.jpg")
	if err != nil {
		t.Fatalf("Get returned an error: %v", err)
	}
	if contentType != "application/octet-stream" {
		t.Fatalf("content type = %q", contentType)
	}

	got, err := io.ReadAll(rc)
	if err != nil {
		t.Fatalf("body became unreadable after Get returned: %v", err)
	}
	if string(got) != payload {
		t.Fatalf("body = %q, want %q", got, payload)
	}
	if body.closed {
		t.Fatal("response body was closed before the caller released it")
	}
	if err := rc.Close(); err != nil {
		t.Fatalf("close returned an error: %v", err)
	}
	if !body.closed {
		t.Fatal("caller close did not release the response body")
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

type observableBody struct {
	*strings.Reader
	closed bool
}

func (b *observableBody) Read(p []byte) (int, error) {
	if b.closed {
		return 0, errors.New("read after close")
	}
	return b.Reader.Read(p)
}

func (b *observableBody) Close() error {
	b.closed = true
	return nil
}
