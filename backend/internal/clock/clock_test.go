package clock

import (
	"testing"
	"time"
)

func TestTodayIsBeijingMidnight(t *testing.T) {
	d := Today()
	if d.Hour() != 0 || d.Minute() != 0 || d.Location() != Beijing {
		t.Fatalf("today %+v", d)
	}
}

func TestParseAndFormatRoundTrip(t *testing.T) {
	d, err := ParseDate("2026-08-23")
	if err != nil {
		t.Fatal(err)
	}
	if FormatDate(d) != "2026-08-23" {
		t.Fatal(FormatDate(d))
	}
	ts := time.Date(2026, 8, 23, 16, 7, 9, 0, Beijing)
	if FormatDateTime(ts) != "2026-08-23 16:07:09" {
		t.Fatal(FormatDateTime(ts))
	}
}

func TestHourUsesBeijingNotUTC(t *testing.T) {
	// 2026-08-23 00:30 CST == 2026-08-22 16:30 UTC
	ts := time.Date(2026, 8, 22, 16, 30, 0, 0, time.UTC)
	if Hour(ts) != 0 {
		t.Fatalf("hour %d, UTC date would be 22nd 16h", Hour(ts))
	}
	if FormatDate(ts) != "2026-08-23" {
		t.Fatalf("civil date drifted to %s", FormatDate(ts))
	}
}
