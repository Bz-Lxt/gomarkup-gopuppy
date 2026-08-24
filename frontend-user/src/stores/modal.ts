import { ref } from 'vue'
import { defineStore } from 'pinia'

interface ConfirmOpts {
  title: string
  message: string
  confirmText?: string
  danger?: boolean
}

export const useModalStore = defineStore('modal', () => {
  const open = ref(false)
  const title = ref('')
  const message = ref('')
  const confirmText = ref('确认')
  const danger = ref(false)
  let resolver: ((ok: boolean) => void) | null = null

  function confirm(opts: ConfirmOpts): Promise<boolean> {
    title.value = opts.title
    message.value = opts.message
    confirmText.value = opts.confirmText ?? '确认'
    danger.value = !!opts.danger
    open.value = true
    return new Promise((resolve) => {
      resolver = resolve
    })
  }

  function close(ok: boolean) {
    open.value = false
    resolver?.(ok)
    resolver = null
  }

  return { open, title, message, confirmText, danger, confirm, close }
})
