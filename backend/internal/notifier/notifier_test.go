package notifier

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"gopuppy/internal/domain"
	"gopuppy/internal/reminder"
)

func TestHTTPSenderMockWithoutURL(t *testing.T) {
	s := &HTTPSender{Name: "WECOM_BOT", URL: "", Mode: "mock"}
	if err := s.Send(context.Background(), Message{Title: "t", Body: "b"}); err != nil {
		t.Fatal(err)
	}
}

func TestHTTPSenderClassifiesAuth(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(401)
	}))
	defer srv.Close()
	s := &HTTPSender{Name: "WEBHOOK", URL: srv.URL, Mode: "real"}
	err := s.Send(context.Background(), Message{Title: "t", Body: "b"})
	if err == nil {
		t.Fatal("expected error")
	}
	tr, st := reminder.ClassifyDeliveryError(err)
	if tr || st != domain.NotifyPermanentFailure {
		t.Fatalf("401 must be permanent: %v %s %v", tr, st, err)
	}
}

func TestHTTPSenderTransient5xx(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(503)
	}))
	defer srv.Close()
	s := &HTTPSender{Name: "WEBHOOK", URL: srv.URL, Mode: "real"}
	err := s.Send(context.Background(), Message{Title: "t", Body: "b"})
	if err == nil || !strings.Contains(err.Error(), "503") {
		t.Fatalf("%v", err)
	}
	tr, _ := reminder.ClassifyDeliveryError(err)
	if !tr {
		t.Fatal("503 should retry")
	}
}

func TestFanoutUnknownChannelPermanent(t *testing.T) {
	f := &Fanout{}
	err := f.Send(context.Background(), Message{Channel: "SMS"})
	if err == nil {
		t.Fatal("expected error")
	}
}
