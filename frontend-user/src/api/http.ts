import axios, { type AxiosResponse } from 'axios'
import type { Envelope } from '@/types/models'

export const http = axios.create({
  baseURL: '/api/v1',
  timeout: 30000,
})

http.interceptors.request.use((config) => {
  const token = localStorage.getItem('gp.access')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

http.interceptors.response.use(
  (res) => res,
  async (err: unknown) => {
    if (axios.isAxiosError(err)) {
      const status = err.response?.status
      const url = String(err.config?.url || '')
      const isAuthCall = url.includes('/auth/login') || url.includes('/auth/register')
      if (status === 401 && !isAuthCall) {
        const { useAuthStore } = await import('@/stores/auth')
        const router = (await import('@/router')).default
        useAuthStore().logout()
        if (router.currentRoute.value.path !== '/login') {
          await router.push({ path: '/login', query: { redirect: router.currentRoute.value.fullPath } })
        }
      }
    }
    return Promise.reject(err)
  },
)

export async function unwrap<T>(p: Promise<AxiosResponse<Envelope<T>>>): Promise<T> {
  const { data } = await p
  return data.data
}

export function asList<T>(data: T[] | null | undefined): T[] {
  return Array.isArray(data) ? data : []
}

export function apiErrorMessage(err: unknown): string {
  if (axios.isAxiosError(err)) {
    const msg = (err.response?.data as Envelope<unknown> | undefined)?.message
    if (msg) return msg
    if (err.response?.status === 413) return '文件过大，单文件不超过 20MB'
    if (err.response?.status === 409) return '资源冲突，请检查是否已存在'
    if (!err.response) return '网络异常，请检查连接'
  }
  return '请求失败，请稍后再试'
}
