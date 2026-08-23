package repo

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"gopuppy/internal/clock"
	"gopuppy/internal/domain"
)

type Checkins struct{ Pool *pgxpool.Pool }

func (r *Checkins) Upsert(ctx context.Context, c *domain.DailyCheckin) error {
	_, err := r.Pool.Exec(ctx, `INSERT INTO daily_checkins(id,pet_id,checkin_date,type,slot,done_by,done_at,revoked_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,NULL)
		ON CONFLICT (pet_id, checkin_date, type, slot)
		DO UPDATE SET done_by=EXCLUDED.done_by, done_at=EXCLUDED.done_at, revoked_at=NULL, id=daily_checkins.id`,
		c.ID, c.PetID, clock.DateOf(c.CheckinDate), c.Type, c.Slot, c.DoneBy, c.DoneAt)
	return err
}

func (r *Checkins) Revoke(ctx context.Context, petID uuid.UUID, day time.Time, typ domain.CheckinType, slot domain.Slot, at time.Time) error {
	_, err := r.Pool.Exec(ctx, `UPDATE daily_checkins SET revoked_at=$5
		WHERE pet_id=$1 AND checkin_date=$2 AND type=$3 AND slot=$4`,
		petID, clock.DateOf(day), typ, slot, at)
	return err
}

func (r *Checkins) ListDay(ctx context.Context, petID uuid.UUID, day time.Time) ([]domain.DailyCheckin, error) {
	rows, err := r.Pool.Query(ctx, `SELECT c.id,c.pet_id,c.checkin_date,c.type,c.slot,c.done_by,u.nickname,c.done_at,c.revoked_at
		FROM daily_checkins c JOIN users u ON u.id=c.done_by
		WHERE c.pet_id=$1 AND c.checkin_date=$2 ORDER BY c.done_at`, petID, clock.DateOf(day))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.DailyCheckin
	for rows.Next() {
		var c domain.DailyCheckin
		if err := rows.Scan(&c.ID, &c.PetID, &c.CheckinDate, &c.Type, &c.Slot, &c.DoneBy, &c.DoneByName, &c.DoneAt, &c.RevokedAt); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *Checkins) CountActive(ctx context.Context, petID uuid.UUID, day time.Time, typ domain.CheckinType, slot domain.Slot) (int, error) {
	var n int
	err := r.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM daily_checkins
		WHERE pet_id=$1 AND checkin_date=$2 AND type=$3 AND slot=$4 AND revoked_at IS NULL`,
		petID, clock.DateOf(day), typ, slot).Scan(&n)
	return n, err
}
