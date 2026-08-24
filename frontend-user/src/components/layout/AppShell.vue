<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useFamilySocket } from '@/composables/useFamilySocket'
import { useAuthStore } from '@/stores/auth'
import { useFamilyStore } from '@/stores/family'
import { useToastStore } from '@/stores/toast'
import { apiErrorMessage } from '@/api/http'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()
const family = useFamilyStore()
const toast = useToastStore()

useFamilySocket()

const links = [
  { to: '/', label: '主页', icon: '⌂' },
  { to: '/health', label: '健康', icon: '✚' },
  { to: '/finance', label: '账本', icon: '¥' },
  { to: '/album', label: '相册', icon: '▣' },
  { to: '/notify', label: '提醒', icon: '◎' },
  { to: '/family', label: '家庭', icon: '⌘' },
]

const active = computed(() => route.path)

onMounted(async () => {
  if (!family.bootstrapped) {
    try {
      await family.bootstrap()
    } catch (e) {
      toast.error(apiErrorMessage(e))
    }
  }
})

async function logout() {
  family.reset()
  auth.logout()
  await router.push('/login')
}

async function onFamilyChange(id: string) {
  family.setFamily(id)
  try {
    await family.loadFamilyContext()
  } catch (e) {
    toast.error(apiErrorMessage(e))
  }
}
</script>

<template>
  <div class="min-h-screen w-full">
    <header class="sticky top-0 z-40 border-b border-line/80 bg-[#F3E6D4]/85 backdrop-blur-md">
      <div class="flex w-full items-center gap-4 px-4 py-3 md:px-6">
        <router-link to="/" class="flex items-center gap-2">
          <span class="grid h-9 w-9 place-items-center rounded-2xl bg-clay text-sm text-white shadow-stamp">爪</span>
          <div class="leading-tight">
            <p class="font-display text-lg tracking-wide">GoPuppy</p>
            <p class="text-[11px] text-ink/50">暖陶管家</p>
          </div>
        </router-link>

        <nav class="hidden flex-1 items-center justify-center gap-1 xl:flex">
          <router-link
            v-for="l in links"
            :key="l.to"
            :to="l.to"
            class="rounded-full px-3.5 py-1.5 text-sm transition"
            :class="active === l.to ? 'bg-clay text-white' : 'text-ink/70 hover:bg-card'"
          >
            {{ l.label }}
          </router-link>
        </nav>

        <div class="ml-auto flex items-center gap-2">
          <select
            v-if="family.families.length > 1"
            class="hidden rounded-2xl bg-card px-3 py-2 text-sm ring-1 ring-line md:block"
            :value="family.currentFamilyId"
            @change="onFamilyChange(($event.target as HTMLSelectElement).value)"
          >
            <option v-for="f in family.families" :key="f.id" :value="f.id">{{ f.name }}</option>
          </select>
          <p class="hidden text-sm text-ink/70 md:block">{{ auth.user?.nickname }}</p>
          <button type="button" class="rounded-2xl px-3 py-1.5 text-sm ring-1 ring-line hover:bg-card" @click="logout">
            退出
          </button>
        </div>
      </div>
    </header>

    <main class="w-full px-4 pb-24 pt-5 md:px-6 md:pb-10">
      <slot />
    </main>

    <nav
      class="fixed inset-x-0 bottom-0 z-40 grid grid-cols-6 border-t border-line bg-[#F3E6D4]/95 px-1 py-1 backdrop-blur md:hidden"
    >
      <router-link
        v-for="l in links"
        :key="l.to"
        :to="l.to"
        class="flex flex-col items-center gap-0.5 rounded-xl py-1.5 text-[11px]"
        :class="active === l.to ? 'text-clay' : 'text-ink/55'"
      >
        <span class="text-base leading-none">{{ l.icon }}</span>
        {{ l.label }}
      </router-link>
    </nav>
  </div>
</template>
