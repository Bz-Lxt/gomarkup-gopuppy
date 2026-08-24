package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"gopuppy/internal/clock"
	"gopuppy/internal/domain"
	"gopuppy/internal/repo"
)

type Pet struct {
	Pets    *repo.Pets
	Family  *Family
}

type PetInput struct {
	FamilyID  uuid.UUID      `json:"family_id"`
	Name      string         `json:"name"`
	Species   domain.Species `json:"species"`
	Breed     string         `json:"breed"`
	Gender    domain.Gender  `json:"gender"`
	Birthday  string         `json:"birthday"`
	Neutered  bool           `json:"neutered"`
	ChipNo    string         `json:"chip_no"`
	WeightMin *float64       `json:"weight_min"`
	WeightMax *float64       `json:"weight_max"`
	Note      string         `json:"note"`
	AvatarKey string         `json:"avatar_key"`
}

func (s *Pet) Create(ctx context.Context, userID uuid.UUID, in PetInput) (*domain.Pet, error) {
	m, err := s.Family.MustMember(ctx, userID, in.FamilyID)
	if err != nil {
		return nil, err
	}
	if !m.Role.CanWrite() {
		return nil, domain.ErrForbidden
	}
	p, err := buildPet(in)
	if err != nil {
		return nil, err
	}
	p.ID = uuid.New()
	p.FamilyID = in.FamilyID
	p.CreatedAt = clock.Now()
	if err := s.Pets.Create(ctx, p); err != nil {
		return nil, err
	}
	p.Age = domain.CalcAge(p.Birthday, clock.Now())
	return p, nil
}

func (s *Pet) Update(ctx context.Context, userID, petID uuid.UUID, in PetInput) (*domain.Pet, error) {
	cur, _, err := s.Family.MustWritePet(ctx, userID, petID)
	if err != nil {
		return nil, err
	}
	p, err := buildPet(in)
	if err != nil {
		return nil, err
	}
	p.ID = cur.ID
	p.FamilyID = cur.FamilyID
	p.CreatedAt = cur.CreatedAt
	if in.AvatarKey == "" {
		p.AvatarKey = cur.AvatarKey
	}
	if err := s.Pets.Update(ctx, p); err != nil {
		return nil, err
	}
	p.Age = domain.CalcAge(p.Birthday, clock.Now())
	return p, nil
}

func (s *Pet) Archive(ctx context.Context, userID, petID uuid.UUID) error {
	p, m, err := s.Family.MustReadPet(ctx, userID, petID)
	if err != nil {
		return err
	}
	if !m.Role.CanDeletePet() {
		return domain.ErrForbidden
	}
	return s.Pets.Archive(ctx, p.ID, clock.Now())
}

func (s *Pet) Get(ctx context.Context, userID, petID uuid.UUID) (*domain.Pet, error) {
	p, _, err := s.Family.MustReadPet(ctx, userID, petID)
	return p, err
}

func (s *Pet) List(ctx context.Context, userID, familyID uuid.UUID) ([]domain.Pet, error) {
	if _, err := s.Family.MustMember(ctx, userID, familyID); err != nil {
		return nil, err
	}
	return s.Pets.ListByFamily(ctx, familyID, false)
}

func buildPet(in PetInput) (*domain.Pet, error) {
	name := strings.TrimSpace(in.Name)
	if name == "" || len([]rune(name)) > 20 {
		return nil, fmt.Errorf("%w: pet name 1-20", domain.ErrValidation)
	}
	if !in.Species.Valid() || !in.Gender.Valid() {
		return nil, fmt.Errorf("%w: species/gender", domain.ErrValidation)
	}
	bd, err := clock.ParseDate(in.Birthday)
	if err != nil {
		return nil, fmt.Errorf("%w: birthday yyyy-MM-dd", domain.ErrValidation)
	}
	if bd.After(clock.Today()) {
		return nil, fmt.Errorf("%w: birthday in future", domain.ErrValidation)
	}
	if in.WeightMin != nil && in.WeightMax != nil && *in.WeightMin > *in.WeightMax {
		return nil, fmt.Errorf("%w: weight range", domain.ErrValidation)
	}
	return &domain.Pet{
		Name: name, Species: in.Species, Breed: strings.TrimSpace(in.Breed), Gender: in.Gender,
		Birthday: bd, Neutered: in.Neutered, ChipNo: strings.TrimSpace(in.ChipNo),
		WeightMin: in.WeightMin, WeightMax: in.WeightMax, Note: strings.TrimSpace(in.Note),
		AvatarKey: strings.TrimSpace(in.AvatarKey),
	}, nil
}
