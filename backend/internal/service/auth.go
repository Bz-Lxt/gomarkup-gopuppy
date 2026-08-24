package service

import (
	"context"
	"fmt"
	"net/mail"
	"strings"
	"unicode"

	"github.com/google/uuid"
	"gopuppy/internal/auth"
	"gopuppy/internal/clock"
	"gopuppy/internal/domain"
	"gopuppy/internal/repo"
)

type Auth struct {
	Users  *repo.Users
	Issuer *auth.Issuer
}

type RegisterInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Nickname string `json:"nickname"`
}

func (s *Auth) Register(ctx context.Context, in RegisterInput) (*domain.User, *auth.Tokens, error) {
	email := strings.ToLower(strings.TrimSpace(in.Email))
	if _, err := mail.ParseAddress(email); err != nil {
		return nil, nil, fmt.Errorf("%w: invalid email", domain.ErrValidation)
	}
	if err := validatePassword(in.Password); err != nil {
		return nil, nil, err
	}
	nick := strings.TrimSpace(in.Nickname)
	if nick == "" || len([]rune(nick)) > 24 {
		return nil, nil, fmt.Errorf("%w: nickname required (1-24)", domain.ErrValidation)
	}
	hash, err := auth.HashPassword(in.Password)
	if err != nil {
		return nil, nil, fmt.Errorf("%w: %v", domain.ErrValidation, err)
	}
	now := clock.Now()
	u := &domain.User{ID: uuid.New(), Email: email, PasswordHash: hash, Nickname: nick, CreatedAt: now, UpdatedAt: now}
	if err := s.Users.Create(ctx, u); err != nil {
		return nil, nil, err
	}
	tok, err := s.Issuer.Issue(u.ID, now)
	return u, tok, err
}

func (s *Auth) Login(ctx context.Context, email, password string) (*domain.User, *auth.Tokens, error) {
	u, err := s.Users.ByEmail(ctx, strings.ToLower(strings.TrimSpace(email)))
	if err != nil {
		return nil, nil, domain.ErrInvalidCredential
	}
	if err := auth.ComparePassword(u.PasswordHash, password); err != nil {
		return nil, nil, domain.ErrInvalidCredential
	}
	tok, err := s.Issuer.Issue(u.ID, clock.Now())
	return u, tok, err
}

func (s *Auth) Refresh(ctx context.Context, refresh string) (*auth.Tokens, error) {
	c, err := s.Issuer.Parse(refresh)
	if err != nil || c.Kind != "refresh" {
		return nil, domain.ErrUnauthorized
	}
	if _, err := s.Users.ByID(ctx, c.UserID); err != nil {
		return nil, domain.ErrUnauthorized
	}
	return s.Issuer.Issue(c.UserID, clock.Now())
}

func (s *Auth) Me(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	return s.Users.ByID(ctx, id)
}

func validatePassword(p string) error {
	if len(p) < 8 || len(p) > 72 {
		return fmt.Errorf("%w: password 8-72 chars", domain.ErrValidation)
	}
	var letter, digit bool
	for _, r := range p {
		if unicode.IsLetter(r) {
			letter = true
		}
		if unicode.IsDigit(r) {
			digit = true
		}
	}
	if !letter || !digit {
		return fmt.Errorf("%w: password needs letters and digits", domain.ErrValidation)
	}
	return nil
}
