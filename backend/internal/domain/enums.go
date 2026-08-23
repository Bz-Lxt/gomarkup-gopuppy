package domain

type Role string

const (
	RoleOwner     Role = "OWNER"
	RoleCaregiver Role = "CAREGIVER"
	RoleViewer    Role = "VIEWER"
)

func (r Role) Valid() bool {
	switch r {
	case RoleOwner, RoleCaregiver, RoleViewer:
		return true
	}
	return false
}

func (r Role) CanWrite() bool {
	return r == RoleOwner || r == RoleCaregiver
}

func (r Role) CanManageMembers() bool {
	return r == RoleOwner
}

func (r Role) CanDeletePet() bool {
	return r == RoleOwner
}

type CheckinType string

const (
	CheckinFeed     CheckinType = "FEED"
	CheckinMedicine CheckinType = "MEDICINE"
)

func (t CheckinType) Valid() bool {
	return t == CheckinFeed || t == CheckinMedicine
}

type Slot string

const (
	SlotMorning Slot = "MORNING"
	SlotNoon    Slot = "NOON"
	SlotNight   Slot = "NIGHT"
)

func (s Slot) Valid() bool {
	return s == SlotMorning || s == SlotNoon || s == SlotNight
}

type EventCategory string

const (
	EventVaccine    EventCategory = "VACCINE"
	EventDeworm     EventCategory = "DEWORM"
	EventSurgery    EventCategory = "SURGERY"
	EventCheckup    EventCategory = "CHECKUP"
	EventSymptom    EventCategory = "SYMPTOM"
	EventMedication EventCategory = "MEDICATION"
	EventOther      EventCategory = "OTHER"
)

func (c EventCategory) Valid() bool {
	switch c {
	case EventVaccine, EventDeworm, EventSurgery, EventCheckup, EventSymptom, EventMedication, EventOther:
		return true
	}
	return false
}

func (c EventCategory) ReminderKind() (ReminderKind, bool) {
	switch c {
	case EventVaccine:
		return ReminderVaccine, true
	case EventDeworm:
		return ReminderDeworm, true
	case EventCheckup:
		return ReminderCheckup, true
	case EventMedication:
		return ReminderMedicine, true
	}
	return "", false
}

type Severity string

const (
	SeverityMild     Severity = "MILD"
	SeverityModerate Severity = "MODERATE"
	SeveritySevere   Severity = "SEVERE"
)

func (s Severity) Valid() bool {
	return s == "" || s == SeverityMild || s == SeverityModerate || s == SeveritySevere
}

type ReminderKind string

const (
	ReminderVaccine ReminderKind = "VACCINE"
	ReminderDeworm  ReminderKind = "DEWORM"
	ReminderMedicine ReminderKind = "MEDICINE"
	ReminderCheckup ReminderKind = "CHECKUP"
)

func (k ReminderKind) Valid() bool {
	switch k {
	case ReminderVaccine, ReminderDeworm, ReminderMedicine, ReminderCheckup:
		return true
	}
	return false
}

type Channel string

const (
	ChannelEmail  Channel = "EMAIL"
	ChannelWecom  Channel = "WECOM_BOT"
	ChannelHook   Channel = "WEBHOOK"
)

func (c Channel) Valid() bool {
	return c == ChannelEmail || c == ChannelWecom || c == ChannelHook
}

type ExpenseCategory string

const (
	ExpenseFood      ExpenseCategory = "FOOD"
	ExpenseMedical   ExpenseCategory = "MEDICAL"
	ExpenseToy       ExpenseCategory = "TOY"
	ExpenseGrooming  ExpenseCategory = "GROOMING"
	ExpenseInsurance ExpenseCategory = "INSURANCE"
	ExpenseOther     ExpenseCategory = "OTHER"
)

func (c ExpenseCategory) Valid() bool {
	switch c {
	case ExpenseFood, ExpenseMedical, ExpenseToy, ExpenseGrooming, ExpenseInsurance, ExpenseOther:
		return true
	}
	return false
}

type MediaKind string

const (
	MediaPhoto         MediaKind = "PHOTO"
	MediaMedicalRecord MediaKind = "MEDICAL_RECORD"
	MediaReportPDF     MediaKind = "REPORT_PDF"
)

func (k MediaKind) Valid() bool {
	return k == MediaPhoto || k == MediaMedicalRecord || k == MediaReportPDF
}

type StorageDriver string

const (
	DriverLocal StorageDriver = "local"
	DriverOSS   StorageDriver = "oss"
	DriverCOS   StorageDriver = "cos"
)

type NotifyStatus string

const (
	NotifyPending           NotifyStatus = "PENDING"
	NotifySent              NotifyStatus = "SENT"
	NotifyFailed            NotifyStatus = "FAILED"
	NotifyPermanentFailure  NotifyStatus = "PERMANENT_FAILURE"
)

type NotifyKind string

const (
	NotifyDue     NotifyKind = "DUE"
	NotifyAdvance NotifyKind = "ADVANCE"
)

type Species string

const (
	SpeciesCat   Species = "CAT"
	SpeciesDog   Species = "DOG"
	SpeciesOther Species = "OTHER"
)

func (s Species) Valid() bool {
	return s == SpeciesCat || s == SpeciesDog || s == SpeciesOther
}

type Gender string

const (
	GenderMale   Gender = "MALE"
	GenderFemale Gender = "FEMALE"
	GenderUnknown Gender = "UNKNOWN"
)

func (g Gender) Valid() bool {
	return g == GenderMale || g == GenderFemale || g == GenderUnknown
}
