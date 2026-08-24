package storage_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"testing"
	"time"

	"gopuppy/internal/domain"
	"gopuppy/internal/storage"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func TestRemotePutStopsWhenCallerIsCanceled(t *testing.T) {
	started := make(chan struct{})
	release := make(chan struct{})
	originalClient := http.DefaultClient
	http.DefaultClient = &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		close(started)
		select {
		case <-r.Context().Done():
			return nil, r.Context().Err()
		case <-release:
			return &http.Response{
				StatusCode: http.StatusNoContent,
				Header:     make(http.Header),
				Body:       io.NopCloser(bytes.NewReader(nil)),
			}, nil
		}
	})}
	t.Cleanup(func() {
		http.DefaultClient = originalClient
		close(release)
	})

	store := &storage.Remote{
		Kind:      domain.DriverOSS,
		Endpoint:  "https://objects.example.test",
		AccessKey: "access-key",
		Secret:    "secret",
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		done <- store.Put(ctx, "pets/pet-1/PHOTO/photo.jpg", bytes.NewReader([]byte("photo")), 5, "image/jpeg")
	}()

	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("remote upload did not start")
	}
	cancel()

	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("Put returned %v after caller cancellation; want context canceled", err)
		}
	case <-time.After(300 * time.Millisecond):
		t.Fatal("Put continued waiting on object storage after caller cancellation")
	}
}
