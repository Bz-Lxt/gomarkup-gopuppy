package domain

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Nickname     string    `json:"nickname"`
	AvatarURL    string    `json:"avatar_url"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Family struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	OwnerID   uuid.UUID `json:"owner_id"`
	CreatedAt time.Time `json:"created_at"`
}

type FamilyMember struct {
	FamilyID uuid.UUID `json:"family_id"`
	UserID   uuid.UUID `json:"user_id"`
	Role     Role      `json:"role"`
	Nickname string    `json:"nickname,omitempty"`
	Email    string    `json:"email,omitempty"`
	JoinedAt time.Time `json:"joined_at"`
}

type FamilyInvite struct {
	ID        uuid.UUID  `json:"id"`
	FamilyID  uuid.UUID  `json:"family_id"`
	Code      string     `json:"code"`
	Role      Role       `json:"role"`
	ExpiresAt time.Time  `json:"expires_at"`
	UsedBy    *uuid.UUID `json:"used_by,omitempty"`
	UsedAt    *time.Time `json:"used_at,omitempty"`
}

type Pet struct {
	ID         uuid.UUID  `json:"id"`
	FamilyID   uuid.UUID  `json:"family_id"`
	Name       string     `json:"name"`
	Species    Species    `json:"species"`
	Breed      string     `json:"breed"`
	Gender     Gender     `json:"gender"`
	Birthday   time.Time  `json:"birthday"`
	AvatarKey  string     `json:"avatar_key"`
	Neutered   bool       `json:"neutered"`
	ChipNo     string     `json:"chip_no"`
	WeightMin  *float64   `json:"weight_min"`
	WeightMax  *float64   `json:"weight_max"`
	Note       string     `json:"note"`
	ArchivedAt *time.Time `json:"archived_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	Age        Age        `json:"age"`
}

type DailyCheckin struct {
	ID          uuid.UUID   `json:"id"`
	PetID       uuid.UUID   `json:"pet_id"`
	CheckinDate time.Time   `json:"checkin_date"`
	Type        CheckinType `json:"type"`
	Slot        Slot        `json:"slot"`
	DoneBy      uuid.UUID   `json:"done_by"`
	DoneByName  string      `json:"done_by_name"`
	DoneAt      time.Time   `json:"done_at"`
	RevokedAt   *time.Time  `json:"revoked_at,omitempty"`
}

type HealthEvent struct {
	ID          uuid.UUID     `json:"id"`
	PetID       uuid.UUID     `json:"pet_id"`
	Category    EventCategory `json:"category"`
	Title       string        `json:"title"`
	Description string        `json:"description"`
	OccurredAt  time.Time     `json:"occurred_at"`
	Clinic      string        `json:"clinic"`
	Severity    Severity      `json:"severity,omitempty"`
	Treated     bool          `json:"treated"`
	AmountCents *int64        `json:"amount_cents,omitempty"`
	CreatedBy   uuid.UUID     `json:"created_by"`
	CreatedAt   time.Time     `json:"created_at"`
	MediaIDs    []uuid.UUID   `json:"media_ids,omitempty"`
}

type ReminderRule struct {
	ID          uuid.UUID    `json:"id"`
	PetID       uuid.UUID    `json:"pet_id"`
	Kind        ReminderKind `json:"kind"`
	Title       string       `json:"title"`
	CycleDays   int          `json:"cycle_days"`
	LastDoneAt  time.Time    `json:"last_done_at"`
	NextDueAt   time.Time    `json:"next_due_at"`
	AdvanceDays int          `json:"advance_days"`
	Channels    []Channel    `json:"channels"`
	Enabled     bool         `json:"enabled"`
	CreatedAt   time.Time    `json:"created_at"`
}

type NotificationLog struct {
	ID          uuid.UUID    `json:"id"`
	RuleID      uuid.UUID    `json:"rule_id"`
	PetID       uuid.UUID    `json:"pet_id"`
	DueDate     time.Time    `json:"due_date"`
	Channel     Channel      `json:"channel"`
	Kind        NotifyKind   `json:"kind"`
	Status      NotifyStatus `json:"status"`
	Attempt     int          `json:"attempt"`
	Error       string       `json:"error,omitempty"`
	Title       string       `json:"title,omitempty"`
	ScheduledAt time.Time    `json:"scheduled_at"`
	SentAt      *time.Time   `json:"sent_at,omitempty"`
}

type WeightRecord struct {
	ID         uuid.UUID `json:"id"`
	PetID      uuid.UUID `json:"pet_id"`
	WeightKG   float64   `json:"weight_kg"`
	MeasuredAt time.Time `json:"measured_at"`
	Note       string    `json:"note"`
	CreatedBy  uuid.UUID `json:"created_by"`
}

type Expense struct {
	ID          uuid.UUID       `json:"id"`
	PetID       uuid.UUID       `json:"pet_id"`
	Category    ExpenseCategory `json:"category"`
	AmountCents int64           `json:"amount_cents"`
	SpentAt     time.Time       `json:"spent_at"`
	Note        string          `json:"note"`
	CreatedBy   uuid.UUID       `json:"created_by"`
}

type MediaFile struct {
	ID            uuid.UUID     `json:"id"`
	FamilyID      uuid.UUID     `json:"family_id"`
	PetID         uuid.UUID     `json:"pet_id"`
	Kind          MediaKind     `json:"kind"`
	StorageDriver StorageDriver `json:"storage_driver"`
	ObjectKey     string        `json:"object_key"`
	Filename      string        `json:"filename"`
	MIME          string        `json:"mime"`
	SizeBytes     int64         `json:"size_bytes"`
	SHA256        string        `json:"sha256"`
	UploadedBy    uuid.UUID     `json:"uploaded_by"`
	CreatedAt     time.Time     `json:"created_at"`
}

type WeightPoint struct {
	Month    string  `json:"month"`
	AvgKG    float64 `json:"avg_kg"`
	MinKG    float64 `json:"min_kg"`
	MaxKG    float64 `json:"max_kg"`
	Anomaly  bool    `json:"anomaly"`
}

type ExpenseMonthBucket struct {
	Month    string           `json:"month"`
	ByCat    map[string]int64 `json:"by_category"`
	Total    int64            `json:"total_cents"`
}

type FinanceSummary struct {
	MonthTotalCents int64            `json:"month_total_cents"`
	YearTotalCents  int64            `json:"year_total_cents"`
	Top3            []CategoryShare  `json:"top3"`
	WeightSeries    []WeightPoint    `json:"weight_series"`
	ExpenseSeries   []ExpenseMonthBucket `json:"expense_series"`
	Pie             []CategoryShare  `json:"pie"`
	WeightMin       *float64         `json:"weight_min,omitempty"`
	WeightMax       *float64         `json:"weight_max,omitempty"`
}

type CategoryShare struct {
	Category string `json:"category"`
	Cents    int64  `json:"cents"`
	Percent  float64 `json:"percent"`
}

type WSMessage struct {
	Type      string      `json:"type"`
	FamilyID  uuid.UUID   `json:"family_id"`
	PetID     uuid.UUID   `json:"pet_id,omitempty"`
	Payload   interface{} `json:"payload"`
	At        time.Time   `json:"at"`
}
