import { ref } from 'vue'
import { defineStore } from 'pinia'

export type ToastKind = 'error' | 'success' | 'info'

export interface ToastItem {
  id: number
  kind: ToastKind
  message: string
}

let seq = 1

export const useToastStore = defineStore('toast', () => {
  const items = ref<ToastItem[]>([])
  const timers = new Map<number, number>()

  function dismiss(id: number) {
    items.value = items.value.filter((t) => t.id !== id)
    const t = timers.get(id)
    if (t) {
      window.clearTimeout(t)
      timers.delete(id)
    }
  }

  function push(kind: ToastKind, message: string) {
    const id = seq++
    items.value.push({ id, kind, message })
    timers.set(
      id,
      window.setTimeout(() => dismiss(id), 5000),
    )
  }

  function error(message: string) {
    push('error', message)
  }
  function success(message: string) {
    push('success', message)
  }
  function info(message: string) {
    push('info', message)
  }

  return { items, dismiss, error, success, info }
})
