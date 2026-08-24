package service

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/google/uuid"
	"gopuppy/internal/clock"
	"gopuppy/internal/domain"
	"gopuppy/internal/repo"
)

type Family struct {
	Families *repo.Families
	Pets     *repo.Pets
}

func (s *Family) Create(ctx context.Context, userID uuid.UUID, name string) (*domain.Family, error) {
	name = strings.TrimSpace(name)
	if name == "" || len([]rune(name)) > 40 {
		return nil, fmt.Errorf("%w: family name 1-40", domain.ErrValidation)
	}
	now := clock.Now()
	f := &domain.Family{ID: uuid.New(), Name: name, OwnerID: userID, CreatedAt: now}
	if err := s.Families.Create(ctx, f, domain.RoleOwner, now); err != nil {
		return nil, err
	}
	return f, nil
}

func (s *Family) List(ctx context.Context, userID uuid.UUID) ([]domain.Family, error) {
	return s.Families.ListForUser(ctx, userID)
}

func (s *Family) Members(ctx context.Context, userID, familyID uuid.UUID) ([]domain.FamilyMember, error) {
	if _, err := s.MustMember(ctx, userID, familyID); err != nil {
		return nil, err
	}
	return s.Families.Members(ctx, familyID)
}

func (s *Family) Invite(ctx context.Context, userID, familyID uuid.UUID, role domain.Role) (*domain.FamilyInvite, error) {
	m, err := s.MustMember(ctx, userID, familyID)
	if err != nil {
		return nil, err
	}
	if !m.Role.CanManageMembers() {
		return nil, domain.ErrForbidden
	}
	if !role.Valid() || role == domain.RoleOwner {
		return nil, fmt.Errorf("%w: invite role must be CAREGIVER or VIEWER", domain.ErrValidation)
	}
	code, err := randomCode(6)
	if err != nil {
		return nil, err
	}
	inv := &domain.FamilyInvite{
		ID: uuid.New(), FamilyID: familyID, Code: code, Role: role, ExpiresAt: clock.Now().Add(24 * time.Hour),
	}
	if err := s.Families.CreateInvite(ctx, inv); err != nil {
		return nil, err
	}
	return inv, nil
}

func (s *Family) Join(ctx context.Context, userID uuid.UUID, code string) (*domain.Family, error) {
	code = strings.ToUpper(strings.TrimSpace(code))
	if len(code) != 6 {
		return nil, fmt.Errorf("%w: invite code", domain.ErrValidation)
	}
	inv, err := s.Families.InviteByCode(ctx, code)
	if err != nil {
		return nil, domain.ErrNotFound
	}
	now := clock.Now()
	if inv.UsedAt != nil {
		return nil, domain.ErrInviteUsed
	}
	if now.After(inv.ExpiresAt) {
		return nil, domain.ErrInviteExpired
	}
	if _, err := s.Families.Member(ctx, inv.FamilyID, userID); err == nil {
		return nil, domain.ErrAlreadyMember
	}
	if err := s.Families.ConsumeInvite(ctx, inv.ID, userID, now); err != nil {
		return nil, err
	}
	if err := s.Families.AddMember(ctx, inv.FamilyID, userID, inv.Role, now); err != nil {
		return nil, err
	}
	return s.Families.ByID(ctx, inv.FamilyID)
}

func (s *Family) Remove(ctx context.Context, actor, familyID, target uuid.UUID) error {
	m, err := s.MustMember(ctx, actor, familyID)
	if err != nil {
		return err
	}
	if !m.Role.CanManageMembers() {
		return domain.ErrForbidden
	}
	f, err := s.Families.ByID(ctx, familyID)
	if err != nil {
		return err
	}
	if target == f.OwnerID {
		return fmt.Errorf("%w: cannot remove owner", domain.ErrValidation)
	}
	return s.Families.RemoveMember(ctx, familyID, target)
}

func (s *Family) MustMember(ctx context.Context, userID, familyID uuid.UUID) (*domain.FamilyMember, error) {
	m, err := s.Families.Member(ctx, familyID, userID)
	if err != nil {
		return nil, domain.ErrNotFound
	}
	return m, nil
}

func (s *Family) MustWritePet(ctx context.Context, userID, petID uuid.UUID) (*domain.Pet, *domain.FamilyMember, error) {
	p, err := s.Pets.ByID(ctx, petID)
	if err != nil {
		return nil, nil, domain.ErrNotFound
	}
	m, err := s.MustMember(ctx, userID, p.FamilyID)
	if err != nil {
		return nil, nil, domain.ErrNotFound
	}
	if !m.Role.CanWrite() {
		return nil, nil, domain.ErrForbidden
	}
	return p, m, nil
}

func (s *Family) MustReadPet(ctx context.Context, userID, petID uuid.UUID) (*domain.Pet, *domain.FamilyMember, error) {
	p, err := s.Pets.ByID(ctx, petID)
	if err != nil {
		return nil, nil, domain.ErrNotFound
	}
	m, err := s.MustMember(ctx, userID, p.FamilyID)
	if err != nil {
		return nil, nil, domain.ErrNotFound
	}
	return p, m, nil
}

func randomCode(n int) (string, error) {
	const letters = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	b := make([]byte, n)
	for i := range b {
		v, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			return "", err
		}
		b[i] = letters[v.Int64()]
	}
	return string(b), nil
}
