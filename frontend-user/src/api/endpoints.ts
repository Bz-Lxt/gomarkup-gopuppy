import { asList, http, unwrap } from '@/api/http'
import type {
  AuthPayload,
  Channel,
  CheckinType,
  DailyCheckin,
  EventCategory,
  ExpenseCategory,
  Family,
  FamilyInvite,
  FamilyMember,
  FinanceSummary,
  HealthEvent,
  MediaFile,
  MediaKind,
  NotificationLog,
  Pet,
  ReminderKind,
  ReminderRule,
  Role,
  Severity,
  Slot,
  Tokens,
  User,
} from '@/types/models'

export const authApi = {
  login: (email: string, password: string) =>
    unwrap<AuthPayload>(http.post('/auth/login', { email, password })),
  register: (email: string, password: string, nickname: string) =>
    unwrap<AuthPayload>(http.post('/auth/register', { email, password, nickname })),
  refresh: (refresh_token: string) => unwrap<Tokens>(http.post('/auth/refresh', { refresh_token })),
  me: () => unwrap<User>(http.get('/me')),
}

export const familyApi = {
  list: async () => asList(await unwrap<Family[]>(http.get('/families'))),
  create: (name: string) => unwrap<Family>(http.post('/families', { name })),
  members: async (familyId: string) =>
    asList(await unwrap<FamilyMember[]>(http.get(`/families/${familyId}/members`))),
  invite: (familyId: string, role: Role) =>
    unwrap<FamilyInvite>(http.post(`/families/${familyId}/invites`, { role })),
  join: (code: string) => unwrap<Family>(http.post('/families/join', { code })),
  removeMember: (familyId: string, userId: string) =>
    unwrap<{ status: string }>(http.delete(`/families/${familyId}/members/${userId}`)),
  notifications: async (familyId: string) =>
    asList(await unwrap<NotificationLog[]>(http.get(`/families/${familyId}/notifications`))),
}

export const petApi = {
  list: async (familyId: string) => asList(await unwrap<Pet[]>(http.get(`/families/${familyId}/pets`))),
  get: (id: string) => unwrap<Pet>(http.get(`/pets/${id}`)),
  create: (familyId: string, body: Record<string, unknown>) =>
    unwrap<Pet>(http.post(`/families/${familyId}/pets`, body)),
  archive: (id: string) => unwrap<{ status: string }>(http.delete(`/pets/${id}`)),
}

export const checkinApi = {
  today: async (petId: string) =>
    asList(await unwrap<DailyCheckin[]>(http.get(`/pets/${petId}/checkins/today`))),
  toggle: async (petId: string, type: CheckinType, slot: Slot, done: boolean) =>
    asList(await unwrap<DailyCheckin[]>(http.post(`/pets/${petId}/checkins`, { type, slot, done }))),
}

export const eventApi = {
  list: async (petId: string, category?: string, year?: number) => {
    const params: Record<string, string | number> = {}
    if (category) params.category = category
    if (year) params.year = year
    return asList(await unwrap<HealthEvent[]>(http.get(`/pets/${petId}/events`, { params })))
  },
  create: (
    petId: string,
    body: {
      category: EventCategory
      title: string
      occurred_at: string
      description?: string
      clinic?: string
      severity?: Severity
      treated?: boolean
      amount_cents?: number | null
    },
  ) => unwrap<HealthEvent>(http.post(`/pets/${petId}/events`, body)),
}

export const financeApi = {
  summary: (petId: string) => unwrap<FinanceSummary>(http.get(`/pets/${petId}/finance`)),
  addWeight: (petId: string, weight_kg: number, measured_at: string, note: string) =>
    unwrap(http.post(`/pets/${petId}/weights`, { weight_kg, measured_at, note })),
  addExpense: (petId: string, category: ExpenseCategory, amount_cents: number, spent_at: string, note: string) =>
    unwrap(http.post(`/pets/${petId}/expenses`, { category, amount_cents, spent_at, note })),
}

export const mediaApi = {
  list: async (petId: string, kind?: MediaKind) =>
    asList(await unwrap<MediaFile[]>(http.get(`/pets/${petId}/media`, { params: kind ? { kind } : {} }))),
  upload: (petId: string, file: File, kind: MediaKind) => {
    const fd = new FormData()
    fd.append('file', file)
    fd.append('kind', kind)
    return unwrap<MediaFile>(http.post(`/pets/${petId}/media`, fd))
  },
  file: (id: string) => http.get(`/media/${id}/file`, { responseType: 'blob' }),
}

export const reminderApi = {
  list: async (petId: string) => asList(await unwrap<ReminderRule[]>(http.get(`/pets/${petId}/reminders`))),
  create: (
    petId: string,
    body: {
      kind: ReminderKind
      title: string
      cycle_days: number
      last_done_at: string
      advance_days: number
      channels: Channel[]
    },
  ) => unwrap<ReminderRule>(http.post(`/pets/${petId}/reminders`, body)),
  replay: (logId: string) => unwrap<{ status: string }>(http.post(`/notifications/${logId}/replay`)),
}
