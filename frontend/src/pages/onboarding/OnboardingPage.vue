<script setup lang="ts">
import { reactive } from 'vue'
import { useRouter } from 'vue-router'

import { useAppUiStore } from '@/app/stores/app-ui'
import { useFinanceStore } from '@/app/stores/finance'
import { useUserStore } from '@/app/stores/user'
import OnboardingChecklist from '@/features/onboarding/OnboardingChecklist.vue'
import { parseAmountToMinor } from '@/shared/lib/money'
import PageContainer from '@/shared/ui/PageContainer.vue'

const router = useRouter()
const userStore = useUserStore()
const financeStore = useFinanceStore()
const appUiStore = useAppUiStore()

const form = reactive({
  fullName: userStore.profile.fullName || 'Alex Petrov',
  currency: userStore.profile.currency || 'RUB',
  timezone: userStore.profile.timezone || 'Asia/Yakutsk',
  primaryAccountName: financeStore.accounts[0]?.name ?? 'Main card',
  openingBalance: '2450',
})

async function submit() {
  userStore.completeOnboarding({
    fullName: form.fullName,
    currency: form.currency,
    timezone: form.timezone,
  })

  if (financeStore.accounts[0]) {
    financeStore.updateAccount({
      id: financeStore.accounts[0].id,
      name: form.primaryAccountName,
    })
  } else {
    financeStore.addAccount({
      name: form.primaryAccountName,
      type: 'bank',
      balanceMinor: parseAmountToMinor(form.openingBalance),
      currency: form.currency,
    })
  }

  appUiStore.pushToast('Onboarding completed.', 'success')
  await router.push('/dashboard')
}
</script>

<template>
  <PageContainer>
    <section class="surface-card stack">
      <div class="page-header">
        <h1>Set your finance baseline</h1>
        <p class="muted">
          Keep onboarding compact: identity, defaults, and the main account name are enough for the MVP.
        </p>
      </div>

      <div class="grid grid--two">
        <form class="stack" @submit.prevent="submit">
          <div class="field">
            <label for="fullName">Full name</label>
            <input id="fullName" v-model="form.fullName" />
          </div>

          <div class="grid grid--two">
            <div class="field">
              <label for="currency">Currency</label>
              <select id="currency" v-model="form.currency">
                <option value="RUB">RUB</option>
                <option value="USD">USD</option>
                <option value="EUR">EUR</option>
              </select>
            </div>

            <div class="field">
              <label for="timezone">Timezone</label>
              <input id="timezone" v-model="form.timezone" />
            </div>
          </div>

          <div class="field">
            <label for="accountName">Primary account label</label>
            <input id="accountName" v-model="form.primaryAccountName" />
          </div>

          <div class="field" v-if="!financeStore.accounts.length">
            <label for="openingBalance">Opening balance</label>
            <input id="openingBalance" v-model="form.openingBalance" inputmode="decimal" />
          </div>

          <button class="button button--primary button--block" type="submit">Open dashboard</button>
        </form>

        <section class="section-card">
          <OnboardingChecklist />
        </section>
      </div>
    </section>
  </PageContainer>
</template>
