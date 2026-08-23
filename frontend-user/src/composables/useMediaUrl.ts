import { onUnmounted, ref, watch } from 'vue'
import { mediaApi } from '@/api/endpoints'

const cache = new Map<string, string>()

export async function loadMediaUrl(id: string): Promise<string> {
  const hit = cache.get(id)
  if (hit) return hit
  const res = await mediaApi.file(id)
  const url = URL.createObjectURL(res.data)
  cache.set(id, url)
  return url
}

export async function downloadMedia(id: string, filename: string) {
  const res = await mediaApi.file(id)
  const url = URL.createObjectURL(res.data)
  const a = document.createElement('a')
  a.href = url
  a.download = filename
  a.click()
  URL.revokeObjectURL(url)
}

export function useMediaUrl(id: () => string | undefined) {
  const url = ref('')
  const loading = ref(false)

  watch(
    id,
    async (next) => {
      url.value = ''
      if (!next) return
      loading.value = true
      try {
        url.value = await loadMediaUrl(next)
      } finally {
        loading.value = false
      }
    },
    { immediate: true },
  )

  onUnmounted(() => {
    /* object URLs are cached for reuse across album tiles */
  })

  return { url, loading }
}
