<script setup lang="ts">
import { useRouter } from 'vue-router'

import { useAppUiStore } from '@/app/stores/app-ui'
import { useAuthStore } from '@/app/stores/auth'
import { useFinanceStore } from '@/app/stores/finance'
import { useReviewStore } from '@/app/stores/review'
import { useSavingsStore } from '@/app/stores/savings'
import { useUserStore } from '@/app/stores/user'
import PageContainer from '@/shared/ui/PageContainer.vue'
import SectionCard from '@/shared/ui/SectionCard.vue'

const router = useRouter()
const appUiStore = useAppUiStore()
const authStore = useAuthStore()
const userStore = useUserStore()
const financeStore = useFinanceStore()
const reviewStore = useReviewStore()
const savingsStore = useSavingsStore()

function resetWorkspace() {
  financeStore.reset()
  savingsStore.reset()
  reviewStore.reset()
  appUiStore.pushToast('Workspace reset to demo finance baseline.', 'success')
}

async function signOut() {
  authStore.logout()
  appUiStore.clearToasts()
  await router.push('/login')
}
</script>

<template>
  <PageContainer>
    <div class="page-header">
      <h1>Settings</h1>
      <p class="muted">Profile, session, and demo workspace controls live here until richer preferences arrive.</p>
    </div>

    <SectionCard>
      <div class="settings-list">
        <div class="settings-row">
          <span class="muted">Full name</span>
          <strong>{{ userStore.profile.fullName }}</strong>
        </div>
        <div class="settings-row">
          <span class="muted">Provider</span>
          <strong>{{ userStore.profile.provider ?? 'Not linked' }}</strong>
        </div>
        <div class="settings-row">
          <span class="muted">Currency</span>
          <strong>{{ userStore.profile.currency }}</strong>
        </div>
        <div class="settings-row">
          <span class="muted">Timezone</span>
          <strong>{{ userStore.profile.timezone }}</strong>
        </div>
      </div>
    </SectionCard>

    <SectionCard>
      <div class="cta-list">
        <button class="button button--secondary" type="button" @click="resetWorkspace">Reset demo data</button>
        <RouterLink class="button button--secondary" to="/categories">Manage categories</RouterLink>
        <button class="button button--danger" type="button" @click="signOut">Sign out</button>
      </div>
    </SectionCard>
  </PageContainer>
</template>
