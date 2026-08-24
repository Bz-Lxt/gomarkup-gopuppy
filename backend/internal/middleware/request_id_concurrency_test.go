package middleware_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"gopuppy/internal/httputil"
	"gopuppy/internal/middleware"
)

func TestRequestIDKeepsConcurrentResponsesIsolated(t *testing.T) {
	slowEntered := make(chan struct{})
	releaseSlow := make(chan struct{})
	fastEntered := make(chan struct{})
	releaseFast := make(chan struct{})
	handler := middleware.RequestID()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := middleware.RequestIDFrom(r.Context())
		status := http.StatusCreated
		switch requestID {
		case "slow-request":
			close(slowEntered)
			<-releaseSlow
			status = http.StatusAccepted
		case "fast-request":
			close(fastEntered)
			<-releaseFast
		}
		httputil.JSON(w, status, requestID, "OK", "ok", map[string]string{"request": requestID})
	}))

	slowRequest := httptest.NewRequest(http.MethodGet, "/slow", nil)
	slowRequest.Header.Set("X-Request-ID", "slow-request")
	slowResponse := httptest.NewRecorder()
	slowDone := make(chan struct{})
	go func() {
		defer close(slowDone)
		handler.ServeHTTP(slowResponse, slowRequest)
	}()

	<-slowEntered
	fastRequest := httptest.NewRequest(http.MethodGet, "/fast", nil)
	fastRequest.Header.Set("X-Request-ID", "fast-request")
	fastResponse := httptest.NewRecorder()
	fastDone := make(chan struct{})
	go func() {
		defer close(fastDone)
		handler.ServeHTTP(fastResponse, fastRequest)
	}()

	<-fastEntered
	close(releaseSlow)
	<-slowDone
	close(releaseFast)
	<-fastDone

	assertResponseOwner(t, slowResponse, http.StatusAccepted, "slow-request")
	assertResponseOwner(t, fastResponse, http.StatusCreated, "fast-request")
}

func assertResponseOwner(t *testing.T, response *httptest.ResponseRecorder, status int, requestID string) {
	t.Helper()
	if response.Code != status {
		t.Fatalf("response %q status = %d, want %d", requestID, response.Code, status)
	}
	var body struct {
		RequestID string            `json:"request_id"`
		Data      map[string]string `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("response %q has invalid JSON %q: %v", requestID, response.Body.String(), err)
	}
	if body.RequestID != requestID || body.Data["request"] != requestID {
		t.Fatalf("response %q contains request_id=%q data.request=%q", requestID, body.RequestID, body.Data["request"])
	}
}
