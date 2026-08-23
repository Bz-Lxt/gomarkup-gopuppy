package httputil

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"gopuppy/internal/domain"
)

func TestMapErrorHidesCrossFamilyAs404(t *testing.T) {
	st, code, _ := MapError(domain.ErrNotFound)
	if st != http.StatusNotFound || code != "NOT_FOUND" {
		t.Fatalf("%d %s", st, code)
	}
}

func TestDecodeRejectsUnknownFields(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"email":"a@b.c","extra":1}`))
	var dst struct {
		Email string `json:"email"`
	}
	if err := DecodeJSON(r, &dst); err == nil {
		t.Fatal("expected validation")
	}
}

func TestJSONEnvelope(t *testing.T) {
	w := httptest.NewRecorder()
	OK(w, "rid-1", map[string]string{"k": "v"})
	var env Envelope
	if err := json.NewDecoder(bytes.NewReader(w.Body.Bytes())).Decode(&env); err != nil {
		t.Fatal(err)
	}
	if env.Code != "OK" || env.RequestID != "rid-1" {
		t.Fatalf("%+v", env)
	}
}
