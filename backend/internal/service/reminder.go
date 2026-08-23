package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gopuppy/internal/clock"
	"gopuppy/internal/domain"
	"gopuppy/internal/notifier"
	"gopuppy/internal/reminder"
	"gopuppy/internal/repo"
)

type Reminder struct {
	Rules    *repo.Reminders
	Family   *Family
	Pets     *repo.Pets
	Users    *repo.Users
	Families *repo.Families
	Sender   *notifier.Fanout
}

type RuleInput struct {
	Kind        domain.ReminderKind `json:"kind"`
	Title       string              `json:"title"`
	CycleDays   int                 `json:"cycle_days"`
	LastDoneAt  string              `json:"last_done_at"`
	AdvanceDays int                 `json:"advance_days"`
	Channels    []domain.Channel    `json:"channels"`
	Enabled     *bool               `json:"enabled"`
}

func (s *Reminder) Create(ctx context.Context, userID, petID uuid.UUID, in RuleInput) (*domain.ReminderRule, error) {
	if _, _, err := s.Family.MustWritePet(ctx, userID, petID); err != nil {
		return nil, err
	}
	rule, err := buildRule(in)
	if err != nil {
		return nil, err
	}
	rule.ID = uuid.New()
	rule.PetID = petID
	rule.CreatedAt = clock.Now()
	if err := s.Rules.Create(ctx, rule); err != nil {
		return nil, err
	}
	return rule, nil
}

func (s *Reminder) List(ctx context.Context, userID, petID uuid.UUID) ([]domain.ReminderRule, error) {
	if _, _, err := s.Family.MustReadPet(ctx, userID, petID); err != nil {
		return nil, err
	}
	return s.Rules.ListByPet(ctx, petID)
}

func (s *Reminder) Logs(ctx context.Context, userID, familyID uuid.UUID) ([]domain.NotificationLog, error) {
	if _, err := s.Family.MustMember(ctx, userID, familyID); err != nil {
		return nil, err
	}
	pets, err := s.Pets.ListByFamily(ctx, familyID, true)
	if err != nil {
		return nil, err
	}
	ids := make([]uuid.UUID, 0, len(pets))
	for _, p := range pets {
		ids = append(ids, p.ID)
	}
	return s.Rules.LogsForFamilyPets(ctx, ids)
}

func (s *Reminder) Replay(ctx context.Context, userID, logID uuid.UUID) error {
	log, err := s.Rules.LogByID(ctx, logID)
	if err != nil {
		return domain.ErrNotFound
	}
	if _, _, err := s.Family.MustWritePet(ctx, userID, log.PetID); err != nil {
		return err
	}
	return s.deliver(ctx, log)
}

func (s *Reminder) ScanDue(ctx context.Context, day time.Time) error {
	rules, err := s.Rules.Enabled(ctx)
	if err != nil {
		return err
	}
	for _, rule := range rules {
		if reminder.IsDueOn(rule, day) {
			s.enqueue(ctx, rule, day, domain.NotifyDue)
		}
		if reminder.IsAdvanceOn(rule, day) {
			s.enqueue(ctx, rule, day, domain.NotifyAdvance)
		}
	}
	return nil
}

func (s *Reminder) Recover(ctx context.Context) error {
	logs, err := s.Rules.Recoverable(ctx)
	if err != nil {
		return err
	}
	for i := range logs {
		_ = s.deliver(ctx, &logs[i])
	}
	return nil
}

func (s *Reminder) enqueue(ctx context.Context, rule domain.ReminderRule, day time.Time, kind domain.NotifyKind) {
	for _, ch := range rule.Channels {
		log := &domain.NotificationLog{
			ID: uuid.New(), RuleID: rule.ID, PetID: rule.PetID, DueDate: day,
			Channel: ch, Kind: kind, Status: domain.NotifyPending, ScheduledAt: clock.Now(), Title: rule.Title,
		}
		ok, err := s.Rules.InsertLog(ctx, log)
		if err != nil || !ok {
			continue
		}
		_ = s.deliver(ctx, log)
	}
}

func (s *Reminder) deliver(ctx context.Context, log *domain.NotificationLog) error {
	if log.Status == domain.NotifySent || log.Status == domain.NotifyPermanentFailure {
		return nil
	}
	if log.Attempt >= 3 && log.Status == domain.NotifyFailed {
		log.Status = domain.NotifyPermanentFailure
		log.Error = "max attempts"
		return s.persistLog(ctx, log)
	}
	rule, err := s.Rules.ByID(ctx, log.RuleID)
	if err != nil {
		return err
	}
	pet, err := s.Pets.ByID(ctx, log.PetID)
	if err != nil {
		return err
	}
	members, _ := s.Families.Members(ctx, pet.FamilyID)
	to := ""
	for _, m := range members {
		if m.Role == domain.RoleOwner {
			to = m.Email
			break
		}
	}
	kindLabel := "到期提醒"
	if log.Kind == domain.NotifyAdvance {
		kindLabel = "提前提醒"
	}
	msg := notifier.Message{
		Channel: log.Channel,
		ToEmail: to,
		Title:   fmt.Sprintf("[GoPuppy] %s · %s", kindLabel, rule.Title),
		Body:    fmt.Sprintf("宠物提醒：%s（%s）将于 %s 到期。请及时处理。", pet.Name, rule.Title, clock.FormatDate(rule.NextDueAt)),
	}
	// If the caller's context is already canceled (e.g. the upstream HTTP
	// request was aborted), don't burn an attempt against a dead context.
	// Persist the log as FAILED so Recover() can retry it later with a fresh
	// context, instead of recording PERMANENT_FAILURE.
	if ctxErr := ctx.Err(); ctxErr != nil {
		if log.Status != domain.NotifyFailed {
			log.Status = domain.NotifyFailed
		}
		log.Error = ctxErr.Error()
		return s.persistLog(ctx, log)
	}
	err = s.Sender.Send(ctx, msg)
	log.Attempt++
	if err == nil {
		now := clock.Now()
		log.Status = domain.NotifySent
		log.SentAt = &now
		log.Error = ""
		return s.persistLog(ctx, log)
	}
	transient, st := reminder.ClassifyDeliveryError(err)
	log.Status = st
	log.Error = err.Error()
	// Only retry in-process when the context is still alive. A canceled
	// context would fail immediately again, so leave the log as FAILED and
	// let Recover() retry it with a fresh context later.
	if transient && log.Attempt < 3 && ctx.Err() == nil {
		time.Sleep(reminder.Backoff(log.Attempt))
		return s.deliver(ctx, log)
	}
	if !transient {
		log.Status = domain.NotifyPermanentFailure
	}
	return s.persistLog(ctx, log)
}

// persistLog writes the notification log using a detached context so that
// the status/error of a delivery is always recorded even when the caller's
// context has been canceled (e.g. upstream request aborted). Without this,
// a cancellation would leave the in-memory log mutated but unsaved, so
// Recover() would never pick it up.
func (s *Reminder) persistLog(ctx context.Context, log *domain.NotificationLog) error {
	return s.Rules.UpdateLog(context.WithoutCancel(ctx), log)
}

func buildRule(in RuleInput) (*domain.ReminderRule, error) {
	if !in.Kind.Valid() {
		return nil, fmt.Errorf("%w: kind", domain.ErrValidation)
	}
	title := strings.TrimSpace(in.Title)
	if title == "" {
		return nil, fmt.Errorf("%w: title", domain.ErrValidation)
	}
	if in.CycleDays <= 0 || in.CycleDays > 3650 {
		return nil, fmt.Errorf("%w: cycle_days", domain.ErrValidation)
	}
	last, err := parseDateTime(in.LastDoneAt)
	if err != nil {
		last = clock.Today()
	}
	if in.AdvanceDays < 0 || in.AdvanceDays > 30 {
		in.AdvanceDays = 3
	}
	if in.AdvanceDays == 0 {
		in.AdvanceDays = 3
	}
	if len(in.Channels) == 0 {
		in.Channels = []domain.Channel{domain.ChannelEmail}
	}
	for _, c := range in.Channels {
		if !c.Valid() {
			return nil, fmt.Errorf("%w: channel", domain.ErrValidation)
		}
	}
	enabled := true
	if in.Enabled != nil {
		enabled = *in.Enabled
	}
	rule := &domain.ReminderRule{
		Kind: in.Kind, Title: title, CycleDays: in.CycleDays, LastDoneAt: last,
		AdvanceDays: in.AdvanceDays, Channels: in.Channels, Enabled: enabled,
	}
	reminder.Recalc(rule, last)
	return rule, nil
}
