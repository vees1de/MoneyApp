<script setup lang="ts">
import { onMounted } from 'vue'
import { useRouter } from 'vue-router'

import { useAppUiStore } from '@/app/stores/app-ui'
import { useAuthStore } from '@/app/stores/auth'
import { useUserStore } from '@/app/stores/user'
import PageContainer from '@/shared/ui/PageContainer.vue'

const router = useRouter()
const appUiStore = useAppUiStore()
const authStore = useAuthStore()
const userStore = useUserStore()

async function signOut() {
  await authStore.logout()
  appUiStore.clearToasts()
  await router.push('/login')
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
      <h1>Settings</h1>
      <p class="muted">Profile and session management.</p>
    </div>

    <!-- Profile card -->
    <section class="settings-profile-card">
      <div class="settings-avatar">
        {{ userStore.profile.fullName?.charAt(0) ?? '?' }}
      </div>
      <div class="settings-profile-info">
        <span class="settings-profile-name">{{ userStore.profile.fullName || 'Anonymous' }}</span>
        <span class="settings-profile-sub">{{ userStore.profile.provider ?? 'No provider' }}</span>
      </div>
    </section>

    <!-- Profile details -->
    <section class="section-card">
      <div class="settings-list">
        <div class="settings-row">
          <span class="settings-row__label">Currency</span>
          <span class="settings-row__value">{{ userStore.profile.currency || '—' }}</span>
        </div>
        <div class="settings-row">
          <span class="settings-row__label">Timezone</span>
          <span class="settings-row__value">{{ userStore.profile.timezone || '—' }}</span>
        </div>
        <div class="settings-row">
          <span class="settings-row__label">Provider</span>
          <span class="settings-row__value">{{ userStore.profile.provider || 'Not linked' }}</span>
        </div>
      </div>
    </section>

    <!-- Actions -->
    <section class="section-card">
      <div class="stack" style="gap:10px">
        <RouterLink class="button button--secondary button--block" to="/categories">
          Manage categories
        </RouterLink>
        <button class="button button--secondary button--block" type="button" @click="userStore.fetchProfile()">
          Refresh profile
        </button>
        <button class="button button--danger button--block" type="button" @click="signOut">
          Sign out
        </button>
      </div>
    </section>

    <!-- App info -->
    <p class="settings-version">Finance Core · Personal Life OS · MVP</p>
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
  padding: 20px;
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
  gap: 3px;
}

.settings-profile-name {
  font-size: 1.125rem;
  font-weight: 700;
  letter-spacing: -0.02em;
  color: var(--text-primary);
}

.settings-profile-sub {
  font-size: 0.875rem;
  color: var(--text-muted);
  text-transform: capitalize;
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
  font-size: 0.78125rem;
  color: var(--text-muted);
  margin: 4px 0 0;
}
</style>
