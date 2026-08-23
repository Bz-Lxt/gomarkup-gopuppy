<script setup lang="ts">
import { useModalStore } from '@/stores/modal'

const modal = useModalStore()
</script>

<template>
  <Teleport to="body">
    <div
      v-if="modal.open"
      class="fixed inset-0 z-[70] flex items-center justify-center bg-[#2A2118]/35 px-4 backdrop-blur-[2px]"
      @click.self="modal.close(false)"
    >
      <div class="w-full max-w-md rounded-page bg-card p-6 shadow-warm ring-1 ring-line rise">
        <h3 class="font-display text-2xl">{{ modal.title }}</h3>
        <p class="mt-3 text-sm leading-7 text-ink/75">{{ modal.message }}</p>
        <div class="mt-6 flex justify-end gap-3">
          <button
            type="button"
            class="rounded-2xl px-4 py-2 text-sm ring-1 ring-line hover:bg-paper"
            @click="modal.close(false)"
          >
            取消
          </button>
          <button
            type="button"
            class="stamp-press rounded-2xl px-4 py-2 text-sm text-white shadow-stamp"
            :class="modal.danger ? 'bg-rose' : 'bg-clay'"
            @click="modal.close(true)"
          >
            {{ modal.confirmText }}
          </button>
        </div>
      </div>
    </div>
  </Teleport>
</template>
