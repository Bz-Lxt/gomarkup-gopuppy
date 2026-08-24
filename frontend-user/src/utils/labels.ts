import type {
  Channel,
  CheckinType,
  EventCategory,
  ExpenseCategory,
  Gender,
  MediaKind,
  NotifyKind,
  NotifyStatus,
  ReminderKind,
  Role,
  Severity,
  Slot,
  Species,
} from '@/types/models'

export const ROLE_LABEL: Record<Role, string> = {
  OWNER: '主人',
  CAREGIVER: '共同养育',
  VIEWER: '访客',
}

export const SPECIES_LABEL: Record<Species, string> = {
  CAT: '猫',
  DOG: '狗',
  OTHER: '其他',
}

export const GENDER_LABEL: Record<Gender, string> = {
  MALE: '弟弟',
  FEMALE: '妹妹',
  UNKNOWN: '未知',
}

export const SLOT_LABEL: Record<Slot, string> = {
  MORNING: '早',
  NOON: '午',
  NIGHT: '晚',
}

export const CHECKIN_LABEL: Record<CheckinType, string> = {
  FEED: '喂食',
  MEDICINE: '吃药',
}

export const EVENT_LABEL: Record<EventCategory, string> = {
  VACCINE: '疫苗',
  DEWORM: '驱虫',
  SURGERY: '手术',
  CHECKUP: '体检',
  SYMPTOM: '症状',
  MEDICATION: '用药',
  OTHER: '其他',
}

export const EVENT_TONE: Record<EventCategory, string> = {
  VACCINE: 'bg-clay/15 text-clay ring-clay/30',
  DEWORM: 'bg-moss/15 text-moss ring-moss/30',
  SURGERY: 'bg-rose/10 text-rose ring-rose/25',
  CHECKUP: 'bg-gold/15 text-[#8A6400] ring-gold/40',
  SYMPTOM: 'bg-rose/10 text-rose ring-rose/25',
  MEDICATION: 'bg-clay/10 text-clay ring-clay/25',
  OTHER: 'bg-ink/5 text-ink/70 ring-line',
}

export const EVENT_DOT: Record<EventCategory, string> = {
  VACCINE: '#C45C26',
  DEWORM: '#3D6B4F',
  SURGERY: '#B42318',
  CHECKUP: '#E0A100',
  SYMPTOM: '#B42318',
  MEDICATION: '#C45C26',
  OTHER: '#2A2118',
}

export const SEVERITY_LABEL: Record<Exclude<Severity, ''>, string> = {
  MILD: '轻微',
  MODERATE: '中等',
  SEVERE: '严重',
}

export const EXPENSE_LABEL: Record<ExpenseCategory, string> = {
  FOOD: '口粮',
  MEDICAL: '医疗',
  TOY: '玩具',
  GROOMING: '美容',
  INSURANCE: '保险',
  OTHER: '其他',
}

export const EXPENSE_COLOR: Record<ExpenseCategory, string> = {
  FOOD: '#C45C26',
  MEDICAL: '#B42318',
  TOY: '#E0A100',
  GROOMING: '#3D6B4F',
  INSURANCE: '#6B4F3D',
  OTHER: '#8A7A68',
}

export const REMINDER_LABEL: Record<ReminderKind, string> = {
  VACCINE: '疫苗',
  DEWORM: '驱虫',
  MEDICINE: '用药',
  CHECKUP: '体检',
}

export const CHANNEL_LABEL: Record<Channel, string> = {
  EMAIL: '邮件',
  WECOM_BOT: '企业微信',
  WEBHOOK: 'Webhook',
}

export const MEDIA_LABEL: Record<MediaKind, string> = {
  PHOTO: '相册',
  MEDICAL_RECORD: '病历',
  REPORT_PDF: '体检报告',
}

export const NOTIFY_STATUS: Record<NotifyStatus, string> = {
  PENDING: '待发送',
  SENT: '已送达',
  FAILED: '失败',
  PERMANENT_FAILURE: '永久失败',
}

export const NOTIFY_KIND: Record<NotifyKind, string> = {
  DUE: '到期',
  ADVANCE: '提前',
}

export function yuan(cents: number): string {
  return `¥${(cents / 100).toFixed(2)}`
}

export function bytes(n: number): string {
  if (n < 1024) return `${n} B`
  if (n < 1024 * 1024) return `${(n / 1024).toFixed(1)} KB`
  return `${(n / 1024 / 1024).toFixed(1)} MB`
}
