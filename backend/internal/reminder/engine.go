package reminder

import (
	"errors"
	"net"
	"strings"
	"time"

	"gopuppy/internal/clock"
	"gopuppy/internal/domain"
)

var (
	ErrAuthFailure      = errors.New("auth failure")
	ErrValidationRemote = errors.New("remote validation")
)

// NextDue returns the next civil due date in Beijing after lastDone + cycleDays.
func NextDue(lastDone time.Time, cycleDays int) time.Time {
	if cycleDays <= 0 {
		cycleDays = 1
	}
	base := clock.DateOf(lastDone)
	return base.AddDate(0, 0, cycleDays)
}

func Recalc(rule *domain.ReminderRule, lastDone time.Time) {
	rule.LastDoneAt = lastDone
	rule.NextDueAt = NextDue(lastDone, rule.CycleDays)
}

func IsDueOn(rule domain.ReminderRule, day time.Time) bool {
	if !rule.Enabled {
		return false
	}
	return clock.DateOf(rule.NextDueAt).Equal(clock.DateOf(day))
}

func IsAdvanceOn(rule domain.ReminderRule, day time.Time) bool {
	if !rule.Enabled || rule.AdvanceDays <= 0 {
		return false
	}
	advance := clock.DateOf(rule.NextDueAt).AddDate(0, 0, -rule.AdvanceDays)
	return advance.Equal(clock.DateOf(day))
}

// ClassifyDeliveryError maps a delivery error to transient vs permanent.
// Auth (401/403) and validation (422) are never retried.
func ClassifyDeliveryError(err error) (transient bool, status domain.NotifyStatus) {
	if err == nil {
		return false, domain.NotifySent
	}
	msg := strings.ToLower(err.Error())
	if errors.Is(err, ErrAuthFailure) || strings.Contains(msg, "401") || strings.Contains(msg, "403") {
		return false, domain.NotifyPermanentFailure
	}
	if errors.Is(err, ErrValidationRemote) || strings.Contains(msg, "422") {
		return false, domain.NotifyPermanentFailure
	}
	var ne net.Error
	if errors.As(err, &ne) && ne.Timeout() {
		return true, domain.NotifyFailed
	}
	if strings.Contains(msg, "timeout") || strings.Contains(msg, "429") ||
		strings.Contains(msg, "500") || strings.Contains(msg, "502") ||
		strings.Contains(msg, "503") || strings.Contains(msg, "504") {
		return true, domain.NotifyFailed
	}
	if errors.Is(err, domain.ErrTransient) {
		return true, domain.NotifyFailed
	}
	if errors.Is(err, domain.ErrPermanent) {
		return false, domain.NotifyPermanentFailure
	}
	return true, domain.NotifyFailed
}

func Backoff(attempt int) time.Duration {
	switch attempt {
	case 1:
		return 2 * time.Second
	case 2:
		return 8 * time.Second
	default:
		return 32 * time.Second
	}
}
