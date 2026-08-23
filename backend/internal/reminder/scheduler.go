package reminder

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"gopuppy/internal/clock"
)

type Scanner interface {
	ScanDue(ctx context.Context, day time.Time) error
	Recover(ctx context.Context) error
}

type Scheduler struct {
	Log      *slog.Logger
	ScanHour int
	Tick     time.Duration
	Scanner  Scanner
	mu       sync.Mutex
	lastDay  string
}

func (s *Scheduler) Start(ctx context.Context) {
	if s.Tick <= 0 {
		s.Tick = 30 * time.Second
	}
	if s.ScanHour < 0 || s.ScanHour > 23 {
		s.ScanHour = 8
	}
	go func() {
		if err := s.Scanner.Recover(ctx); err != nil {
			s.Log.Error("reminder recover", "err", err)
		}
		t := time.NewTicker(s.Tick)
		defer t.Stop()
		s.maybeScan(ctx)
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				s.maybeScan(ctx)
			}
		}
	}()
}

func (s *Scheduler) maybeScan(ctx context.Context) {
	now := clock.Now()
	if clock.Hour(now) != s.ScanHour {
		return
	}
	day := clock.FormatDate(now)
	s.mu.Lock()
	if s.lastDay == day {
		s.mu.Unlock()
		return
	}
	s.lastDay = day
	s.mu.Unlock()
	start := clock.Now()
	if err := s.Scanner.ScanDue(ctx, clock.Today()); err != nil {
		s.Log.Error("reminder scan", "err", err, "day", day)
		s.mu.Lock()
		s.lastDay = ""
		s.mu.Unlock()
		return
	}
	elapsed := clock.Now().Sub(start)
	s.Log.Info("reminder scan done", "day", day, "elapsed_ms", elapsed.Milliseconds())
}

func (s *Scheduler) ForceScan(ctx context.Context) error {
	s.mu.Lock()
	s.lastDay = ""
	s.mu.Unlock()
	return s.Scanner.ScanDue(ctx, clock.Today())
}
