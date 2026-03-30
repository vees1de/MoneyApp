<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'

import { useAppUiStore } from '@/app/stores/app-ui'
import { useAuthStore } from '@/app/stores/auth'
import { useUserStore } from '@/app/stores/user'
import { translateProvider, useI18n } from '@/shared/i18n'
import { runCicdSmoke } from '@/shared/api/services/cicd'
import { formatDateTime } from '@/shared/lib/date'
import PageContainer from '@/shared/ui/PageContainer.vue'
import LanguageSwitch from '@/shared/ui/LanguageSwitch.vue'

const router = useRouter()
const appUiStore = useAppUiStore()
const authStore = useAuthStore()
const userStore = useUserStore()
const { t } = useI18n()
const isSmokeRunning = ref(false)
const lastSmokeCheck = ref<{
  createdAt: string
  totalRuns: number
} | null>(null)

async function signOut() {
  await authStore.logout()
  appUiStore.clearToasts()
  await router.push('/login')
}

async function triggerSmokeTest() {
  if (isSmokeRunning.value) {
    return
  }

  try {
    isSmokeRunning.value = true

    const response = await runCicdSmoke()
    lastSmokeCheck.value = {
      createdAt: response.check.created_at,
      totalRuns: response.total_runs,
    }

    appUiStore.pushToast(t('settings.cicdSuccess', { count: response.total_runs }), 'success')
  } catch (error) {
    const message = error instanceof Error ? error.message : t('settings.cicdFailed')
    appUiStore.pushToast(message, 'warning')
  } finally {
    isSmokeRunning.value = false
  }
}

onMounted(async () => {
  if (!userStore.profile.id) {
    await userStore.fetchProfile()
  }
})
</script>

<template>
  <PageContainer>
    <div class="page-header">
      <h1>{{ t('settings.title') }}</h1>
      <p class="muted">{{ t('settings.subtitle') }}</p>
    </div>

    <section class="settings-profile-card">
      <div class="settings-avatar">
        {{ userStore.profile.fullName?.charAt(0) ?? '?' }}
      </div>
      <div class="settings-profile-info">
        <span class="settings-profile-name">{{ userStore.profile.fullName || t('common.anonymous') }}</span>
        <span class="settings-profile-sub">{{ translateProvider(userStore.profile.provider) }}</span>
      </div>
    </section>

    <section class="section-card">
      <div class="settings-list">
        <div class="settings-row">
          <span class="settings-row__label">{{ t('common.currency') }}</span>
          <span class="settings-row__value">{{ userStore.profile.currency || '—' }}</span>
        </div>
        <div class="settings-row">
          <span class="settings-row__label">{{ t('common.timezone') }}</span>
          <span class="settings-row__value">{{ userStore.profile.timezone || '—' }}</span>
        </div>
        <div class="settings-row">
          <span class="settings-row__label">{{ t('common.provider') }}</span>
          <span class="settings-row__value">{{ userStore.profile.provider ? translateProvider(userStore.profile.provider) : t('common.notLinked') }}</span>
        </div>
      </div>
    </section>

    <section class="section-card">
      <div class="settings-language">
        <div>
          <h2 class="settings-section-title">{{ t('settings.languageTitle') }}</h2>
          <p class="muted settings-language__body">{{ t('settings.languageBody') }}</p>
        </div>
        <LanguageSwitch />
      </div>
    </section>

    <section class="section-card">
      <div class="stack" style="gap: 10px">
        <RouterLink class="button button--secondary button--block" to="/categories">
          {{ t('settings.manageCategories') }}
        </RouterLink>
        <button class="button button--secondary button--block" type="button" @click="userStore.fetchProfile()">
          {{ t('settings.refreshProfile') }}
        </button>
        <button
          class="button button--secondary button--block"
          type="button"
          :disabled="isSmokeRunning"
          @click="triggerSmokeTest"
        >
          {{ isSmokeRunning ? t('settings.cicdRunning') : t('settings.cicdTrigger') }}
        </button>
        <button class="button button--danger button--block" type="button" @click="signOut">
          {{ t('settings.signOut') }}
        </button>
        <p v-if="lastSmokeCheck" class="muted settings-smoke-meta">
          {{ t('settings.cicdLastRun', { date: formatDateTime(lastSmokeCheck.createdAt), count: lastSmokeCheck.totalRuns }) }}
        </p>
      </div>
    </section>

    <p class="settings-version">{{ t('settings.version') }}</p>
  </PageContainer>
</template>

<style scoped>
.settings-profile-card {
  display: flex;
  align-items: center;
  gap: 16px;
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: var(--radius-xl);
  box-shadow: var(--shadow-sm);
  padding: 22px;
}

.settings-avatar {
  width: 56px;
  height: 56px;
  border-radius: 50%;
  background: var(--brand-soft);
  color: var(--brand);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.375rem;
  font-weight: 700;
  flex-shrink: 0;
}

.settings-profile-info {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.settings-profile-name {
  font-size: 1.1875rem;
  font-weight: 700;
  letter-spacing: -0.025em;
  color: var(--text-primary);
}

.settings-profile-sub {
  font-size: 0.875rem;
  color: var(--text-muted);
  text-transform: capitalize;
}

.settings-language {
  display: flex;
  justify-content: space-between;
  gap: 16px;
  align-items: center;
}

.settings-language__body {
  margin: 6px 0 0;
  max-width: 420px;
  font-size: 0.875rem;
}

.settings-section-title {
  margin: 0;
  font-size: 1.0625rem;
  font-weight: 600;
  letter-spacing: -0.015em;
  color: var(--text-primary);
}

.settings-row__label {
  font-size: 0.9375rem;
  color: var(--text-secondary);
}

.settings-row__value {
  font-size: 0.9375rem;
  font-weight: 600;
  color: var(--text-primary);
}

.settings-version {
  text-align: center;
  font-size: 0.75rem;
  color: var(--text-muted);
  margin: 4px 0 0;
}

.settings-smoke-meta {
  margin: 2px 4px 0;
  font-size: 0.8125rem;
  line-height: 1.5;
}
</style>
