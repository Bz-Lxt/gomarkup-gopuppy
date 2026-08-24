package repo

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"gopuppy/internal/domain"
)

type Families struct{ Pool *pgxpool.Pool }

func (r *Families) Create(ctx context.Context, f *domain.Family, ownerRole domain.Role, joined time.Time) error {
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, `INSERT INTO families(id,name,owner_id,created_at) VALUES ($1,$2,$3,$4)`,
		f.ID, f.Name, f.OwnerID, f.CreatedAt); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `INSERT INTO family_members(family_id,user_id,role,joined_at) VALUES ($1,$2,$3,$4)`,
		f.ID, f.OwnerID, ownerRole, joined); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *Families) ByID(ctx context.Context, id uuid.UUID) (*domain.Family, error) {
	var f domain.Family
	err := r.Pool.QueryRow(ctx, `SELECT id,name,owner_id,created_at FROM families WHERE id=$1`, id).
		Scan(&f.ID, &f.Name, &f.OwnerID, &f.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return &f, err
}

func (r *Families) ListForUser(ctx context.Context, userID uuid.UUID) ([]domain.Family, error) {
	rows, err := r.Pool.Query(ctx, `SELECT f.id,f.name,f.owner_id,f.created_at
		FROM families f JOIN family_members m ON m.family_id=f.id WHERE m.user_id=$1 ORDER BY f.created_at`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Family
	for rows.Next() {
		var f domain.Family
		if err := rows.Scan(&f.ID, &f.Name, &f.OwnerID, &f.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, f)
	}
	return out, rows.Err()
}

func (r *Families) Member(ctx context.Context, familyID, userID uuid.UUID) (*domain.FamilyMember, error) {
	var m domain.FamilyMember
	err := r.Pool.QueryRow(ctx, `SELECT family_id,user_id,role,joined_at FROM family_members WHERE family_id=$1 AND user_id=$2`,
		familyID, userID).Scan(&m.FamilyID, &m.UserID, &m.Role, &m.JoinedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return &m, err
}

func (r *Families) Members(ctx context.Context, familyID uuid.UUID) ([]domain.FamilyMember, error) {
	rows, err := r.Pool.Query(ctx, `SELECT m.family_id,m.user_id,m.role,m.joined_at,u.nickname,u.email
		FROM family_members m JOIN users u ON u.id=m.user_id WHERE m.family_id=$1 ORDER BY m.joined_at`, familyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.FamilyMember
	for rows.Next() {
		var m domain.FamilyMember
		if err := rows.Scan(&m.FamilyID, &m.UserID, &m.Role, &m.JoinedAt, &m.Nickname, &m.Email); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func (r *Families) AddMember(ctx context.Context, familyID, userID uuid.UUID, role domain.Role, at time.Time) error {
	_, err := r.Pool.Exec(ctx, `INSERT INTO family_members(family_id,user_id,role,joined_at) VALUES ($1,$2,$3,$4)
		ON CONFLICT DO NOTHING`, familyID, userID, role, at)
	return err
}

func (r *Families) RemoveMember(ctx context.Context, familyID, userID uuid.UUID) error {
	_, err := r.Pool.Exec(ctx, `DELETE FROM family_members WHERE family_id=$1 AND user_id=$2`, familyID, userID)
	return err
}

func (r *Families) CreateInvite(ctx context.Context, inv *domain.FamilyInvite) error {
	_, err := r.Pool.Exec(ctx, `INSERT INTO family_invites(id,family_id,code,role,expires_at) VALUES ($1,$2,$3,$4,$5)`,
		inv.ID, inv.FamilyID, inv.Code, inv.Role, inv.ExpiresAt)
	return err
}

func (r *Families) InviteByCode(ctx context.Context, code string) (*domain.FamilyInvite, error) {
	var inv domain.FamilyInvite
	err := r.Pool.QueryRow(ctx, `SELECT id,family_id,code,role,expires_at,used_by,used_at FROM family_invites WHERE code=$1`, code).
		Scan(&inv.ID, &inv.FamilyID, &inv.Code, &inv.Role, &inv.ExpiresAt, &inv.UsedBy, &inv.UsedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return &inv, err
}

func (r *Families) ConsumeInvite(ctx context.Context, id, userID uuid.UUID, at time.Time) error {
	tag, err := r.Pool.Exec(ctx, `UPDATE family_invites SET used_by=$2, used_at=$3 WHERE id=$1 AND used_at IS NULL`, id, userID, at)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrInviteUsed
	}
	return nil
}

func (r *Families) UpdateOwner(ctx context.Context, familyID, ownerID uuid.UUID) error {
	_, err := r.Pool.Exec(ctx, `UPDATE families SET owner_id=$2 WHERE id=$1`, familyID, ownerID)
	return err
}
