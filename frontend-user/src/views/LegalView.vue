<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { useToastStore } from '@/stores/toast'

const route = useRoute()
const toast = useToastStore()
const title = computed(() => (route.path === '/privacy' ? '隐私政策' : '服务条款'))

toast.info('Coming Soon')
</script>

<template>
  <div class="flex min-h-screen w-full items-center justify-center px-6">
    <div class="w-full max-w-lg rounded-page bg-card p-8 text-center shadow-warm ring-1 ring-line">
      <p class="text-xs tracking-[0.3em] text-ink/40">LEGAL</p>
      <h1 class="mt-2 font-display text-4xl">{{ title }}</h1>
      <p class="mt-4 text-sm leading-7 text-ink/60">内容即将到来。GoPuppy 只把数据留在你家的院子里。</p>
      <router-link to="/login" class="mt-6 inline-block text-sm text-clay underline">返回登录</router-link>
    </div>
  </div>
</template>
