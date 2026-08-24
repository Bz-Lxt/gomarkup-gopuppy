<script setup lang="ts">
defineProps<{
  label: string
  error?: string
  required?: boolean
}>()
</script>

<template>
  <label class="block">
    <span class="mb-1.5 block text-sm text-ink/80">
      {{ label }}<span v-if="required" class="ml-0.5 text-rose">*</span>
    </span>
    <slot />
    <p v-if="error" class="mt-1 text-sm text-rose">{{ error }}</p>
  </label>
</template>
