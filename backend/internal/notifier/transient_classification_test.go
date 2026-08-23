package notifier_test

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"gopuppy/internal/domain"
	"gopuppy/internal/notifier"
	"gopuppy/internal/reminder"
)

func TestServerErrorBodyDoesNotOverrideTransientClassification(t *testing.T) {
	originalClient := http.DefaultClient
	http.DefaultClient = &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusServiceUnavailable,
			Body:       io.NopCloser(strings.NewReader(`{"message":"dependency unavailable","last_upstream_status":403}`)),
			Header:     make(http.Header),
		}, nil
	})}
	t.Cleanup(func() { http.DefaultClient = originalClient })

	sender := &notifier.HTTPSender{Name: "WEBHOOK", URL: "https://notify.example.test/send", Mode: "real"}
	err := sender.Send(context.Background(), notifier.Message{
		Channel: domain.ChannelHook,
		Title:   "scheduled reminder",
		Body:    "vaccination due",
	})
	if err == nil {
		t.Fatal("expected delivery error")
	}

	transient, status := reminder.ClassifyDeliveryError(err)
	if !transient || status != domain.NotifyFailed {
		t.Fatalf("503 delivery must remain retryable regardless of response diagnostics: transient=%v status=%s err=%v", transient, status, err)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
