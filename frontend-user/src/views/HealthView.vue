<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { eventApi } from '@/api/endpoints'
import { apiErrorMessage } from '@/api/http'
import HealthTimeline from '@/components/health/HealthTimeline.vue'
import AppShell from '@/components/layout/AppShell.vue'
import AppButton from '@/components/ui/AppButton.vue'
import AppField from '@/components/ui/AppField.vue'
import EmptyState from '@/components/ui/EmptyState.vue'
import SkeletonCard from '@/components/ui/SkeletonCard.vue'
import { useFamilyStore } from '@/stores/family'
import { useToastStore } from '@/stores/toast'
import type { EventCategory, HealthEvent, Severity } from '@/types/models'
import { nowBeijing } from '@/utils/datetime'
import { EVENT_LABEL } from '@/utils/labels'
import { dateTimeRule, hasErrors, required, validate, type FieldErrors } from '@/utils/validate'

const family = useFamilyStore()
const toast = useToastStore()
const loading = ref(false)
const events = ref<HealthEvent[]>([])
const category = ref('')
const year = ref(0)
const showForm = ref(false)
const errors = ref<FieldErrors>({})

const cats: EventCategory[] = ['VACCINE', 'DEWORM', 'SURGERY', 'CHECKUP', 'SYMPTOM', 'MEDICATION', 'OTHER']
const years = computed(() => {
  const y = new Date().getFullYear()
  return [0, y, y - 1, y - 2]
})

const form = reactive({
  category: 'VACCINE' as EventCategory,
  title: '',
  occurred_at: nowBeijing(),
  description: '',
  clinic: '',
  severity: '' as Severity,
  treated: false,
  amount_yuan: '',
})

async function load() {
  const pet = family.currentPet
  if (!pet) {
    events.value = []
    return
  }
  loading.value = true
  try {
    events.value = await eventApi.list(pet.id, category.value || undefined, year.value || undefined)
  } catch (e) {
    toast.error(apiErrorMessage(e))
  } finally {
    loading.value = false
  }
}

watch(
  () => [family.currentPetId, family.pets.length, category.value, year.value],
  () => {
    if (family.bootstrapped) void load()
  },
)

onMounted(() => {
  if (family.bootstrapped) void load()
  else {
    const stop = watch(
      () => family.bootstrapped,
      (ok) => {
        if (ok) {
          void load()
          stop()
        }
      },
    )
  }
})

async function save() {
  const rules: Record<string, Parameters<typeof validate>[1][string]> = {
    category: [required('分类')],
    title: [required('标题')],
    occurred_at: [dateTimeRule('发生时间')],
  }
  if (form.category === 'SYMPTOM') {
    rules.severity = [
      (v) => {
        if (!String(v || '')) return '症状需标记严重度'
        return null
      },
    ]
  }
  errors.value = validate(form, rules)
  if (hasErrors(errors.value)) {
    toast.error('请先修正表单中的红色提示')
    return
  }
  const pet = family.currentPet
  if (!pet) return
  const amount = form.amount_yuan.trim() ? Math.round(Number(form.amount_yuan) * 100) : null
  if (form.amount_yuan.trim() && Number.isNaN(Number(form.amount_yuan))) {
    errors.value.amount_yuan = '费用需为数字（元）'
    toast.error('请先修正表单中的红色提示')
    return
  }
  try {
    await eventApi.create(pet.id, {
      category: form.category,
      title: form.title,
      occurred_at: form.occurred_at,
      description: form.description,
      clinic: form.clinic,
      severity: form.severity,
      treated: form.treated,
      amount_cents: amount,
    })
    toast.success('已写入时间轴')
    showForm.value = false
    await load()
  } catch (e) {
    toast.error(apiErrorMessage(e))
  }
}
</script>

<template>
  <AppShell>
    <div class="mb-5 flex flex-wrap items-end justify-between gap-3">
      <div>
        <p class="text-xs tracking-[0.25em] text-ink/40">HEALTH TIMELINE</p>
        <h1 class="font-display text-4xl">{{ family.currentPet?.name || '健康' }} 的时间轴</h1>
      </div>
      <div class="flex flex-wrap items-center gap-2">
        <select
          v-if="family.pets.length"
          class="rounded-2xl bg-card px-3 py-2 text-sm ring-1 ring-line"
          :value="family.currentPetId"
          @change="family.setPet(($event.target as HTMLSelectElement).value)"
        >
          <option v-for="p in family.pets" :key="p.id" :value="p.id">{{ p.name }}</option>
        </select>
        <AppButton :disabled="!family.canWrite || !family.currentPet" @click="showForm = true">补录事件</AppButton>
      </div>
    </div>

    <div class="mb-6 flex w-full flex-wrap gap-2">
      <button
        type="button"
        class="rounded-full px-3 py-1 text-sm ring-1"
        :class="!category ? 'bg-clay text-white ring-clay' : 'bg-card ring-line'"
        @click="category = ''"
      >
        全部
      </button>
      <button
        v-for="c in cats"
        :key="c"
        type="button"
        class="rounded-full px-3 py-1 text-sm ring-1"
        :class="category === c ? 'bg-clay text-white ring-clay' : 'bg-card ring-line'"
        @click="category = c"
      >
        {{ EVENT_LABEL[c] }}
      </button>
      <select v-model.number="year" class="ml-auto rounded-2xl bg-card px-3 py-1.5 text-sm ring-1 ring-line">
        <option :value="0">全部年份</option>
        <option v-for="y in years.slice(1)" :key="y" :value="y">{{ y }}</option>
      </select>
    </div>

    <SkeletonCard v-if="loading" />
    <EmptyState v-else-if="!family.currentPet" title="先选一只宠物" hint="主页登记后，健康事件会按时间倒序铺开。" />
    <EmptyState v-else-if="!events.length" title="这一年很安静" hint="还没有符合筛选的健康记录。补录疫苗或症状吧。" />
    <HealthTimeline v-else :events="events" />

    <Teleport to="body">
      <div
        v-if="showForm"
        class="fixed inset-0 z-[60] overflow-y-auto bg-[#2A2118]/35 px-4 py-10"
        @click.self="showForm = false"
      >
        <form class="mx-auto w-full max-w-lg space-y-3 rounded-page bg-card p-6 shadow-warm" @submit.prevent="save">
          <h3 class="font-display text-2xl">补录健康事件</h3>
          <AppField label="分类" required>
            <select v-model="form.category" class="w-full rounded-2xl bg-paper px-3 py-2 ring-1 ring-line">
              <option v-for="c in cats" :key="c" :value="c">{{ EVENT_LABEL[c] }}</option>
            </select>
          </AppField>
          <AppField label="标题" required :error="errors.title">
            <input v-model="form.title" class="w-full rounded-2xl bg-paper px-3 py-2 ring-1 ring-line" />
          </AppField>
          <AppField label="发生时间" required :error="errors.occurred_at">
            <input
              v-model="form.occurred_at"
              class="w-full rounded-2xl bg-paper px-3 py-2 ring-1 ring-line"
              placeholder="yyyy-MM-dd HH:mm:ss"
            />
          </AppField>
          <AppField label="描述">
            <textarea v-model="form.description" rows="3" class="w-full rounded-2xl bg-paper px-3 py-2 ring-1 ring-line" />
          </AppField>
          <AppField label="就诊机构">
            <input v-model="form.clinic" class="w-full rounded-2xl bg-paper px-3 py-2 ring-1 ring-line" />
          </AppField>
          <AppField v-if="form.category === 'SYMPTOM'" label="严重度" required :error="errors.severity">
            <select v-model="form.severity" class="w-full rounded-2xl bg-paper px-3 py-2 ring-1 ring-line">
              <option value="">请选择</option>
              <option value="MILD">轻微</option>
              <option value="MODERATE">中等</option>
              <option value="SEVERE">严重</option>
            </select>
          </AppField>
          <label class="flex items-center gap-2 text-sm">
            <input v-model="form.treated" type="checkbox" /> 已就医
          </label>
          <AppField label="费用（元，可选）" :error="errors.amount_yuan">
            <input v-model="form.amount_yuan" class="w-full rounded-2xl bg-paper px-3 py-2 ring-1 ring-line" />
          </AppField>
          <div class="flex justify-end gap-2">
            <AppButton variant="ghost" @click="showForm = false">取消</AppButton>
            <AppButton type="submit">保存</AppButton>
          </div>
        </form>
      </div>
    </Teleport>
  </AppShell>
</template>
