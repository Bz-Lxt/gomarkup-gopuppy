<script setup lang="ts">
import { onMounted, reactive, ref, watch } from 'vue'
import { financeApi } from '@/api/endpoints'
import { apiErrorMessage } from '@/api/http'
import FinanceCharts from '@/components/finance/FinanceCharts.vue'
import AppShell from '@/components/layout/AppShell.vue'
import AppButton from '@/components/ui/AppButton.vue'
import AppField from '@/components/ui/AppField.vue'
import EmptyState from '@/components/ui/EmptyState.vue'
import SkeletonCard from '@/components/ui/SkeletonCard.vue'
import { useFamilyStore } from '@/stores/family'
import { useToastStore } from '@/stores/toast'
import type { ExpenseCategory, FinanceSummary } from '@/types/models'
import { nowBeijing } from '@/utils/datetime'
import { EXPENSE_LABEL, yuan } from '@/utils/labels'
import { dateTimeRule, hasErrors, numberRange, required, validate, type FieldErrors } from '@/utils/validate'

const family = useFamilyStore()
const toast = useToastStore()
const loading = ref(false)
const summary = ref<FinanceSummary | null>(null)
const showW = ref(false)
const showE = ref(false)
const wErr = ref<FieldErrors>({})
const eErr = ref<FieldErrors>({})
const weight = reactive({ weight_kg: '', measured_at: nowBeijing(), note: '' })
const expense = reactive({
  category: 'FOOD' as ExpenseCategory,
  amount_yuan: '',
  spent_at: nowBeijing(),
  note: '',
})
const cats = Object.keys(EXPENSE_LABEL) as ExpenseCategory[]

async function load() {
  const pet = family.currentPet
  if (!pet) {
    summary.value = null
    return
  }
  loading.value = true
  try {
    summary.value = await financeApi.summary(pet.id)
  } catch (e) {
    toast.error(apiErrorMessage(e))
  } finally {
    loading.value = false
  }
}

function whenReady() {
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
}

watch(
  () => family.currentPetId,
  () => {
    if (family.bootstrapped) void load()
  },
)
onMounted(whenReady)

async function saveWeight() {
  wErr.value = validate(weight, {
    weight_kg: [numberRange('体重', 0.1, 200)],
    measured_at: [dateTimeRule('测量时间')],
  })
  if (hasErrors(wErr.value) || !family.currentPet) {
    if (hasErrors(wErr.value)) toast.error('请先修正表单中的红色提示')
    return
  }
  try {
    await financeApi.addWeight(family.currentPet.id, Number(weight.weight_kg), weight.measured_at, weight.note)
    toast.success('体重已记录')
    showW.value = false
    await load()
  } catch (e) {
    toast.error(apiErrorMessage(e))
  }
}

async function saveExpense() {
  eErr.value = validate(expense, {
    category: [required('分类')],
    amount_yuan: [numberRange('金额', 0.01, 1000000)],
    spent_at: [dateTimeRule('消费时间')],
  })
  if (hasErrors(eErr.value) || !family.currentPet) {
    if (hasErrors(eErr.value)) toast.error('请先修正表单中的红色提示')
    return
  }
  const cents = Math.round(Number(expense.amount_yuan) * 100)
  try {
    await financeApi.addExpense(family.currentPet.id, expense.category, cents, expense.spent_at, expense.note)
    toast.success('开销已入账')
    showE.value = false
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
        <p class="text-xs tracking-[0.25em] text-ink/40">LEDGER</p>
        <h1 class="font-display text-4xl">{{ family.currentPet?.name || '账本' }} 的账本与体重</h1>
      </div>
      <div class="flex flex-wrap gap-2">
        <select
          v-if="family.pets.length"
          class="rounded-2xl bg-card px-3 py-2 text-sm ring-1 ring-line"
          :value="family.currentPetId"
          @change="family.setPet(($event.target as HTMLSelectElement).value)"
        >
          <option v-for="p in family.pets" :key="p.id" :value="p.id">{{ p.name }}</option>
        </select>
        <AppButton variant="ghost" :disabled="!family.canWrite" @click="showW = true">记体重</AppButton>
        <AppButton :disabled="!family.canWrite" @click="showE = true">记一笔</AppButton>
      </div>
    </div>

    <SkeletonCard v-if="loading" />
    <EmptyState v-else-if="!summary" title="还没有账本" hint="选一只宠物后，这里会展开近一年的曲线。" />
    <template v-else>
      <div class="mb-4 grid gap-3 md:grid-cols-3">
        <div class="rounded-page bg-card p-5 shadow-warm ring-1 ring-line">
          <p class="text-xs text-ink/45">本月开销</p>
          <p class="mt-1 font-display text-3xl text-clay">{{ yuan(summary.month_total_cents) }}</p>
        </div>
        <div class="rounded-page bg-card p-5 shadow-warm ring-1 ring-line">
          <p class="text-xs text-ink/45">年度累计</p>
          <p class="mt-1 font-display text-3xl">{{ yuan(summary.year_total_cents) }}</p>
        </div>
        <div class="rounded-page bg-card p-5 shadow-warm ring-1 ring-line">
          <p class="text-xs text-ink/45">分类 Top3</p>
          <p class="mt-2 text-sm leading-7">
            <span v-for="t in summary.top3" :key="t.category" class="mr-3">
              {{ EXPENSE_LABEL[t.category as ExpenseCategory] || t.category }} {{ yuan(t.cents) }}
            </span>
            <span v-if="!summary.top3?.length" class="text-ink/40">暂无</span>
          </p>
        </div>
      </div>
      <EmptyState
        v-if="!summary.weight_series?.length && !summary.expense_series?.length"
        title="账本还是空白页"
        hint="记一笔开销或称一次体重，曲线就会长出来。"
      />
      <FinanceCharts v-else :summary="summary" />
    </template>

    <Teleport to="body">
      <div v-if="showW" class="fixed inset-0 z-[60] flex items-center justify-center bg-[#2A2118]/35 px-4" @click.self="showW = false">
        <form class="w-full max-w-md space-y-3 rounded-page bg-card p-6" @submit.prevent="saveWeight">
          <h3 class="font-display text-2xl">记录体重</h3>
          <AppField label="体重 kg" required :error="wErr.weight_kg">
            <input v-model="weight.weight_kg" class="w-full rounded-2xl bg-paper px-3 py-2 ring-1 ring-line" />
          </AppField>
          <AppField label="测量时间" required :error="wErr.measured_at">
            <input v-model="weight.measured_at" class="w-full rounded-2xl bg-paper px-3 py-2 ring-1 ring-line" placeholder="yyyy-MM-dd HH:mm:ss" />
          </AppField>
          <AppField label="备注">
            <input v-model="weight.note" class="w-full rounded-2xl bg-paper px-3 py-2 ring-1 ring-line" />
          </AppField>
          <div class="flex justify-end gap-2">
            <AppButton variant="ghost" @click="showW = false">取消</AppButton>
            <AppButton type="submit">保存</AppButton>
          </div>
        </form>
      </div>
      <div v-if="showE" class="fixed inset-0 z-[60] flex items-center justify-center bg-[#2A2118]/35 px-4" @click.self="showE = false">
        <form class="w-full max-w-md space-y-3 rounded-page bg-card p-6" @submit.prevent="saveExpense">
          <h3 class="font-display text-2xl">记一笔开销</h3>
          <AppField label="分类" required>
            <select v-model="expense.category" class="w-full rounded-2xl bg-paper px-3 py-2 ring-1 ring-line">
              <option v-for="c in cats" :key="c" :value="c">{{ EXPENSE_LABEL[c] }}</option>
            </select>
          </AppField>
          <AppField label="金额（元）" required :error="eErr.amount_yuan">
            <input v-model="expense.amount_yuan" class="w-full rounded-2xl bg-paper px-3 py-2 ring-1 ring-line" />
          </AppField>
          <AppField label="消费时间" required :error="eErr.spent_at">
            <input v-model="expense.spent_at" class="w-full rounded-2xl bg-paper px-3 py-2 ring-1 ring-line" placeholder="yyyy-MM-dd HH:mm:ss" />
          </AppField>
          <AppField label="备注">
            <input v-model="expense.note" class="w-full rounded-2xl bg-paper px-3 py-2 ring-1 ring-line" />
          </AppField>
          <div class="flex justify-end gap-2">
            <AppButton variant="ghost" @click="showE = false">取消</AppButton>
            <AppButton type="submit">保存</AppButton>
          </div>
        </form>
      </div>
    </Teleport>
  </AppShell>
</template>
