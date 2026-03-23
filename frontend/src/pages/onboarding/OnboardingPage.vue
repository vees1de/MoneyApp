<script setup lang="ts">
import { onMounted, reactive } from 'vue'
import { useRouter } from 'vue-router'

import { useAppUiStore } from '@/app/stores/app-ui'
import { useFinanceStore } from '@/app/stores/finance'
import { useUserStore } from '@/app/stores/user'
import OnboardingChecklist from '@/features/onboarding/OnboardingChecklist.vue'
import { useI18n } from '@/shared/i18n'
import { parseAmountToMinor } from '@/shared/lib/money'
import PageContainer from '@/shared/ui/PageContainer.vue'

const router = useRouter()
const userStore = useUserStore()
const financeStore = useFinanceStore()
const appUiStore = useAppUiStore()
const { t } = useI18n()

const form = reactive({
  fullName: userStore.profile.fullName || t('onboarding.defaultName'),
  currency: userStore.profile.currency || 'RUB',
  timezone: userStore.profile.timezone || 'Asia/Yakutsk',
  primaryAccountName: financeStore.accounts[0]?.name ?? t('onboarding.defaultAccountName'),
  openingBalance: '2450',
})

async function submit() {
  try {
    await userStore.completeOnboarding({
      fullName: form.fullName,
      currency: form.currency,
      timezone: form.timezone,
    })

    if (financeStore.accounts[0]) {
      await financeStore.renameAccount({
        id: financeStore.accounts[0].id,
        name: form.primaryAccountName,
      })
    } else {
      await financeStore.addAccount({
        name: form.primaryAccountName,
        type: 'bank',
        balanceMinor: parseAmountToMinor(form.openingBalance),
        currency: form.currency,
      })
    }

    appUiStore.pushToast(t('onboarding.success'), 'success')
    await router.push('/dashboard')
  } catch (error) {
    const message = error instanceof Error ? error.message : t('onboarding.failure')
    appUiStore.pushToast(message, 'warning')
  }
}

onMounted(async () => {
  await financeStore.fetchAccounts()
})
</script>

<template>
  <PageContainer>
    <section class="surface-card stack">
      <div class="page-header">
        <h1>{{ t('onboarding.title') }}</h1>
        <p class="muted">{{ t('onboarding.subtitle') }}</p>
      </div>

      <div class="grid grid--two">
        <form class="stack" @submit.prevent="submit">
          <div class="field">
            <label for="fullName">{{ t('common.fullName') }}</label>
            <input id="fullName" v-model="form.fullName" />
          </div>

          <div class="grid grid--two">
            <div class="field">
              <label for="currency">{{ t('common.currency') }}</label>
              <select id="currency" v-model="form.currency">
                <option value="RUB">RUB</option>
                <option value="USD">USD</option>
                <option value="EUR">EUR</option>
              </select>
            </div>

            <div class="field">
              <label for="timezone">{{ t('common.timezone') }}</label>
              <input id="timezone" v-model="form.timezone" />
            </div>
          </div>

          <div class="field">
            <label for="accountName">{{ t('common.primaryAccountLabel') }}</label>
            <input id="accountName" v-model="form.primaryAccountName" />
          </div>

          <div class="field" v-if="!financeStore.accounts.length">
            <label for="openingBalance">{{ t('common.openingBalance') }}</label>
            <input id="openingBalance" v-model="form.openingBalance" inputmode="decimal" />
          </div>

          <button class="button button--primary button--block" type="submit">{{ t('onboarding.openDashboard') }}</button>
        </form>

        <section class="section-card">
          <OnboardingChecklist />
        </section>
      </div>
    </section>
  </PageContainer>
</template>
