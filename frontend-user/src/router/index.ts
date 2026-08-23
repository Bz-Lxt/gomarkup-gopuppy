import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/login', name: 'login', component: () => import('@/views/AuthView.vue'), meta: { public: true } },
    { path: '/', name: 'home', component: () => import('@/views/HomeView.vue') },
    { path: '/health', name: 'health', component: () => import('@/views/HealthView.vue') },
    { path: '/finance', name: 'finance', component: () => import('@/views/FinanceView.vue') },
    { path: '/album', name: 'album', component: () => import('@/views/AlbumView.vue') },
    { path: '/notify', name: 'notify', component: () => import('@/views/NotifyView.vue') },
    { path: '/family', name: 'family', component: () => import('@/views/FamilyView.vue') },
    { path: '/privacy', name: 'privacy', component: () => import('@/views/LegalView.vue'), meta: { public: true } },
    { path: '/terms', name: 'terms', component: () => import('@/views/LegalView.vue'), meta: { public: true } },
  ],
  scrollBehavior() {
    return { top: 0 }
  },
})

router.beforeEach(async (to) => {
  const auth = useAuthStore()
  if (to.meta.public) {
    if (to.path === '/login' && auth.isAuthed) return { path: '/' }
    return true
  }
  if (!auth.accessToken) {
    return { path: '/login', query: { redirect: to.fullPath } }
  }
  if (!auth.user) {
    try {
      await auth.fetchMe()
    } catch {
      auth.logout()
      return { path: '/login' }
    }
  }
  return true
})

export default router
