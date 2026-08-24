package repo

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"gopuppy/internal/clock"
	"gopuppy/internal/domain"
)

type Finance struct{ Pool *pgxpool.Pool }

func (r *Finance) AddWeight(ctx context.Context, w *domain.WeightRecord) error {
	_, err := r.Pool.Exec(ctx, `INSERT INTO weight_records(id,pet_id,weight_kg,measured_at,note,created_by) VALUES ($1,$2,$3,$4,$5,$6)`,
		w.ID, w.PetID, w.WeightKG, w.MeasuredAt, w.Note, w.CreatedBy)
	return err
}

func (r *Finance) AddExpense(ctx context.Context, e *domain.Expense) error {
	_, err := r.Pool.Exec(ctx, `INSERT INTO expenses(id,pet_id,category,amount_cents,spent_at,note,created_by) VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		e.ID, e.PetID, e.Category, e.AmountCents, e.SpentAt, e.Note, e.CreatedBy)
	return err
}

func (r *Finance) WeightsSince(ctx context.Context, petID uuid.UUID, since time.Time) ([]domain.WeightRecord, error) {
	rows, err := r.Pool.Query(ctx, `SELECT id,pet_id,weight_kg,measured_at,note,created_by FROM weight_records
		WHERE pet_id=$1 AND measured_at>=$2 ORDER BY measured_at`, petID, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.WeightRecord
	for rows.Next() {
		var w domain.WeightRecord
		if err := rows.Scan(&w.ID, &w.PetID, &w.WeightKG, &w.MeasuredAt, &w.Note, &w.CreatedBy); err != nil {
			return nil, err
		}
		out = append(out, w)
	}
	return out, rows.Err()
}

func (r *Finance) ExpensesSince(ctx context.Context, petID uuid.UUID, since time.Time) ([]domain.Expense, error) {
	rows, err := r.Pool.Query(ctx, `SELECT id,pet_id,category,amount_cents,spent_at,note,created_by FROM expenses
		WHERE pet_id=$1 AND spent_at>=$2 ORDER BY spent_at`, petID, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Expense
	for rows.Next() {
		var e domain.Expense
		if err := rows.Scan(&e.ID, &e.PetID, &e.Category, &e.AmountCents, &e.SpentAt, &e.Note, &e.CreatedBy); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

func (r *Finance) SumCents(items []domain.Expense) int64 {
	var n int64
	for _, e := range items {
		n += e.AmountCents
	}
	return n
}

func MonthStart12() time.Time {
	t := clock.Today()
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, clock.Beijing).AddDate(0, -11, 0)
}
