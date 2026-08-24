import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import { authApi } from '@/api/endpoints'
import type { Tokens, User } from '@/types/models'

export const useAuthStore = defineStore('auth', () => {
  const accessToken = ref(localStorage.getItem('gp.access') || '')
  const refreshToken = ref(localStorage.getItem('gp.refresh') || '')
  const user = ref<User | null>(null)

  const isAuthed = computed(() => !!accessToken.value)

  function persist(tokens: Tokens) {
    accessToken.value = tokens.access_token
    refreshToken.value = tokens.refresh_token
    localStorage.setItem('gp.access', tokens.access_token)
    localStorage.setItem('gp.refresh', tokens.refresh_token)
  }

  async function login(email: string, password: string) {
    const data = await authApi.login(email, password)
    persist(data.tokens)
    user.value = data.user
    return data.user
  }

  async function register(email: string, password: string, nickname: string) {
    const data = await authApi.register(email, password, nickname)
    persist(data.tokens)
    user.value = data.user
    return data.user
  }

  async function fetchMe() {
    user.value = await authApi.me()
    return user.value
  }

  function logout() {
    accessToken.value = ''
    refreshToken.value = ''
    user.value = null
    localStorage.removeItem('gp.access')
    localStorage.removeItem('gp.refresh')
  }

  return { accessToken, refreshToken, user, isAuthed, login, register, fetchMe, logout, persist }
})
