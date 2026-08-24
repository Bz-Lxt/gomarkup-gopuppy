import { onUnmounted, watch } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { useFamilyStore } from '@/stores/family'
import type { WSMessage } from '@/types/models'

export function useFamilySocket() {
  const auth = useAuthStore()
  const family = useFamilyStore()
  let socket: WebSocket | null = null
  let delay = 1000
  let timer: number | null = null
  let stopped = false
  let generation = 0

  function clearTimer() {
    if (timer !== null) {
      window.clearTimeout(timer)
      timer = null
    }
  }

  function disconnect(intentional = false) {
    if (intentional) stopped = true
    clearTimer()
    const s = socket
    socket = null
    if (s && s.readyState <= WebSocket.OPEN) s.close()
  }

  function connect() {
    if (stopped || !auth.accessToken || !family.currentFamilyId) return
    const myGen = ++generation
    const proto = location.protocol === 'https:' ? 'wss' : 'ws'
    const url = `${proto}://${location.host}/ws?token=${encodeURIComponent(auth.accessToken)}&family_id=${family.currentFamilyId}`
    const ws = new WebSocket(url)
    socket = ws
    ws.onopen = () => {
      if (myGen !== generation) return
      delay = 1000
      void family.refreshTodayCheckins()
    }
    ws.onmessage = (ev) => {
      if (myGen !== generation) return
      try {
        const msg = JSON.parse(String(ev.data)) as WSMessage
        if (msg.type === 'checkin.updated' && msg.pet_id) {
          const items = msg.payload?.items
          if (Array.isArray(items)) family.setCheckins(msg.pet_id, items)
          else void family.loadCheckins(msg.pet_id)
        }
      } catch {
        /* ignore malformed */
      }
    }
    ws.onclose = () => {
      if (myGen !== generation || stopped) return
      timer = window.setTimeout(() => {
        connect()
      }, delay)
      delay = Math.min(delay * 2, 30000)
    }
  }

  function restart() {
    generation += 1
    clearTimer()
    const s = socket
    socket = null
    if (s && s.readyState <= WebSocket.OPEN) s.close()
    delay = 1000
    stopped = false
    connect()
  }

  watch(
    () => [auth.accessToken, family.currentFamilyId] as const,
    ([token, fid]) => {
      if (!token || !fid) {
        disconnect(true)
        stopped = false
        return
      }
      restart()
    },
    { immediate: true },
  )

  onUnmounted(() => disconnect(true))
}
