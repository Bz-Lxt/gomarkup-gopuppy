export type Role = 'OWNER' | 'CAREGIVER' | 'VIEWER'
export type CheckinType = 'FEED' | 'MEDICINE'
export type Slot = 'MORNING' | 'NOON' | 'NIGHT'
export type EventCategory =
  | 'VACCINE'
  | 'DEWORM'
  | 'SURGERY'
  | 'CHECKUP'
  | 'SYMPTOM'
  | 'MEDICATION'
  | 'OTHER'
export type Severity = 'MILD' | 'MODERATE' | 'SEVERE' | ''
export type ReminderKind = 'VACCINE' | 'DEWORM' | 'MEDICINE' | 'CHECKUP'
export type Channel = 'EMAIL' | 'WECOM_BOT' | 'WEBHOOK'
export type ExpenseCategory = 'FOOD' | 'MEDICAL' | 'TOY' | 'GROOMING' | 'INSURANCE' | 'OTHER'
export type MediaKind = 'PHOTO' | 'MEDICAL_RECORD' | 'REPORT_PDF'
export type NotifyStatus = 'PENDING' | 'SENT' | 'FAILED' | 'PERMANENT_FAILURE'
export type NotifyKind = 'DUE' | 'ADVANCE'
export type Species = 'CAT' | 'DOG' | 'OTHER'
export type Gender = 'MALE' | 'FEMALE' | 'UNKNOWN'

export interface Envelope<T> {
  code: string
  message: string
  request_id: string
  data: T
}

export interface Tokens {
  access_token: string
  refresh_token: string
  expires_at: string
}

export interface User {
  id: string
  email: string
  nickname: string
  avatar_url: string
  created_at: string
}

export interface AuthPayload {
  user: User
  tokens: Tokens
}

export interface Family {
  id: string
  name: string
  owner_id: string
  created_at: string
}

export interface FamilyMember {
  family_id: string
  user_id: string
  role: Role
  nickname?: string
  email?: string
  joined_at: string
}

export interface FamilyInvite {
  id: string
  family_id: string
  code: string
  role: Role
  expires_at: string
}

export interface Age {
  years: number
  months: number
  days: number
  total_days: number
}

export interface Pet {
  id: string
  family_id: string
  name: string
  species: Species
  breed: string
  gender: Gender
  birthday: string
  avatar_key: string
  neutered: boolean
  chip_no: string
  weight_min?: number | null
  weight_max?: number | null
  note: string
  archived_at?: string | null
  created_at: string
  age: Age
}

export interface DailyCheckin {
  id: string
  pet_id: string
  checkin_date: string
  type: CheckinType
  slot: Slot
  done_by: string
  done_by_name: string
  done_at: string
  revoked_at?: string | null
}

export interface HealthEvent {
  id: string
  pet_id: string
  category: EventCategory
  title: string
  description: string
  occurred_at: string
  clinic: string
  severity?: Severity
  treated: boolean
  amount_cents?: number | null
  created_by: string
  created_at: string
}

export interface ReminderRule {
  id: string
  pet_id: string
  kind: ReminderKind
  title: string
  cycle_days: number
  last_done_at: string
  next_due_at: string
  advance_days: number
  channels: Channel[]
  enabled: boolean
  created_at: string
}

export interface NotificationLog {
  id: string
  rule_id: string
  pet_id: string
  due_date: string
  channel: Channel
  kind: NotifyKind
  status: NotifyStatus
  attempt: number
  error?: string
  title?: string
  scheduled_at: string
  sent_at?: string | null
}

export interface WeightPoint {
  month: string
  avg_kg: number
  min_kg: number
  max_kg: number
  anomaly: boolean
}

export interface ExpenseMonthBucket {
  month: string
  by_category: Record<string, number>
  total_cents: number
}

export interface CategoryShare {
  category: string
  cents: number
  percent: number
}

export interface FinanceSummary {
  month_total_cents: number
  year_total_cents: number
  top3: CategoryShare[]
  weight_series: WeightPoint[]
  expense_series: ExpenseMonthBucket[]
  pie: CategoryShare[]
  weight_min?: number | null
  weight_max?: number | null
}

export interface MediaFile {
  id: string
  family_id: string
  pet_id: string
  kind: MediaKind
  filename: string
  mime: string
  size_bytes: number
  created_at: string
}

export interface WSMessage {
  type: string
  family_id: string
  pet_id?: string
  payload?: {
    date?: string
    items?: DailyCheckin[]
  }
  at: string
}
