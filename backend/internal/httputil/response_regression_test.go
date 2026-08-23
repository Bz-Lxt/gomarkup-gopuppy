package httputil

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDecodeErrorMapsToValidationResponse(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/api/v1/families", strings.NewReader(`{"name":`))
	var input struct {
		Name string `json:"name"`
	}
	err := DecodeJSON(r, &input)
	if err == nil {
		t.Fatal("expected malformed JSON to be rejected")
	}

	w := httptest.NewRecorder()
	Error(w, "req-malformed-json", err)

	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want %d; body = %s", w.Code, http.StatusUnprocessableEntity, w.Body.String())
	}
	var got Envelope
	if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if got.Code != "VALIDATION" {
		t.Fatalf("code = %q, want VALIDATION", got.Code)
	}
}
