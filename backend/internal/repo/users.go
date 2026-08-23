package repo

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"gopuppy/internal/domain"
)

type Users struct{ Pool *pgxpool.Pool }

func (r *Users) Create(ctx context.Context, u *domain.User) error {
	_, err := r.Pool.Exec(ctx, `INSERT INTO users(id,email,password_hash,nickname,avatar_url,created_at,updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		u.ID, strings.ToLower(u.Email), u.PasswordHash, u.Nickname, u.AvatarURL, u.CreatedAt, u.UpdatedAt)
	if err != nil && strings.Contains(err.Error(), "users_email_key") {
		return domain.ErrConflict
	}
	return err
}

func (r *Users) ByEmail(ctx context.Context, email string) (*domain.User, error) {
	return r.scan(ctx, `SELECT id,email,password_hash,nickname,avatar_url,created_at,updated_at FROM users WHERE email=$1`, strings.ToLower(email))
}

func (r *Users) ByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	return r.scan(ctx, `SELECT id,email,password_hash,nickname,avatar_url,created_at,updated_at FROM users WHERE id=$1`, id)
}

func (r *Users) scan(ctx context.Context, q string, arg any) (*domain.User, error) {
	var u domain.User
	err := r.Pool.QueryRow(ctx, q, arg).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Nickname, &u.AvatarURL, &u.CreatedAt, &u.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return &u, err
}
