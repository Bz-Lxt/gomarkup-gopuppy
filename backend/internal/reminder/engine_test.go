package reminder

import (
	"errors"
	"net"
	"testing"
	"time"

	"gopuppy/internal/clock"
	"gopuppy/internal/domain"
)

func TestNextDue(t *testing.T) {
	last := time.Date(2025, 6, 18, 15, 30, 0, 0, clock.Beijing)
	got := NextDue(last, 365)
	want := time.Date(2026, 6, 18, 0, 0, 0, 0, clock.Beijing)
	if !got.Equal(want) {
		t.Fatalf("got %s want %s", got, want)
	}
}

func TestRecalcDynamic(t *testing.T) {
	rule := domain.ReminderRule{CycleDays: 90, AdvanceDays: 3, Enabled: true}
	done := time.Date(2026, 8, 23, 10, 0, 0, 0, clock.Beijing)
	Recalc(&rule, done)
	if !clock.DateOf(rule.NextDueAt).Equal(time.Date(2026, 11, 21, 0, 0, 0, 0, clock.Beijing)) {
		t.Fatalf("next %s", rule.NextDueAt)
	}
}

func TestIsDueAndAdvance(t *testing.T) {
	rule := domain.ReminderRule{
		Enabled: true, AdvanceDays: 3,
		NextDueAt: time.Date(2026, 8, 26, 0, 0, 0, 0, clock.Beijing),
	}
	if !IsAdvanceOn(rule, time.Date(2026, 8, 23, 9, 0, 0, 0, clock.Beijing)) {
		t.Fatal("expected advance")
	}
	if IsDueOn(rule, time.Date(2026, 8, 23, 9, 0, 0, 0, clock.Beijing)) {
		t.Fatal("not due yet")
	}
	if !IsDueOn(rule, time.Date(2026, 8, 26, 8, 0, 0, 0, clock.Beijing)) {
		t.Fatal("expected due")
	}
}

func TestClassifyNeverRetryAuth(t *testing.T) {
	tr, st := ClassifyDeliveryError(ErrAuthFailure)
	if tr || st != domain.NotifyPermanentFailure {
		t.Fatalf("auth should be permanent, got %v %s", tr, st)
	}
	tr, st = ClassifyDeliveryError(ErrValidationRemote)
	if tr || st != domain.NotifyPermanentFailure {
		t.Fatalf("422 should be permanent")
	}
}

func TestClassifyTransientTimeout(t *testing.T) {
	tr, st := ClassifyDeliveryError(&timeoutErr{})
	if !tr || st != domain.NotifyFailed {
		t.Fatalf("timeout should retry, got %v %s", tr, st)
	}
	tr, st = ClassifyDeliveryError(errors.New("503 bad gateway"))
	if !tr {
		t.Fatal("5xx should retry")
	}
}

type timeoutErr struct{}

func (timeoutErr) Error() string   { return "i/o timeout" }
func (timeoutErr) Timeout() bool   { return true }
func (timeoutErr) Temporary() bool { return true }

var _ net.Error = timeoutErr{}

func TestBackoff(t *testing.T) {
	if Backoff(1) != 2*time.Second || Backoff(2) != 8*time.Second || Backoff(3) != 32*time.Second {
		t.Fatal("backoff schedule")
	}
}
