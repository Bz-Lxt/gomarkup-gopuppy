<script setup lang="ts">
import { onMounted, reactive, ref, watch } from 'vue'
import { familyApi } from '@/api/endpoints'
import { apiErrorMessage } from '@/api/http'
import AppShell from '@/components/layout/AppShell.vue'
import AppButton from '@/components/ui/AppButton.vue'
import AppField from '@/components/ui/AppField.vue'
import EmptyState from '@/components/ui/EmptyState.vue'
import SkeletonCard from '@/components/ui/SkeletonCard.vue'
import { useAuthStore } from '@/stores/auth'
import { useFamilyStore } from '@/stores/family'
import { useModalStore } from '@/stores/modal'
import { useToastStore } from '@/stores/toast'
import type { FamilyInvite, Role } from '@/types/models'
import { formatDateTime } from '@/utils/datetime'
import { ROLE_LABEL } from '@/utils/labels'
import { hasErrors, maxLen, required, validate, type FieldErrors } from '@/utils/validate'

const family = useFamilyStore()
const auth = useAuthStore()
const toast = useToastStore()
const modal = useModalStore()
const invite = ref<FamilyInvite | null>(null)
const joinCode = ref('')
const familyName = ref('')
const errors = ref<FieldErrors>({})
const joinErrors = ref<FieldErrors>({})
const loading = ref(false)

const form = reactive({ role: 'CAREGIVER' as Role })

async function load() {
  loading.value = true
  try {
    if (family.currentFamilyId) await family.loadFamilyContext()
  } catch (e) {
    toast.error(apiErrorMessage(e))
  } finally {
    loading.value = false
  }
}

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

async function makeInvite() {
  if (!family.currentFamilyId) return
  try {
    invite.value = await familyApi.invite(family.currentFamilyId, form.role)
    toast.success('邀请码已生成，24 小时有效')
  } catch (e) {
    toast.error(apiErrorMessage(e))
  }
}

async function copyCode() {
  if (!invite.value) return
  try {
    await navigator.clipboard.writeText(invite.value.code)
    toast.success('已复制邀请码')
  } catch {
    toast.info(invite.value.code)
  }
}

async function removeMember(userId: string, name: string) {
  const ok = await modal.confirm({
    title: `移出 ${name}`,
    message: '移出后对方将无法再看到这个家庭的宠物。确认继续？',
    confirmText: '移出',
    danger: true,
  })
  if (!ok || !family.currentFamilyId) return
  try {
    await familyApi.removeMember(family.currentFamilyId, userId)
    toast.success('已移出成员')
    await family.loadFamilyContext()
  } catch (e) {
    toast.error(apiErrorMessage(e))
  }
}

async function createFamily() {
  errors.value = validate({ name: familyName.value }, { name: [required('家庭名'), maxLen('家庭名', 40)] })
  if (hasErrors(errors.value)) {
    toast.error('请先修正表单中的红色提示')
    return
  }
  try {
    const f = await familyApi.create(familyName.value)
    family.families = [...family.families, f]
    family.setFamily(f.id)
    await family.loadFamilyContext()
    toast.success('家庭已创建')
    familyName.value = ''
  } catch (e) {
    toast.error(apiErrorMessage(e))
  }
}

async function join() {
  joinErrors.value = validate(
    { code: joinCode.value },
    {
      code: [
        (v) => {
          if (String(v ?? '').trim().length !== 6) return '邀请码需为 6 位'
          return null
        },
      ],
    },
  )
  if (hasErrors(joinErrors.value)) {
    toast.error('请先修正表单中的红色提示')
    return
  }
  try {
    const f = await familyApi.join(joinCode.value)
    toast.success(`已加入 ${f.name}`)
    await family.bootstrap()
    joinCode.value = ''
  } catch (e) {
    toast.error(apiErrorMessage(e))
  }
}

</script>

<template>
  <AppShell>
    <div class="mb-5">
      <p class="text-xs tracking-[0.25em] text-ink/40">HOUSEHOLD</p>
      <h1 class="font-display text-4xl">{{ family.currentFamily?.name || '家庭成员' }}</h1>
    </div>

    <SkeletonCard v-if="loading && !family.members.length" />

    <div class="grid w-full gap-6 xl:grid-cols-5">
      <section class="xl:col-span-3">
        <h2 class="mb-3 font-display text-2xl">成员</h2>
        <EmptyState v-if="!family.members.length" title="还没有成员" hint="创建家庭或用邀请码加入。" />
        <ul v-else class="space-y-3">
          <li
            v-for="m in family.members"
            :key="m.user_id"
            class="flex flex-wrap items-center justify-between gap-3 rounded-page bg-card px-4 py-4 shadow-warm ring-1 ring-line"
          >
            <div>
              <p class="font-display text-xl">{{ m.nickname || m.email || m.user_id.slice(0, 8) }}</p>
              <p class="text-xs text-ink/45">{{ m.email }} · 加入于 {{ formatDateTime(m.joined_at) }}</p>
            </div>
            <div class="flex items-center gap-3">
              <span class="rounded-full bg-paper px-2.5 py-0.5 text-xs">{{ ROLE_LABEL[m.role] }}</span>
              <button
                v-if="family.canManage && m.user_id !== auth.user?.id"
                type="button"
                class="text-xs text-rose"
                @click="removeMember(m.user_id, m.nickname || '成员')"
              >
                移出
              </button>
            </div>
          </li>
        </ul>
      </section>

      <section class="space-y-5 xl:col-span-2">
        <div class="rounded-page bg-card p-5 shadow-warm ring-1 ring-line">
          <h3 class="font-display text-2xl">邀请码</h3>
          <p class="mt-1 text-sm text-ink/55">主人可生成 6 位码，24 小时有效。</p>
          <AppField label="赋予角色" class="mt-4">
            <select v-model="form.role" class="w-full rounded-2xl bg-paper px-3 py-2 ring-1 ring-line">
              <option value="CAREGIVER">共同养育</option>
              <option value="VIEWER">访客</option>
            </select>
          </AppField>
          <AppButton class="mt-3" :disabled="!family.canManage" @click="makeInvite">生成邀请码</AppButton>
          <div v-if="invite" class="mt-4 rounded-2xl bg-paper p-4 text-center">
            <p class="font-display text-4xl tracking-[0.3em]">{{ invite.code }}</p>
            <p class="mt-1 text-xs text-ink/45">有效至 {{ formatDateTime(invite.expires_at) }}</p>
            <AppButton variant="ghost" class="mt-2" @click="copyCode">复制</AppButton>
          </div>
        </div>

        <form class="rounded-page bg-card p-5 ring-1 ring-line" @submit.prevent="join">
          <h3 class="font-display text-2xl">加入家庭</h3>
          <AppField label="邀请码" class="mt-3" required :error="joinErrors.code">
            <input v-model="joinCode" maxlength="6" class="w-full rounded-2xl bg-paper px-3 py-2 uppercase tracking-[0.25em] ring-1 ring-line" />
          </AppField>
          <AppButton type="submit" variant="moss" class="mt-3">加入</AppButton>
        </form>

        <form class="rounded-page bg-card p-5 ring-1 ring-line" @submit.prevent="createFamily">
          <h3 class="font-display text-2xl">新建家庭</h3>
          <AppField label="家庭名" class="mt-3" required :error="errors.name">
            <input v-model="familyName" class="w-full rounded-2xl bg-paper px-3 py-2 ring-1 ring-line" />
          </AppField>
          <AppButton type="submit" class="mt-3">创建</AppButton>
        </form>
      </section>
    </div>
  </AppShell>
</template>
