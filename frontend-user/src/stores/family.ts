import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import { checkinApi, familyApi, petApi } from '@/api/endpoints'
import { useAuthStore } from '@/stores/auth'
import type { DailyCheckin, Family, FamilyMember, Pet, Role } from '@/types/models'

export const useFamilyStore = defineStore('family', () => {
  const families = ref<Family[]>([])
  const currentFamilyId = ref(localStorage.getItem('gp.family') || '')
  const members = ref<FamilyMember[]>([])
  const pets = ref<Pet[]>([])
  const currentPetId = ref(localStorage.getItem('gp.pet') || '')
  const checkins = ref<Record<string, DailyCheckin[]>>({})
  const loading = ref(false)
  const bootstrapped = ref(false)

  const currentFamily = computed(() => families.value.find((f) => f.id === currentFamilyId.value) || null)
  const currentPet = computed(() => pets.value.find((p) => p.id === currentPetId.value) || pets.value[0] || null)
  const myRole = computed<Role | ''>(() => {
    const uid = useAuthStore().user?.id
    if (!uid) return ''
    return members.value.find((m) => m.user_id === uid)?.role ?? ''
  })
  const canWrite = computed(() => myRole.value === 'OWNER' || myRole.value === 'CAREGIVER')
  const canManage = computed(() => myRole.value === 'OWNER')

  function setFamily(id: string) {
    currentFamilyId.value = id
    if (id) localStorage.setItem('gp.family', id)
    else localStorage.removeItem('gp.family')
  }

  function setPet(id: string) {
    currentPetId.value = id
    if (id) localStorage.setItem('gp.pet', id)
    else localStorage.removeItem('gp.pet')
  }

  function setCheckins(petId: string, items: DailyCheckin[]) {
    checkins.value = { ...checkins.value, [petId]: items.filter((c) => !c.revoked_at) }
  }

  async function loadCheckins(petId: string) {
    setCheckins(petId, await checkinApi.today(petId))
  }

  async function refreshTodayCheckins() {
    await Promise.all(pets.value.map((p) => loadCheckins(p.id)))
  }

  async function bootstrap() {
    loading.value = true
    try {
      families.value = await familyApi.list()
      if (!families.value.length) {
        setFamily('')
        pets.value = []
        members.value = []
        bootstrapped.value = true
        return
      }
      if (!currentFamilyId.value || !families.value.some((f) => f.id === currentFamilyId.value)) {
        setFamily(families.value[0].id)
      }
      await loadFamilyContext()
      bootstrapped.value = true
    } finally {
      loading.value = false
    }
  }

  async function loadFamilyContext() {
    if (!currentFamilyId.value) return
    const [ms, ps] = await Promise.all([
      familyApi.members(currentFamilyId.value),
      petApi.list(currentFamilyId.value),
    ])
    members.value = ms
    pets.value = ps
    if (ps.length && !ps.some((p) => p.id === currentPetId.value)) {
      setPet(ps[0].id)
    }
    if (!ps.length) setPet('')
    await refreshTodayCheckins()
  }

  async function toggleCheckin(petId: string, type: DailyCheckin['type'], slot: DailyCheckin['slot'], done: boolean) {
    const items = await checkinApi.toggle(petId, type, slot, done)
    setCheckins(petId, items)
  }

  function reset() {
    families.value = []
    members.value = []
    pets.value = []
    checkins.value = {}
    bootstrapped.value = false
    setFamily('')
    setPet('')
  }

  return {
    families,
    currentFamilyId,
    currentFamily,
    members,
    pets,
    currentPetId,
    currentPet,
    checkins,
    loading,
    bootstrapped,
    myRole,
    canWrite,
    canManage,
    setFamily,
    setPet,
    setCheckins,
    loadCheckins,
    refreshTodayCheckins,
    bootstrap,
    loadFamilyContext,
    toggleCheckin,
    reset,
  }
})
