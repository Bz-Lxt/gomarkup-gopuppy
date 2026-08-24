<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { loadMediaUrl } from '@/composables/useMediaUrl'
import type { MediaFile } from '@/types/models'
import { formatDateTime } from '@/utils/datetime'

const props = defineProps<{
  items: MediaFile[]
  index: number
}>()
const emit = defineEmits<{ close: []; index: [number] }>()
const url = ref('')

const current = computed(() => props.items[props.index])

async function load() {
  url.value = ''
  if (!current.value) return
  url.value = await loadMediaUrl(current.value.id)
}

watch(() => props.index, load)
onMounted(load)

function prev() {
  if (props.index > 0) emit('index', props.index - 1)
}
function next() {
  if (props.index < props.items.length - 1) emit('index', props.index + 1)
}
</script>

<template>
  <Teleport to="body">
    <div class="fixed inset-0 z-[75] flex flex-col bg-[#2A2118]/88" @click.self="emit('close')">
      <div class="flex items-center justify-between px-4 py-3 text-card">
        <p class="text-sm">{{ current?.filename }} · {{ formatDateTime(current?.created_at) }}</p>
        <button type="button" class="rounded-full px-3 py-1 ring-1 ring-white/30" @click="emit('close')">关闭</button>
      </div>
      <div class="flex flex-1 items-center justify-center gap-4 px-4">
        <button type="button" class="hidden rounded-full bg-white/10 px-3 py-2 text-card md:block" @click="prev">‹</button>
        <img v-if="url" :src="url" class="max-h-[80vh] max-w-full rounded-2xl object-contain" alt="" />
        <button type="button" class="hidden rounded-full bg-white/10 px-3 py-2 text-card md:block" @click="next">›</button>
      </div>
    </div>
  </Teleport>
</template>
