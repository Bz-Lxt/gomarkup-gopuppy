<script setup lang="ts">
defineProps<{
  title: string
  hint?: string
}>()
</script>

<template>
  <div class="flex flex-col items-center justify-center px-6 py-16 text-center">
    <svg viewBox="0 0 160 96" class="h-24 w-40 text-clay/70" fill="none" aria-hidden="true">
      <path d="M20 78c18-22 38-22 56 0" stroke="currentColor" stroke-width="1.4" />
      <ellipse cx="48" cy="42" rx="16" ry="14" stroke="#3D6B4F" stroke-width="1.5" fill="#FFF8EE" />
      <ellipse cx="38" cy="28" rx="6" ry="8" fill="#C45C26" opacity=".35" />
      <ellipse cx="58" cy="26" rx="6" ry="8" fill="#C45C26" opacity=".35" />
      <path d="M88 70c8-18 28-18 36 0" stroke="#3D6B4F" stroke-width="1.4" />
      <path d="M118 38c8 4 12 14 8 22" stroke="#C45C26" stroke-width="1.4" />
      <circle cx="126" cy="34" r="5" fill="#E0A100" opacity=".7" />
    </svg>
    <h3 class="mt-4 font-display text-2xl">{{ title }}</h3>
    <p v-if="hint" class="mt-2 max-w-md text-sm leading-7 text-ink/60">{{ hint }}</p>
    <div class="mt-5">
      <slot />
    </div>
  </div>
</template>
