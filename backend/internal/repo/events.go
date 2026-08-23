package repo

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"gopuppy/internal/domain"
)

type Events struct{ Pool *pgxpool.Pool }

func (r *Events) Create(ctx context.Context, e *domain.HealthEvent) error {
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, `INSERT INTO health_events(id,pet_id,category,title,description,occurred_at,clinic,severity,treated,amount_cents,created_by,created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
		e.ID, e.PetID, e.Category, e.Title, e.Description, e.OccurredAt, e.Clinic, e.Severity, e.Treated, e.AmountCents, e.CreatedBy, e.CreatedAt); err != nil {
		return err
	}
	for _, mid := range e.MediaIDs {
		if _, err := tx.Exec(ctx, `INSERT INTO event_attachments(event_id,media_id) VALUES ($1,$2)`, e.ID, mid); err != nil {
			return err
		}
	}
	if e.AmountCents != nil && *e.AmountCents > 0 {
		cat := domain.ExpenseMedical
		if e.Category == domain.EventOther {
			cat = domain.ExpenseOther
		}
		if _, err := tx.Exec(ctx, `INSERT INTO expenses(id,pet_id,category,amount_cents,spent_at,note,created_by)
			VALUES ($1,$2,$3,$4,$5,$6,$7)`, uuid.New(), e.PetID, cat, *e.AmountCents, e.OccurredAt, "联动自健康事件: "+e.Title, e.CreatedBy); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (r *Events) List(ctx context.Context, petID uuid.UUID, category string, year int) ([]domain.HealthEvent, error) {
	q := `SELECT id,pet_id,category,title,description,occurred_at,clinic,severity,treated,amount_cents,created_by,created_at
		FROM health_events WHERE pet_id=$1`
	args := []any{petID}
	if category != "" {
		args = append(args, category)
		q += ` AND category=$2`
	}
	if year > 0 {
		args = append(args, year)
		q += ` AND EXTRACT(YEAR FROM occurred_at AT TIME ZONE 'Asia/Shanghai')=$` + itoa(len(args))
	}
	q += ` ORDER BY occurred_at DESC`
	rows, err := r.Pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.HealthEvent
	for rows.Next() {
		var e domain.HealthEvent
		if err := rows.Scan(&e.ID, &e.PetID, &e.Category, &e.Title, &e.Description, &e.OccurredAt, &e.Clinic, &e.Severity, &e.Treated, &e.AmountCents, &e.CreatedBy, &e.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

func (r *Events) ByID(ctx context.Context, id uuid.UUID) (*domain.HealthEvent, error) {
	var e domain.HealthEvent
	err := r.Pool.QueryRow(ctx, `SELECT id,pet_id,category,title,description,occurred_at,clinic,severity,treated,amount_cents,created_by,created_at
		FROM health_events WHERE id=$1`, id).
		Scan(&e.ID, &e.PetID, &e.Category, &e.Title, &e.Description, &e.OccurredAt, &e.Clinic, &e.Severity, &e.Treated, &e.AmountCents, &e.CreatedBy, &e.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return &e, err
}

func itoa(n int) string {
	return []string{"", "1", "2", "3", "4", "5", "6"}[n]
}
