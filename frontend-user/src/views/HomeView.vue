<script setup lang="ts">
import { reactive, ref } from 'vue'
import { familyApi, petApi } from '@/api/endpoints'
import { apiErrorMessage } from '@/api/http'
import AppShell from '@/components/layout/AppShell.vue'
import PetHomeCard from '@/components/pets/PetHomeCard.vue'
import AppButton from '@/components/ui/AppButton.vue'
import AppField from '@/components/ui/AppField.vue'
import EmptyState from '@/components/ui/EmptyState.vue'
import SkeletonCard from '@/components/ui/SkeletonCard.vue'
import { useFamilyStore } from '@/stores/family'
import { useModalStore } from '@/stores/modal'
import { useToastStore } from '@/stores/toast'
import type { CheckinType, Gender, Slot, Species } from '@/types/models'
import { dateRule, hasErrors, maxLen, required, validate, type FieldErrors } from '@/utils/validate'

const family = useFamilyStore()
const toast = useToastStore()
const modal = useModalStore()

const showCreate = ref(false)
const showFamily = ref(false)
const errors = ref<FieldErrors>({})
const familyErrors = ref<FieldErrors>({})
const petForm = reactive({
  name: '',
  species: 'CAT' as Species,
  breed: '',
  gender: 'FEMALE' as Gender,
  birthday: '2023-03-15',
  neutered: true,
  chip_no: '',
  note: '',
})
const familyForm = reactive({ name: '林家小院', code: '' })

async function toggle(petId: string, type: CheckinType, slot: Slot, done: boolean) {
  try {
    await family.toggleCheckin(petId, type, slot, done)
  } catch (e) {
    toast.error(apiErrorMessage(e))
  }
}

async function archive(petId: string, name: string) {
  const ok = await modal.confirm({
    title: `归档 ${name}`,
    message: '归档后主页不再展示，历史档案仍会保留。确认继续？',
    confirmText: '归档',
    danger: true,
  })
  if (!ok) return
  try {
    await petApi.archive(petId)
    toast.success(`${name} 已归档`)
    await family.loadFamilyContext()
  } catch (e) {
    toast.error(apiErrorMessage(e))
  }
}

async function createPet() {
  errors.value = validate(petForm, {
    name: [required('名字'), maxLen('名字', 20)],
    species: [required('物种')],
    gender: [required('性别')],
    birthday: [dateRule('生日')],
  })
  if (hasErrors(errors.value)) {
    toast.error('请先修正表单中的红色提示')
    return
  }
  if (!family.currentFamilyId) {
    toast.error('请先创建或加入家庭')
    return
  }
  try {
    await petApi.create(family.currentFamilyId, { ...petForm })
    toast.success('新成员入册')
    showCreate.value = false
    await family.loadFamilyContext()
  } catch (e) {
    toast.error(apiErrorMessage(e))
  }
}

async function createFamily() {
  familyErrors.value = validate(familyForm, { name: [required('家庭名'), maxLen('家庭名', 40)] })
  if (hasErrors(familyErrors.value)) {
    toast.error('请先修正表单中的红色提示')
    return
  }
  try {
    const f = await familyApi.create(familyForm.name)
    family.families = [...family.families, f]
    family.setFamily(f.id)
    await family.loadFamilyContext()
    toast.success('家庭已创建')
    showFamily.value = false
  } catch (e) {
    toast.error(apiErrorMessage(e))
  }
}

async function joinFamily() {
  familyErrors.value = validate(familyForm, {
    code: [
      (v) => {
        const s = String(v ?? '').trim()
        if (s.length !== 6) return '邀请码需为 6 位'
        return null
      },
    ],
  })
  if (hasErrors(familyErrors.value)) {
    toast.error('请先修正表单中的红色提示')
    return
  }
  try {
    const f = await familyApi.join(familyForm.code)
    toast.success(`已加入 ${f.name}`)
    await family.bootstrap()
    showFamily.value = false
  } catch (e) {
    toast.error(apiErrorMessage(e))
  }
}
</script>

<template>
  <AppShell>
    <div class="mb-5 flex flex-wrap items-end justify-between gap-3">
      <div>
        <p class="text-xs tracking-[0.25em] text-ink/40">TODAY · 林家手帐</p>
        <h1 class="font-display text-4xl">{{ family.currentFamily?.name || '还没有家庭' }}</h1>
      </div>
      <div class="flex gap-2">
        <AppButton variant="ghost" @click="showFamily = true">家庭</AppButton>
        <AppButton :disabled="!family.canWrite" @click="showCreate = true">登记宠物</AppButton>
      </div>
    </div>

    <div v-if="family.loading && !family.bootstrapped" class="grid gap-4 md:grid-cols-2">
      <SkeletonCard />
      <SkeletonCard />
    </div>

    <EmptyState
      v-else-if="!family.families.length"
      title="院子还是空的"
      hint="创建一个家庭，或用 6 位邀请码加入林家小院。"
    >
      <AppButton @click="showFamily = true">创建 / 加入</AppButton>
    </EmptyState>

    <EmptyState
      v-else-if="!family.pets.length"
      title="还没有登记的小家伙"
      hint="把奶油、豆豆写进手帐的第一页吧。"
    >
      <AppButton :disabled="!family.canWrite" @click="showCreate = true">登记宠物</AppButton>
    </EmptyState>

    <div v-else class="grid w-full gap-4 md:grid-cols-2">
      <PetHomeCard
        v-for="p in family.pets"
        :key="p.id"
        :pet="p"
        :checkins="family.checkins[p.id] || []"
        :can-write="family.canWrite"
        :can-archive="family.canManage"
        @toggle="(t, s, d) => toggle(p.id, t, s, d)"
        @archive="archive(p.id, p.name)"
      />
    </div>

    <Teleport to="body">
      <div
        v-if="showCreate"
        class="fixed inset-0 z-[60] flex items-center justify-center bg-[#2A2118]/35 px-4"
        @click.self="showCreate = false"
      >
        <form class="w-full max-w-lg space-y-3 rounded-page bg-card p-6 shadow-warm" @submit.prevent="createPet">
          <h3 class="font-display text-2xl">登记新宠物</h3>
          <AppField label="名字" required :error="errors.name">
            <input v-model="petForm.name" class="w-full rounded-2xl bg-paper px-3 py-2 ring-1 ring-line" />
          </AppField>
          <div class="grid grid-cols-2 gap-3">
            <AppField label="物种" required>
              <select v-model="petForm.species" class="w-full rounded-2xl bg-paper px-3 py-2 ring-1 ring-line">
                <option value="CAT">猫</option>
                <option value="DOG">狗</option>
                <option value="OTHER">其他</option>
              </select>
            </AppField>
            <AppField label="性别" required>
              <select v-model="petForm.gender" class="w-full rounded-2xl bg-paper px-3 py-2 ring-1 ring-line">
                <option value="FEMALE">妹妹</option>
                <option value="MALE">弟弟</option>
                <option value="UNKNOWN">未知</option>
              </select>
            </AppField>
          </div>
          <AppField label="生日" required :error="errors.birthday">
            <input
              v-model="petForm.birthday"
              class="w-full rounded-2xl bg-paper px-3 py-2 ring-1 ring-line"
              placeholder="yyyy-MM-dd"
            />
          </AppField>
          <AppField label="品种">
            <input v-model="petForm.breed" class="w-full rounded-2xl bg-paper px-3 py-2 ring-1 ring-line" />
          </AppField>
          <label class="flex items-center gap-2 text-sm">
            <input v-model="petForm.neutered" type="checkbox" /> 已绝育
          </label>
          <AppField label="芯片号">
            <input v-model="petForm.chip_no" class="w-full rounded-2xl bg-paper px-3 py-2 ring-1 ring-line" />
          </AppField>
          <AppField label="备注">
            <textarea v-model="petForm.note" rows="2" class="w-full rounded-2xl bg-paper px-3 py-2 ring-1 ring-line" />
          </AppField>
          <div class="flex justify-end gap-2 pt-2">
            <AppButton variant="ghost" @click="showCreate = false">取消</AppButton>
            <AppButton type="submit">保存</AppButton>
          </div>
        </form>
      </div>
    </Teleport>

    <Teleport to="body">
      <div
        v-if="showFamily"
        class="fixed inset-0 z-[60] flex items-center justify-center bg-[#2A2118]/35 px-4"
        @click.self="showFamily = false"
      >
        <div class="w-full max-w-md space-y-5 rounded-page bg-card p-6 shadow-warm">
          <h3 class="font-display text-2xl">家庭</h3>
          <form class="space-y-3" @submit.prevent="createFamily">
            <AppField label="新家庭名" required :error="familyErrors.name">
              <input v-model="familyForm.name" class="w-full rounded-2xl bg-paper px-3 py-2 ring-1 ring-line" />
            </AppField>
            <AppButton type="submit">创建家庭</AppButton>
          </form>
          <div class="ink-rule" />
          <form class="space-y-3" @submit.prevent="joinFamily">
            <AppField label="邀请码" required :error="familyErrors.code">
              <input
                v-model="familyForm.code"
                maxlength="6"
                class="w-full rounded-2xl bg-paper px-3 py-2 uppercase tracking-[0.3em] ring-1 ring-line"
                placeholder="AB12CD"
              />
            </AppField>
            <AppButton type="submit" variant="moss">加入家庭</AppButton>
          </form>
        </div>
      </div>
    </Teleport>
  </AppShell>
</template>
