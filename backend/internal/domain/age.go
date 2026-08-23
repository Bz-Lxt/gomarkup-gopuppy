package domain

import (
	"gopuppy/internal/clock"
	"time"
)

type Age struct {
	Years     int `json:"years"`
	Months    int `json:"months"`
	Days      int `json:"days"`
	TotalDays int `json:"total_days"`
}

// CalcAge computes civil age in Asia/Shanghai. Leap-day birthdays clamp to Feb 28
// in non-leap years so the day count stays exact.
func CalcAge(birthday, now time.Time) Age {
	b := birthday.In(clock.Beijing)
	n := now.In(clock.Beijing)
	bDate := time.Date(b.Year(), b.Month(), b.Day(), 0, 0, 0, 0, clock.Beijing)
	nDate := time.Date(n.Year(), n.Month(), n.Day(), 0, 0, 0, 0, clock.Beijing)
	if nDate.Before(bDate) {
		return Age{}
	}
	years, months, days := 0, 0, 0
	y, m, d := bDate.Year(), bDate.Month(), bDate.Day()
	for {
		next := addCivilMonth(y, m, d, 1)
		if next.After(nDate) {
			break
		}
		y, m = next.Year(), next.Month()
		months++
	}
	years = months / 12
	months = months % 12
	anchor := civilDate(y, m, d)
	days = int(nDate.Sub(anchor).Hours() / 24)
	total := int(nDate.Sub(bDate).Hours() / 24)
	return Age{Years: years, Months: months, Days: days, TotalDays: total}
}

func addCivilMonth(year int, month time.Month, day, delta int) time.Time {
	m := int(month) + delta
	for m > 12 {
		year++
		m -= 12
	}
	for m < 1 {
		year--
		m += 12
	}
	return civilDate(year, time.Month(m), day)
}

func civilDate(year int, month time.Month, day int) time.Time {
	last := time.Date(year, month+1, 0, 0, 0, 0, 0, clock.Beijing).Day()
	if day > last {
		day = last
	}
	return time.Date(year, month, day, 0, 0, 0, 0, clock.Beijing)
}
