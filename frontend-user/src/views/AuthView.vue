<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import AppButton from '@/components/ui/AppButton.vue'
import AppField from '@/components/ui/AppField.vue'
import { apiErrorMessage } from '@/api/http'
import { useAuthStore } from '@/stores/auth'
import { useToastStore } from '@/stores/toast'
import { emailRule, hasErrors, maxLen, passwordRule, required, validate, type FieldErrors, type RuleFn } from '@/utils/validate'

const router = useRouter()
const route = useRoute()
const auth = useAuthStore()
const toast = useToastStore()

const mode = ref<'login' | 'register'>('login')
const submitting = ref(false)
const errors = ref<FieldErrors>({})
const form = reactive({
  email: 'dad@gopuppy.test',
  password: 'Puppy123!',
  nickname: '',
})

async function submit() {
  const rules: Record<string, RuleFn[]> =
    mode.value === 'login'
      ? { email: [emailRule()], password: [required('密码')] }
      : { email: [emailRule()], password: [passwordRule()], nickname: [required('昵称'), maxLen('昵称', 24)] }
  errors.value = validate(form, rules)
  if (hasErrors(errors.value)) {
    toast.error('请先修正表单中的红色提示')
    return
  }
  submitting.value = true
  try {
    if (mode.value === 'login') await auth.login(form.email, form.password)
    else await auth.register(form.email, form.password, form.nickname)
    toast.success(mode.value === 'login' ? '欢迎回家' : '注册成功，已为你登录')
    const redirect = typeof route.query.redirect === 'string' ? route.query.redirect : '/'
    await router.replace(redirect)
  } catch (e) {
    toast.error(apiErrorMessage(e))
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <div class="flex min-h-screen w-full items-center justify-center px-4 py-10">
    <div class="grid w-full max-w-4xl overflow-hidden rounded-[28px] bg-card shadow-warm ring-1 ring-line md:grid-cols-2">
      <aside class="relative hidden bg-[#E9D3B4] p-8 md:block">
        <p class="text-xs tracking-[0.3em] text-ink/50">WINDOW SILL · AFTERNOON</p>
        <h1 class="mt-3 font-display text-4xl leading-tight">窗台午后<br />的宠物手帐</h1>
        <p class="mt-4 text-sm leading-7 text-ink/70">
          陶土、苔藓与羊皮纸。给奶油和豆豆记下每一顿饭、每一次疫苗，以及谁在八点十二分喂过食。
        </p>
        <svg viewBox="0 0 280 220" class="absolute bottom-4 right-2 w-64" fill="none" aria-hidden="true">
          <rect x="30" y="20" width="200" height="140" rx="8" stroke="#2A2118" stroke-width="2" />
          <path d="M30 90h200" stroke="#2A2118" stroke-width="1.2" />
          <path d="M130 20v140" stroke="#2A2118" stroke-width="1.2" />
          <ellipse cx="88" cy="168" rx="28" ry="16" fill="#C45C26" opacity=".85" />
          <ellipse cx="78" cy="148" rx="10" ry="12" fill="#2A2118" />
          <ellipse cx="98" cy="148" rx="10" ry="12" fill="#2A2118" />
          <circle cx="210" cy="168" r="16" fill="#3D6B4F" />
          <path d="M210 152c10-18 24-10 18 4" stroke="#3D6B4F" stroke-width="3" />
        </svg>
      </aside>

      <section class="p-6 md:p-9">
        <p class="font-display text-3xl">{{ mode === 'login' ? '回家打卡' : '加入家庭' }}</p>
        <p class="mt-1 text-sm text-ink/55">测试账号 dad@gopuppy.test / Puppy123!</p>

        <div class="mt-5 grid grid-cols-2 rounded-2xl bg-paper p-1">
          <button
            type="button"
            class="rounded-xl py-2 text-sm"
            :class="mode === 'login' ? 'bg-card shadow-warm' : 'text-ink/50'"
            @click="mode = 'login'"
          >
            登录
          </button>
          <button
            type="button"
            class="rounded-xl py-2 text-sm"
            :class="mode === 'register' ? 'bg-card shadow-warm' : 'text-ink/50'"
            @click="mode = 'register'"
          >
            注册
          </button>
        </div>

        <form class="mt-6 space-y-4" @submit.prevent="submit">
          <AppField label="邮箱" required :error="errors.email">
            <input
              v-model="form.email"
              type="email"
              class="w-full rounded-2xl bg-paper px-3 py-2.5 ring-1 ring-line outline-none focus:ring-clay"
              placeholder="you@gopuppy.test"
            />
          </AppField>
          <AppField v-if="mode === 'register'" label="昵称" required :error="errors.nickname">
            <input
              v-model="form.nickname"
              class="w-full rounded-2xl bg-paper px-3 py-2.5 ring-1 ring-line outline-none focus:ring-clay"
              placeholder="林爸爸"
            />
          </AppField>
          <AppField label="密码" required :error="errors.password">
            <input
              v-model="form.password"
              type="password"
              class="w-full rounded-2xl bg-paper px-3 py-2.5 ring-1 ring-line outline-none focus:ring-clay"
              placeholder="至少 8 位，含字母与数字"
            />
          </AppField>
          <AppButton type="submit" :disabled="submitting" class="w-full">
            {{ submitting ? '正在开门…' : mode === 'login' ? '进入林家小院' : '创建帐号' }}
          </AppButton>
        </form>

        <p class="mt-6 text-center text-xs text-ink/40">
          <router-link to="/privacy" class="underline">隐私</router-link>
          ·
          <router-link to="/terms" class="underline">条款</router-link>
        </p>
      </section>
    </div>
  </div>
</template>
