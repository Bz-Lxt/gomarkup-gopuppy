<script setup lang="ts">
import type { EventCategory, HealthEvent } from '@/types/models'
import { formatDateTime } from '@/utils/datetime'
import { EVENT_DOT, EVENT_LABEL, EVENT_TONE, SEVERITY_LABEL, yuan } from '@/utils/labels'

defineProps<{
  events: HealthEvent[]
}>()

function severity(ev: HealthEvent) {
  if (ev.category !== 'SYMPTOM' || !ev.severity) return ''
  return SEVERITY_LABEL[ev.severity as keyof typeof SEVERITY_LABEL] || ev.severity
}
</script>

<template>
  <ol class="relative ml-3 space-y-0 border-l-2 border-line pl-8">
    <li v-for="ev in events" :key="ev.id" class="relative pb-10">
      <span
        class="absolute -left-[41px] top-1.5 grid h-5 w-5 place-items-center rounded-full ring-4 ring-paper"
        :style="{ background: EVENT_DOT[ev.category as EventCategory] }"
      />
      <article class="rounded-page bg-card p-5 shadow-warm ring-1 ring-line">
        <div class="flex flex-wrap items-center gap-2">
          <span class="rounded-full px-2.5 py-0.5 text-xs ring-1" :class="EVENT_TONE[ev.category]">
            {{ EVENT_LABEL[ev.category] }}
          </span>
          <span v-if="severity(ev)" class="text-xs text-rose">{{ severity(ev) }}</span>
          <span v-if="ev.treated" class="text-xs text-moss">已就医</span>
          <time class="ml-auto text-xs text-ink/45">{{ formatDateTime(ev.occurred_at) }}</time>
        </div>
        <h3 class="mt-2 font-display text-2xl">{{ ev.title }}</h3>
        <p v-if="ev.description" class="mt-2 text-sm leading-7 text-ink/70">{{ ev.description }}</p>
        <div class="mt-3 flex flex-wrap gap-x-4 gap-y-1 text-xs text-ink/50">
          <span v-if="ev.clinic">就诊：{{ ev.clinic }}</span>
          <span v-if="ev.amount_cents">费用：{{ yuan(ev.amount_cents) }}</span>
        </div>
      </article>
    </li>
  </ol>
</template>
