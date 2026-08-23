<script setup lang="ts">
defineProps<{ lines?: number }>()
</script>

<template>
  <div class="rounded-page bg-card p-5 shadow-warm ring-1 ring-line">
    <div class="flex gap-4">
      <div class="skeleton h-20 w-20 rounded-2xl" />
      <div class="flex-1 space-y-3 pt-1">
        <div class="skeleton h-6 w-1/3 rounded-full" />
        <div class="skeleton h-4 w-2/3 rounded-full" />
        <div class="skeleton h-4 w-1/2 rounded-full" />
      </div>
    </div>
    <div class="mt-6 grid grid-cols-3 gap-3">
      <div v-for="n in lines ?? 6" :key="n" class="skeleton h-12 rounded-2xl" />
    </div>
  </div>
</template>
