const TZ = 'Asia/Shanghai'

function pad(n: number): string {
  return String(n).padStart(2, '0')
}

export function formatDateTime(input: string | Date | null | undefined): string {
  if (!input) return '—'
  const d = typeof input === 'string' ? new Date(input) : input
  if (Number.isNaN(d.getTime())) {
    const raw = String(input)
    if (/^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}$/.test(raw)) return raw
    return raw.replace('T', ' ').slice(0, 19)
  }
  const parts = new Intl.DateTimeFormat('en-GB', {
    timeZone: TZ,
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
    hour12: false,
  }).formatToParts(d)
  const get = (t: string) => parts.find((p) => p.type === t)?.value ?? '00'
  return `${get('year')}-${get('month')}-${get('day')} ${get('hour')}:${get('minute')}:${get('second')}`
}

export function formatDate(input: string | Date | null | undefined): string {
  return formatDateTime(input).slice(0, 10)
}

export function nowBeijing(): string {
  return formatDateTime(new Date())
}

export function todayBeijing(): string {
  return nowBeijing().slice(0, 10)
}

export function formatAge(age: { years: number; months: number; days: number; total_days: number }): string {
  return `${age.years}岁 ${age.months}个月 ${age.days}天`
}

export function isDateTime(v: string): boolean {
  return /^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}$/.test(v.trim())
}

export function isDate(v: string): boolean {
  return /^\d{4}-\d{2}-\d{2}$/.test(v.trim())
}

export { pad }
