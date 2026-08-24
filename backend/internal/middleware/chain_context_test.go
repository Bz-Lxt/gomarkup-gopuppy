package middleware_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"gopuppy/internal/auth"
	"gopuppy/internal/middleware"
)

func TestAuthPreservesRequestCancellation(t *testing.T) {
	issuer := &auth.Issuer{
		Secret:     []byte("middleware-test-secret"),
		AccessTTL:  time.Hour,
		RefreshTTL: 24 * time.Hour,
	}
	tokens, err := issuer.Issue(uuid.New(), time.Now())
	if err != nil {
		t.Fatal(err)
	}

	requestCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil).WithContext(requestCtx)
	req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)

	entered := make(chan struct{})
	released := make(chan struct{})
	observed := make(chan error, 1)
	next := http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		close(entered)
		select {
		case <-r.Context().Done():
			observed <- r.Context().Err()
		case <-released:
			observed <- nil
		}
	})
	h := middleware.RequestID()(middleware.Auth(issuer)(next))
	done := make(chan struct{})
	go func() {
		defer close(done)
		h.ServeHTTP(httptest.NewRecorder(), req)
	}()

	select {
	case <-entered:
	case <-time.After(time.Second):
		t.Fatal("authenticated handler was not reached")
	}
	cancel()

	select {
	case got := <-observed:
		if !errors.Is(got, context.Canceled) {
			t.Fatalf("handler context error = %v, want %v", got, context.Canceled)
		}
	case <-time.After(200 * time.Millisecond):
		close(released)
		<-done
		t.Fatal("authenticated handler did not observe request cancellation")
	}
	<-done
}
