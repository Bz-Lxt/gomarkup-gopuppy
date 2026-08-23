package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gopuppy/internal/clock"
	"gopuppy/internal/domain"
	"gopuppy/internal/repo"
	"gopuppy/internal/ws"
)

type Checkin struct {
	Repo   *repo.Checkins
	Family *Family
	Hub    *ws.Hub
}

type CheckinInput struct {
	Type   domain.CheckinType `json:"type"`
	Slot   domain.Slot        `json:"slot"`
	Done   bool               `json:"done"`
	Date   string             `json:"date"`
}

func (s *Checkin) Toggle(ctx context.Context, userID, petID uuid.UUID, in CheckinInput) ([]domain.DailyCheckin, error) {
	p, _, err := s.Family.MustWritePet(ctx, userID, petID)
	if err != nil {
		return nil, err
	}
	if !in.Type.Valid() || !in.Slot.Valid() {
		return nil, fmt.Errorf("%w: type/slot", domain.ErrValidation)
	}
	day := clock.Today()
	if in.Date != "" {
		d, err := clock.ParseDate(in.Date)
		if err != nil {
			return nil, fmt.Errorf("%w: date", domain.ErrValidation)
		}
		day = d
	}
	now := clock.Now()
	if in.Done {
		c := &domain.DailyCheckin{
			ID: uuid.New(), PetID: petID, CheckinDate: day, Type: in.Type, Slot: in.Slot, DoneBy: userID, DoneAt: now,
		}
		if err := s.Repo.Upsert(ctx, c); err != nil {
			return nil, err
		}
	} else {
		if err := s.Repo.Revoke(ctx, petID, day, in.Type, in.Slot, now); err != nil {
			return nil, err
		}
	}
	list, err := s.Repo.ListDay(ctx, petID, day)
	if err != nil {
		return nil, err
	}
	s.Hub.Broadcast(p.FamilyID, domain.WSMessage{
		Type: "checkin.updated", PetID: petID, Payload: map[string]any{"date": clock.FormatDate(day), "items": list},
	})
	return list, nil
}

func (s *Checkin) Today(ctx context.Context, userID, petID uuid.UUID) ([]domain.DailyCheckin, error) {
	if _, _, err := s.Family.MustReadPet(ctx, userID, petID); err != nil {
		return nil, err
	}
	return s.Repo.ListDay(ctx, petID, clock.Today())
}
