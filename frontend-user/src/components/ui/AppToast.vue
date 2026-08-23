<script setup lang="ts">
import { useToastStore } from '@/stores/toast'

const toast = useToastStore()

const tone: Record<string, string> = {
  error: 'bg-[#FDF2F0] text-rose ring-rose/20',
  success: 'bg-[#F1F6F2] text-moss ring-moss/20',
  info: 'bg-[#FFF8EE] text-ink ring-line',
}
</script>

<template>
  <div class="pointer-events-none fixed right-4 top-4 z-[80] flex w-[min(92vw,380px)] flex-col gap-2">
    <div
      v-for="item in toast.items"
      :key="item.id"
      class="pointer-events-auto flex items-start gap-3 rounded-2xl px-4 py-3 shadow-warm ring-1 rise"
      :class="tone[item.kind]"
    >
      <p class="flex-1 text-sm leading-6">{{ item.message }}</p>
      <button
        type="button"
        class="mt-0.5 grid h-6 w-6 place-items-center rounded-full text-current/70 hover:bg-black/5"
        aria-label="关闭"
        @click="toast.dismiss(item.id)"
      >
        ×
      </button>
    </div>
  </div>
</template>
