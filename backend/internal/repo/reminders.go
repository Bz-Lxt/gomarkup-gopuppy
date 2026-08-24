package repo

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"gopuppy/internal/clock"
	"gopuppy/internal/domain"
)

type Reminders struct{ Pool *pgxpool.Pool }

func (r *Reminders) Create(ctx context.Context, rule *domain.ReminderRule) error {
	_, err := r.Pool.Exec(ctx, `INSERT INTO reminder_rules(id,pet_id,kind,title,cycle_days,last_done_at,next_due_at,advance_days,channels,enabled,created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		rule.ID, rule.PetID, rule.Kind, rule.Title, rule.CycleDays, rule.LastDoneAt,
		clock.DateOf(rule.NextDueAt), rule.AdvanceDays, channelsToText(rule.Channels), rule.Enabled, rule.CreatedAt)
	return err
}

func (r *Reminders) Update(ctx context.Context, rule *domain.ReminderRule) error {
	_, err := r.Pool.Exec(ctx, `UPDATE reminder_rules SET title=$2,cycle_days=$3,last_done_at=$4,next_due_at=$5,advance_days=$6,channels=$7,enabled=$8 WHERE id=$1`,
		rule.ID, rule.Title, rule.CycleDays, rule.LastDoneAt, clock.DateOf(rule.NextDueAt), rule.AdvanceDays, channelsToText(rule.Channels), rule.Enabled)
	return err
}

func (r *Reminders) ByID(ctx context.Context, id uuid.UUID) (*domain.ReminderRule, error) {
	return scanRule(r.Pool.QueryRow(ctx, `SELECT id,pet_id,kind,title,cycle_days,last_done_at,next_due_at,advance_days,channels,enabled,created_at FROM reminder_rules WHERE id=$1`, id))
}

func (r *Reminders) ListByPet(ctx context.Context, petID uuid.UUID) ([]domain.ReminderRule, error) {
	rows, err := r.Pool.Query(ctx, `SELECT id,pet_id,kind,title,cycle_days,last_done_at,next_due_at,advance_days,channels,enabled,created_at
		FROM reminder_rules WHERE pet_id=$1 ORDER BY next_due_at`, petID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.ReminderRule
	for rows.Next() {
		rule, err := scanRule(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *rule)
	}
	return out, rows.Err()
}

func (r *Reminders) Enabled(ctx context.Context) ([]domain.ReminderRule, error) {
	rows, err := r.Pool.Query(ctx, `SELECT id,pet_id,kind,title,cycle_days,last_done_at,next_due_at,advance_days,channels,enabled,created_at
		FROM reminder_rules WHERE enabled=TRUE`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.ReminderRule
	for rows.Next() {
		rule, err := scanRule(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *rule)
	}
	return out, rows.Err()
}

func (r *Reminders) RecalcByKind(ctx context.Context, petID uuid.UUID, kind domain.ReminderKind, lastDone time.Time, next time.Time) error {
	_, err := r.Pool.Exec(ctx, `UPDATE reminder_rules SET last_done_at=$3, next_due_at=$4 WHERE pet_id=$1 AND kind=$2 AND enabled=TRUE`,
		petID, kind, lastDone, clock.DateOf(next))
	return err
}

func (r *Reminders) InsertLog(ctx context.Context, log *domain.NotificationLog) (inserted bool, err error) {
	tag, err := r.Pool.Exec(ctx, `INSERT INTO notification_logs(id,rule_id,pet_id,due_date,channel,kind,status,attempt,error,scheduled_at,sent_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		ON CONFLICT (rule_id, due_date, channel, kind) DO NOTHING`,
		log.ID, log.RuleID, log.PetID, clock.DateOf(log.DueDate), log.Channel, log.Kind, log.Status, log.Attempt, log.Error, log.ScheduledAt, log.SentAt)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() == 1, nil
}

func (r *Reminders) UpdateLog(ctx context.Context, log *domain.NotificationLog) error {
	_, err := r.Pool.Exec(ctx, `UPDATE notification_logs SET status=$2, attempt=$3, error=$4, sent_at=$5 WHERE id=$1`,
		log.ID, log.Status, log.Attempt, log.Error, log.SentAt)
	return err
}

func (r *Reminders) Recoverable(ctx context.Context) ([]domain.NotificationLog, error) {
	rows, err := r.Pool.Query(ctx, `SELECT id,rule_id,pet_id,due_date,channel,kind,status,attempt,error,scheduled_at,sent_at
		FROM notification_logs WHERE status IN ('PENDING','FAILED') ORDER BY scheduled_at`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.NotificationLog
	for rows.Next() {
		var l domain.NotificationLog
		if err := rows.Scan(&l.ID, &l.RuleID, &l.PetID, &l.DueDate, &l.Channel, &l.Kind, &l.Status, &l.Attempt, &l.Error, &l.ScheduledAt, &l.SentAt); err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	return out, rows.Err()
}

func (r *Reminders) LogsForFamilyPets(ctx context.Context, petIDs []uuid.UUID) ([]domain.NotificationLog, error) {
	if len(petIDs) == 0 {
		return nil, nil
	}
	rows, err := r.Pool.Query(ctx, `SELECT n.id,n.rule_id,n.pet_id,n.due_date,n.channel,n.kind,n.status,n.attempt,n.error,n.scheduled_at,n.sent_at,COALESCE(rr.title,'')
		FROM notification_logs n LEFT JOIN reminder_rules rr ON rr.id=n.rule_id
		WHERE n.pet_id = ANY($1) ORDER BY n.scheduled_at DESC LIMIT 200`, petIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.NotificationLog
	for rows.Next() {
		var l domain.NotificationLog
		if err := rows.Scan(&l.ID, &l.RuleID, &l.PetID, &l.DueDate, &l.Channel, &l.Kind, &l.Status, &l.Attempt, &l.Error, &l.ScheduledAt, &l.SentAt, &l.Title); err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	return out, rows.Err()
}

func (r *Reminders) LogByID(ctx context.Context, id uuid.UUID) (*domain.NotificationLog, error) {
	var l domain.NotificationLog
	err := r.Pool.QueryRow(ctx, `SELECT id,rule_id,pet_id,due_date,channel,kind,status,attempt,error,scheduled_at,sent_at
		FROM notification_logs WHERE id=$1`, id).
		Scan(&l.ID, &l.RuleID, &l.PetID, &l.DueDate, &l.Channel, &l.Kind, &l.Status, &l.Attempt, &l.Error, &l.ScheduledAt, &l.SentAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return &l, err
}

type ruleScanner interface {
	Scan(dest ...any) error
}

func scanRule(s ruleScanner) (*domain.ReminderRule, error) {
	var rule domain.ReminderRule
	var ch []string
	err := s.Scan(&rule.ID, &rule.PetID, &rule.Kind, &rule.Title, &rule.CycleDays, &rule.LastDoneAt, &rule.NextDueAt, &rule.AdvanceDays, &ch, &rule.Enabled, &rule.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	for _, c := range ch {
		rule.Channels = append(rule.Channels, domain.Channel(c))
	}
	return &rule, nil
}

func channelsToText(ch []domain.Channel) []string {
	out := make([]string, 0, len(ch))
	for _, c := range ch {
		out = append(out, string(c))
	}
	return out
}
