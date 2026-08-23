<script setup lang="ts">
import { computed } from 'vue'
import type { CheckinType, DailyCheckin, Pet, Slot } from '@/types/models'
import { formatAge, formatDateTime } from '@/utils/datetime'
import { CHECKIN_LABEL, GENDER_LABEL, SLOT_LABEL, SPECIES_LABEL } from '@/utils/labels'

const props = defineProps<{
  pet: Pet
  checkins: DailyCheckin[]
  canWrite: boolean
  canArchive?: boolean
}>()

const emit = defineEmits<{
  toggle: [type: CheckinType, slot: Slot, done: boolean]
  archive: []
}>()

const slots: Slot[] = ['MORNING', 'NOON', 'NIGHT']
const types: CheckinType[] = ['FEED', 'MEDICINE']

function find(type: CheckinType, slot: Slot) {
  return props.checkins.find((c) => c.type === type && c.slot === slot && !c.revoked_at)
}

const speciesMark = computed(() => {
  if (props.pet.species === 'DOG') return '汪'
  if (props.pet.species === 'CAT') return '喵'
  return '爪'
})

function onToggle(type: CheckinType, slot: Slot) {
  if (!props.canWrite) return
  const cur = find(type, slot)
  emit('toggle', type, slot, !cur)
}
</script>

<template>
  <article class="rounded-page bg-card p-5 shadow-warm ring-1 ring-line rise md:p-6">
    <div class="flex items-start gap-4">
      <div
        class="relative grid h-20 w-20 shrink-0 place-items-center rounded-[22px] bg-paper text-2xl font-display text-clay ring-1 ring-line"
      >
        {{ speciesMark }}
        <span class="absolute -bottom-1 -right-1 rounded-full bg-moss px-1.5 text-[10px] text-white">{{
          SPECIES_LABEL[pet.species]
        }}</span>
      </div>
      <div class="min-w-0 flex-1">
        <div class="flex items-start justify-between gap-3">
          <div>
            <h2 class="font-display text-3xl leading-none">{{ pet.name }}</h2>
            <p class="mt-1 text-sm text-ink/55">
              {{ pet.breed || '品种未填' }} · {{ GENDER_LABEL[pet.gender] }}
              <span v-if="pet.neutered"> · 已绝育</span>
            </p>
          </div>
          <button
            v-if="canArchive"
            type="button"
            class="text-xs text-ink/40 hover:text-rose"
            @click="emit('archive')"
          >
            归档
          </button>
        </div>
        <p class="mt-3 font-display text-xl text-clay">
          {{ formatAge(pet.age) }}
          <span class="ml-2 text-sm text-ink/40">共 {{ pet.age.total_days }} 天</span>
        </p>
      </div>
    </div>

    <p v-if="pet.note" class="mt-4 text-sm leading-7 text-ink/65">{{ pet.note }}</p>

    <div class="mt-5 space-y-3">
      <div v-for="type in types" :key="type">
        <p class="mb-2 text-xs uppercase tracking-[0.18em] text-ink/40">今日{{ CHECKIN_LABEL[type] }}</p>
        <div class="grid grid-cols-3 gap-2">
          <button
            v-for="slot in slots"
            :key="type + slot"
            type="button"
            :disabled="!canWrite"
            class="rounded-2xl px-2 py-3 text-left ring-1 transition disabled:cursor-not-allowed"
            :class="
              find(type, slot)
                ? type === 'FEED'
                  ? 'bg-clay text-white ring-clay shadow-stamp'
                  : 'bg-moss text-white ring-moss shadow-stamp'
                : 'bg-paper/60 ring-line hover:bg-paper'
            "
            @click="onToggle(type, slot)"
          >
            <p class="text-sm font-medium">{{ SLOT_LABEL[slot] }}</p>
            <p v-if="find(type, slot)" class="mt-1 text-[11px] leading-4 opacity-90">
              {{ find(type, slot)?.done_by_name }}
              <br />
              {{ formatDateTime(find(type, slot)?.done_at).slice(11, 16) }}
            </p>
            <p v-else class="mt-1 text-[11px] text-ink/35">未打卡</p>
          </button>
        </div>
      </div>
    </div>
  </article>
</template>
