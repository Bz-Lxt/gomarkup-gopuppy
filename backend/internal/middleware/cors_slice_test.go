package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"gopuppy/internal/middleware"
)

func TestCORSWhitelistRemainsStableAcrossAllowedOrigins(t *testing.T) {
	const (
		dashboardOrigin = "https://dashboard.gopuppy.example"
		mobileOrigin    = "https://m.gopuppy.example"
	)
	h := middleware.CORS([]string{dashboardOrigin, mobileOrigin})(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	assertAllowed := func(origin string) {
		t.Helper()
		req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
		req.Header.Set("Origin", origin)
		w := httptest.NewRecorder()

		h.ServeHTTP(w, req)

		if got := w.Header().Get("Access-Control-Allow-Origin"); got != origin {
			t.Fatalf("request from %q received Access-Control-Allow-Origin %q", origin, got)
		}
	}

	assertAllowed(mobileOrigin)
	assertAllowed(dashboardOrigin)
}
