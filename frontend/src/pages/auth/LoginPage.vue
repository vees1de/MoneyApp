<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'

import { useAppUiStore } from '@/app/stores/app-ui'
import { useAuthStore } from '@/app/stores/auth'
import { useUserStore } from '@/app/stores/user'
import AuthProviderButtons from '@/features/auth/AuthProviderButtons.vue'
import { env } from '@/shared/config/env'
import PageContainer from '@/shared/ui/PageContainer.vue'

const authStore = useAuthStore()
const userStore = useUserStore()
const appUiStore = useAppUiStore()
const router = useRouter()

const loading = ref(false)

async function login(provider: 'telegram' | 'yandex') {
  loading.value = true
  authStore.login(provider)
  appUiStore.pushToast(`Signed in with ${provider}.`, 'success')

  await router.push(userStore.profile.onboardingCompleted ? '/dashboard' : '/onboarding')
  loading.value = false
}
</script>

<template>
  <PageContainer>
    <section class="surface-card auth-hero stack">
      <span class="auth-hero__eyebrow">{{ env.appName }} / {{ env.productName }}</span>

      <div class="page-header">
        <h1>Finance Core for the future Personal Life OS</h1>
        <p class="muted">
          Start with cash flow clarity, weekly reconciliation, and savings momentum in a mobile-first
          PWA shell.
        </p>
      </div>

      <div class="grid grid--two">
        <div class="insight-card">
          <strong>Track what changed</strong>
          <p class="tiny">Income, expenses, categories, and account balances stay visible from the first screen.</p>
        </div>
        <div class="insight-card">
          <strong>Close the week cleanly</strong>
          <p class="tiny">Weekly review turns balance mismatch into guided action instead of bookkeeping friction.</p>
        </div>
      </div>

      <AuthProviderButtons :loading="loading" @select="login" />

      <p class="tiny">
        Foundation mode currently uses local demo state. Provider buttons already follow the real auth entry points.
      </p>
    </section>
  </PageContainer>
</template>
