package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Claims struct {
	UserID uuid.UUID `json:"uid"`
	Kind   string    `json:"knd"`
	jwt.RegisteredClaims
}

type Tokens struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

type Issuer struct {
	Secret     []byte
	AccessTTL  time.Duration
	RefreshTTL time.Duration
}

func (i *Issuer) Issue(userID uuid.UUID, now time.Time) (*Tokens, error) {
	access, err := i.sign(userID, "access", now, i.AccessTTL)
	if err != nil {
		return nil, err
	}
	refresh, err := i.sign(userID, "refresh", now, i.RefreshTTL)
	if err != nil {
		return nil, err
	}
	return &Tokens{
		AccessToken:  access,
		RefreshToken: refresh,
		ExpiresAt:    now.Add(i.AccessTTL),
	}, nil
}

func (i *Issuer) sign(userID uuid.UUID, kind string, now time.Time, ttl time.Duration) (string, error) {
	c := Claims{
		UserID: userID,
		Kind:   kind,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			ID:        uuid.NewString(),
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	return t.SignedString(i.Secret)
}

func (i *Issuer) Parse(token string) (*Claims, error) {
	parsed, err := jwt.ParseWithClaims(token, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return i.Secret, nil
	})
	if err != nil {
		return nil, err
	}
	c, ok := parsed.Claims.(*Claims)
	if !ok || !parsed.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	return c, nil
}
