<script setup lang="ts">
import { reactive } from 'vue'

import { useAppUiStore } from '@/app/stores/app-ui'
import { useFinanceStore } from '@/app/stores/finance'
import { parseAmountToMinor, formatMoney } from '@/shared/lib/money'
import EmptyState from '@/shared/ui/EmptyState.vue'
import PageContainer from '@/shared/ui/PageContainer.vue'
import SectionCard from '@/shared/ui/SectionCard.vue'

const financeStore = useFinanceStore()
const appUiStore = useAppUiStore()

const form = reactive({
  balance: '',
  name: '',
  type: 'bank' as 'bank' | 'cash' | 'savings',
})

function submit() {
  if (!form.name.trim()) {
    appUiStore.pushToast('Account name is required.', 'warning')
    return
  }

  financeStore.addAccount({
    name: form.name,
    type: form.type,
    balanceMinor: parseAmountToMinor(form.balance),
    currency: 'RUB',
  })

  appUiStore.pushToast('Account added.', 'success')

  form.balance = ''
  form.name = ''
  form.type = 'bank'
}
</script>

<template>
  <PageContainer>
    <div class="page-header">
      <h1>Accounts</h1>
      <p class="muted">Balances live at account level. Every transaction should have a clear source of truth.</p>
    </div>

    <div class="grid grid--two">
      <form class="surface-card stack" @submit.prevent="submit">
        <div class="page-header">
          <h2>Add account</h2>
          <p class="muted">Keep account creation lightweight in the MVP.</p>
        </div>

        <div class="field">
          <label for="accountName">Name</label>
          <input id="accountName" v-model="form.name" placeholder="Daily card" />
        </div>

        <div class="grid grid--two">
          <div class="field">
            <label for="accountType">Type</label>
            <select id="accountType" v-model="form.type">
              <option value="bank">Bank</option>
              <option value="cash">Cash</option>
              <option value="savings">Savings</option>
            </select>
          </div>

          <div class="field">
            <label for="balance">Opening balance</label>
            <input id="balance" v-model="form.balance" inputmode="decimal" placeholder="0.00" />
          </div>
        </div>

        <button class="button button--primary button--block" type="submit">Create account</button>
      </form>

      <SectionCard>
        <div class="page-header">
          <h2>Linked flows</h2>
          <p class="muted">Accounts connect transactions, weekly review, savings, and future typed links.</p>
        </div>

        <div class="cta-list">
          <RouterLink class="button button--secondary" to="/transactions">Open transactions</RouterLink>
          <RouterLink class="button button--secondary" to="/categories">Manage categories</RouterLink>
          <RouterLink class="button button--secondary" to="/review">Open review</RouterLink>
        </div>
      </SectionCard>
    </div>

    <SectionCard v-if="financeStore.accounts.length">
      <div class="list">
        <div v-for="account in financeStore.accounts" :key="account.id" class="list-item">
          <div class="stack" style="gap: 4px;">
            <strong>{{ account.name }}</strong>
            <span class="tiny">{{ account.type }} · updated {{ account.updatedAt.slice(0, 10) }}</span>
          </div>
          <strong>{{ formatMoney(account.balanceMinor, account.currency) }}</strong>
        </div>
      </div>
    </SectionCard>

    <EmptyState
      v-else
      title="No accounts yet"
      description="Create the first account to unlock the dashboard, transaction capture, and weekly review."
    />
  </PageContainer>
</template>
