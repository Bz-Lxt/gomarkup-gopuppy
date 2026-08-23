package notifier_test

import (
	"context"
	"testing"

	"gopuppy/internal/domain"
	"gopuppy/internal/notifier"
	"gopuppy/internal/reminder"
)

func TestHTTPSenderCancellationRemainsRetryable(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	sender := &notifier.HTTPSender{
		Name: "WEBHOOK",
		URL:  "http://127.0.0.1:1/deliver",
		Mode: "live",
	}
	err := sender.Send(ctx, notifier.Message{
		Channel: domain.ChannelHook,
		Title:   "vaccine reminder",
		Body:    "scheduled delivery",
	})
	if err == nil {
		t.Fatal("expected canceled delivery to fail")
	}

	transient, status := reminder.ClassifyDeliveryError(err)
	if !transient || status != domain.NotifyFailed {
		t.Fatalf("canceled delivery should remain retryable: transient=%v status=%s err=%v", transient, status, err)
	}
}
