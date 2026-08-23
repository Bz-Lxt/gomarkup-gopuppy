package domain

import (
	"testing"
	"time"

	"gopuppy/internal/clock"
)

func TestCalcAgeExactDays(t *testing.T) {
	b := time.Date(2023, 3, 15, 8, 0, 0, 0, clock.Beijing)
	now := time.Date(2026, 8, 23, 16, 0, 0, 0, clock.Beijing)
	age := CalcAge(b, now)
	if age.Years != 3 || age.Months != 5 || age.Days != 8 {
		t.Fatalf("got %+v want 3y5m8d", age)
	}
	wantTotal := int(time.Date(2026, 8, 23, 0, 0, 0, 0, clock.Beijing).Sub(time.Date(2023, 3, 15, 0, 0, 0, 0, clock.Beijing)).Hours() / 24)
	if age.TotalDays != wantTotal {
		t.Fatalf("total %d want %d", age.TotalDays, wantTotal)
	}
}

func TestCalcAgeLeapDay(t *testing.T) {
	b := time.Date(2020, 2, 29, 0, 0, 0, 0, clock.Beijing)
	now := time.Date(2021, 3, 1, 0, 0, 0, 0, clock.Beijing)
	age := CalcAge(b, now)
	if age.Years != 1 || age.Days != 1 {
		t.Fatalf("leap clamp got %+v", age)
	}
	if age.TotalDays != 366 {
		t.Fatalf("2020-02-29 to 2021-03-01 should be 366 days, got %d", age.TotalDays)
	}
}

func TestCalcAgeSameDay(t *testing.T) {
	b := time.Date(2026, 8, 23, 0, 0, 0, 0, clock.Beijing)
	age := CalcAge(b, b.Add(3*time.Hour))
	if age.TotalDays != 0 || age.Years != 0 {
		t.Fatalf("same day %+v", age)
	}
}
