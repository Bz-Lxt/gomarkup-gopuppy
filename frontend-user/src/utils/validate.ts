export type FieldErrors = Record<string, string>
export type RuleFn = (value: unknown, all: Record<string, unknown>) => string | null

export function required(label: string): RuleFn {
  return (v) => {
    if (v === undefined || v === null || String(v).trim() === '') return `${label}不能为空`
    return null
  }
}

export function emailRule(label = '邮箱'): RuleFn {
  return (v) => {
    const s = String(v ?? '').trim()
    if (!s) return `${label}不能为空`
    if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(s)) return `${label}格式不正确`
    return null
  }
}

export function passwordRule(): RuleFn {
  return (v) => {
    const s = String(v ?? '')
    if (s.length < 8 || s.length > 72) return '密码需 8–72 位'
    if (!/[A-Za-z]/.test(s) || !/\d/.test(s)) return '密码需同时包含字母与数字'
    return null
  }
}

export function dateTimeRule(label: string): RuleFn {
  return (v) => {
    const s = String(v ?? '').trim()
    if (!s) return `${label}不能为空`
    if (!/^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}$/.test(s)) return `${label}需为 yyyy-MM-dd HH:mm:ss`
    return null
  }
}

export function dateRule(label: string): RuleFn {
  return (v) => {
    const s = String(v ?? '').trim()
    if (!s) return `${label}不能为空`
    if (!/^\d{4}-\d{2}-\d{2}$/.test(s)) return `${label}需为 yyyy-MM-dd`
    return null
  }
}

export function numberRange(label: string, min: number, max: number): RuleFn {
  return (v) => {
    if (v === '' || v === undefined || v === null) return `${label}不能为空`
    const n = Number(v)
    if (Number.isNaN(n)) return `${label}必须是数字`
    if (n < min || n > max) return `${label}需在 ${min}–${max} 之间`
    return null
  }
}

export function maxLen(label: string, n: number): RuleFn {
  return (v) => {
    if (String(v ?? '').trim().length > n) return `${label}不能超过 ${n} 字`
    return null
  }
}

export function validate(
  values: Record<string, unknown>,
  rules: Record<string, RuleFn[]>,
): FieldErrors {
  const errors: FieldErrors = {}
  for (const [key, fns] of Object.entries(rules)) {
    for (const fn of fns) {
      const msg = fn(values[key], values)
      if (msg) {
        errors[key] = msg
        break
      }
    }
  }
  return errors
}

export function hasErrors(errors: FieldErrors): boolean {
  return Object.keys(errors).length > 0
}
