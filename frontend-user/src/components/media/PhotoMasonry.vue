<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { loadMediaUrl } from '@/composables/useMediaUrl'
import type { MediaFile } from '@/types/models'

const props = defineProps<{ items: MediaFile[] }>()
const emit = defineEmits<{ open: [index: number] }>()
const urls = ref<Record<string, string>>({})

onMounted(async () => {
  await Promise.all(
    props.items.map(async (m) => {
      try {
        urls.value[m.id] = await loadMediaUrl(m.id)
      } catch {
        urls.value[m.id] = ''
      }
    }),
  )
})
</script>

<template>
  <div class="columns-2 gap-3 md:columns-3 xl:columns-4">
    <button
      v-for="(m, i) in items"
      :key="m.id"
      type="button"
      class="mb-3 block w-full break-inside-avoid overflow-hidden rounded-2xl bg-card shadow-warm ring-1 ring-line"
      @click="emit('open', i)"
    >
      <img v-if="urls[m.id]" :src="urls[m.id]" :alt="m.filename" class="w-full object-cover" />
      <div v-else class="skeleton h-40 w-full" />
    </button>
  </div>
</template>
