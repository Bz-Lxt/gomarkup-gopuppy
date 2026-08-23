package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gopuppy/internal/clock"
	"gopuppy/internal/domain"
	"gopuppy/internal/reminder"
	"gopuppy/internal/repo"
	"gopuppy/internal/ws"
)

type Event struct {
	Events    *repo.Events
	Reminders *repo.Reminders
	Family    *Family
	Hub       *ws.Hub
}

type EventInput struct {
	Category    domain.EventCategory `json:"category"`
	Title       string               `json:"title"`
	Description string               `json:"description"`
	OccurredAt  string               `json:"occurred_at"`
	Clinic      string               `json:"clinic"`
	Severity    domain.Severity      `json:"severity"`
	Treated     bool                 `json:"treated"`
	AmountCents *int64               `json:"amount_cents"`
	MediaIDs    []uuid.UUID          `json:"media_ids"`
}

func (s *Event) Create(ctx context.Context, userID, petID uuid.UUID, in EventInput) (*domain.HealthEvent, error) {
	p, _, err := s.Family.MustWritePet(ctx, userID, petID)
	if err != nil {
		return nil, err
	}
	if !in.Category.Valid() {
		return nil, fmt.Errorf("%w: category", domain.ErrValidation)
	}
	title := strings.TrimSpace(in.Title)
	if title == "" || len([]rune(title)) > 60 {
		return nil, fmt.Errorf("%w: title 1-60", domain.ErrValidation)
	}
	if !in.Severity.Valid() {
		return nil, fmt.Errorf("%w: severity", domain.ErrValidation)
	}
	if in.Category == domain.EventSymptom && in.Severity == "" {
		return nil, fmt.Errorf("%w: symptom requires severity", domain.ErrValidation)
	}
	if in.AmountCents != nil && *in.AmountCents < 0 {
		return nil, fmt.Errorf("%w: amount", domain.ErrValidation)
	}
	occurred, err := parseDateTime(in.OccurredAt)
	if err != nil {
		return nil, fmt.Errorf("%w: occurred_at", domain.ErrValidation)
	}
	e := &domain.HealthEvent{
		ID: uuid.New(), PetID: petID, Category: in.Category, Title: title,
		Description: strings.TrimSpace(in.Description), OccurredAt: occurred, Clinic: strings.TrimSpace(in.Clinic),
		Severity: in.Severity, Treated: in.Treated, AmountCents: in.AmountCents, CreatedBy: userID, CreatedAt: clock.Now(),
		MediaIDs: in.MediaIDs,
	}
	if err := s.Events.Create(ctx, e); err != nil {
		return nil, err
	}
	if kind, ok := in.Category.ReminderKind(); ok {
		next := reminder.NextDue(occurred, 1)
		_ = next
		rules, err := s.Reminders.ListByPet(ctx, petID)
		if err == nil {
			for _, rule := range rules {
				if rule.Kind == kind && rule.Enabled {
					reminder.Recalc(&rule, occurred)
					_ = s.Reminders.Update(ctx, &rule)
				}
			}
		}
	}
	s.Hub.Broadcast(p.FamilyID, domain.WSMessage{Type: "health.event", PetID: petID, Payload: e})
	return e, nil
}

func (s *Event) List(ctx context.Context, userID, petID uuid.UUID, category string, year int) ([]domain.HealthEvent, error) {
	if _, _, err := s.Family.MustReadPet(ctx, userID, petID); err != nil {
		return nil, err
	}
	if category != "" && !domain.EventCategory(category).Valid() {
		return nil, fmt.Errorf("%w: category", domain.ErrValidation)
	}
	return s.Events.List(ctx, petID, category, year)
}

func parseDateTime(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}, fmt.Errorf("empty")
	}
	if t, err := time.ParseInLocation("2006-01-02 15:04:05", s, clock.Beijing); err == nil {
		return t, nil
	}
	if t, err := time.ParseInLocation("2006-01-02", s, clock.Beijing); err == nil {
		return t, nil
	}
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t.In(clock.Beijing), nil
	}
	return time.Time{}, fmt.Errorf("bad datetime")
}
