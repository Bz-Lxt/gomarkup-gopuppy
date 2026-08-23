<script setup lang="ts">
import { onMounted, reactive, ref, watch } from 'vue'
import { familyApi, reminderApi } from '@/api/endpoints'
import { apiErrorMessage } from '@/api/http'
import AppShell from '@/components/layout/AppShell.vue'
import AppButton from '@/components/ui/AppButton.vue'
import AppField from '@/components/ui/AppField.vue'
import EmptyState from '@/components/ui/EmptyState.vue'
import SkeletonCard from '@/components/ui/SkeletonCard.vue'
import { useFamilyStore } from '@/stores/family'
import { useToastStore } from '@/stores/toast'
import type { Channel, NotificationLog, ReminderKind, ReminderRule } from '@/types/models'
import { formatDateTime, nowBeijing } from '@/utils/datetime'
import { CHANNEL_LABEL, NOTIFY_KIND, NOTIFY_STATUS, REMINDER_LABEL } from '@/utils/labels'
import { dateTimeRule, hasErrors, numberRange, required, validate, type FieldErrors } from '@/utils/validate'

const family = useFamilyStore()
const toast = useToastStore()
const logs = ref<NotificationLog[]>([])
const rules = ref<ReminderRule[]>([])
const loading = ref(false)
const showForm = ref(false)
const errors = ref<FieldErrors>({})
const form = reactive({
  kind: 'DEWORM' as ReminderKind,
  title: '',
  cycle_days: '90',
  last_done_at: nowBeijing(),
  advance_days: '3',
  channels: ['EMAIL'] as Channel[],
})

const kinds: ReminderKind[] = ['VACCINE', 'DEWORM', 'MEDICINE', 'CHECKUP']
const channels: Channel[] = ['EMAIL', 'WECOM_BOT', 'WEBHOOK']

function statusClass(s: NotificationLog['status']) {
  if (s === 'SENT') return 'text-moss'
  if (s === 'PENDING') return 'text-[#8A6400]'
  return 'text-rose'
}

async function load() {
  if (!family.currentFamilyId) {
    logs.value = []
    rules.value = []
    return
  }
  loading.value = true
  try {
    logs.value = await familyApi.notifications(family.currentFamilyId)
    if (family.currentPet) rules.value = await reminderApi.list(family.currentPet.id)
    else rules.value = []
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
  () => [family.currentFamilyId, family.currentPetId],
  () => {
    if (family.bootstrapped) void load()
  },
)
onMounted(whenReady)

function toggleChannel(c: Channel) {
  if (form.channels.includes(c)) form.channels = form.channels.filter((x) => x !== c)
  else form.channels = [...form.channels, c]
}

async function save() {
  errors.value = validate(
    { ...form, channels: form.channels.join(',') },
    {
      kind: [required('类型')],
      title: [required('标题')],
      cycle_days: [numberRange('周期天数', 1, 3650)],
      last_done_at: [dateTimeRule('上次完成')],
      advance_days: [numberRange('提前天数', 0, 90)],
      channels: [
        () => (form.channels.length ? null : '至少选择一个通道'),
      ],
    },
  )
  if (hasErrors(errors.value) || !family.currentPet) {
    if (hasErrors(errors.value)) toast.error('请先修正表单中的红色提示')
    return
  }
  try {
    await reminderApi.create(family.currentPet.id, {
      kind: form.kind,
      title: form.title,
      cycle_days: Number(form.cycle_days),
      last_done_at: form.last_done_at,
      advance_days: Number(form.advance_days),
      channels: form.channels,
    })
    toast.success('提醒规则已建立')
    showForm.value = false
    await load()
  } catch (e) {
    toast.error(apiErrorMessage(e))
  }
}

async function replay(id: string) {
  try {
    await reminderApi.replay(id)
    toast.success('已重新投递')
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
        <p class="text-xs tracking-[0.25em] text-ink/40">REMINDERS</p>
        <h1 class="font-display text-4xl">通知中心</h1>
      </div>
      <div class="flex gap-2">
        <select
          v-if="family.pets.length"
          class="rounded-2xl bg-card px-3 py-2 text-sm ring-1 ring-line"
          :value="family.currentPetId"
          @change="family.setPet(($event.target as HTMLSelectElement).value)"
        >
          <option v-for="p in family.pets" :key="p.id" :value="p.id">{{ p.name }}</option>
        </select>
        <AppButton :disabled="!family.canWrite || !family.currentPet" @click="showForm = true">新建规则</AppButton>
      </div>
    </div>

    <SkeletonCard v-if="loading" />
    <div v-else class="grid w-full gap-6 xl:grid-cols-5">
      <section class="xl:col-span-3">
        <h2 class="mb-3 font-display text-2xl">投递记录</h2>
        <EmptyState v-if="!logs.length" title="还没有通知" hint="到期与提前提醒会写在这里，失败可重放。" />
        <ul v-else class="space-y-3">
          <li v-for="n in logs" :key="n.id" class="rounded-page bg-card p-4 shadow-warm ring-1 ring-line">
            <div class="flex flex-wrap items-center justify-between gap-2">
              <p class="font-medium">{{ n.title || '提醒' }} · {{ NOTIFY_KIND[n.kind] }}</p>
              <span class="text-xs" :class="statusClass(n.status)">{{ NOTIFY_STATUS[n.status] }}</span>
            </div>
            <p class="mt-1 text-xs text-ink/50">
              {{ CHANNEL_LABEL[n.channel] }} · 到期 {{ formatDateTime(n.due_date) }} · 计划 {{ formatDateTime(n.scheduled_at) }}
            </p>
            <p v-if="n.error" class="mt-1 text-xs text-rose">{{ n.error }}</p>
            <AppButton
              v-if="n.status === 'FAILED' || n.status === 'PERMANENT_FAILURE'"
              variant="ghost"
              class="mt-2"
              @click="replay(n.id)"
            >
              重放
            </AppButton>
          </li>
        </ul>
      </section>
      <section class="xl:col-span-2">
        <h2 class="mb-3 font-display text-2xl">{{ family.currentPet?.name || '宠物' }} 的规则</h2>
        <EmptyState v-if="!rules.length" title="尚未设定周期" hint="疫苗、驱虫、用药、体检都可以按天数循环。" />
        <ul v-else class="space-y-3">
          <li v-for="r in rules" :key="r.id" class="rounded-page bg-card p-4 ring-1 ring-line">
            <p class="text-xs text-clay">{{ REMINDER_LABEL[r.kind] }}</p>
            <p class="font-display text-xl">{{ r.title }}</p>
            <p class="mt-1 text-xs text-ink/50">
              每 {{ r.cycle_days }} 天 · 下次 {{ formatDateTime(r.next_due_at) }} · 提前 {{ r.advance_days }} 天
            </p>
            <p class="mt-1 text-xs text-ink/40">{{ r.channels.map((c) => CHANNEL_LABEL[c]).join(' / ') }}</p>
          </li>
        </ul>
      </section>
    </div>

    <Teleport to="body">
      <div
        v-if="showForm"
        class="fixed inset-0 z-[60] overflow-y-auto bg-[#2A2118]/35 px-4 py-10"
        @click.self="showForm = false"
      >
        <form class="mx-auto w-full max-w-lg space-y-3 rounded-page bg-card p-6" @submit.prevent="save">
          <h3 class="font-display text-2xl">提醒规则</h3>
          <AppField label="类型" required>
            <select v-model="form.kind" class="w-full rounded-2xl bg-paper px-3 py-2 ring-1 ring-line">
              <option v-for="k in kinds" :key="k" :value="k">{{ REMINDER_LABEL[k] }}</option>
            </select>
          </AppField>
          <AppField label="标题" required :error="errors.title">
            <input v-model="form.title" class="w-full rounded-2xl bg-paper px-3 py-2 ring-1 ring-line" />
          </AppField>
          <div class="grid grid-cols-2 gap-3">
            <AppField label="周期天数" required :error="errors.cycle_days">
              <input v-model="form.cycle_days" class="w-full rounded-2xl bg-paper px-3 py-2 ring-1 ring-line" />
            </AppField>
            <AppField label="提前天数" required :error="errors.advance_days">
              <input v-model="form.advance_days" class="w-full rounded-2xl bg-paper px-3 py-2 ring-1 ring-line" />
            </AppField>
          </div>
          <AppField label="上次完成" required :error="errors.last_done_at">
            <input v-model="form.last_done_at" class="w-full rounded-2xl bg-paper px-3 py-2 ring-1 ring-line" placeholder="yyyy-MM-dd HH:mm:ss" />
          </AppField>
          <AppField label="通道" required :error="errors.channels">
            <div class="flex flex-wrap gap-3 text-sm">
              <label v-for="c in channels" :key="c" class="flex items-center gap-1.5">
                <input type="checkbox" :checked="form.channels.includes(c)" @change="toggleChannel(c)" />
                {{ CHANNEL_LABEL[c] }}
              </label>
            </div>
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
