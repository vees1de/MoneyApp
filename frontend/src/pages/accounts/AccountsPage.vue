<script setup lang="ts">
import { onMounted, reactive } from 'vue'

import { useAppUiStore } from '@/app/stores/app-ui'
import { useFinanceStore } from '@/app/stores/finance'
import { useUserStore } from '@/app/stores/user'
import { parseAmountToMinor, formatMoney } from '@/shared/lib/money'
import EmptyState from '@/shared/ui/EmptyState.vue'
import PageContainer from '@/shared/ui/PageContainer.vue'

const financeStore = useFinanceStore()
const appUiStore = useAppUiStore()
const userStore = useUserStore()

const form = reactive({
  balance: '',
  name: '',
  type: 'bank' as 'bank' | 'cash' | 'savings',
})

const typeIcon: Record<string, string> = { bank: '🏦', cash: '💵', savings: '🎯' }

function submit() {
  if (!form.name.trim()) {
    appUiStore.pushToast('Account name is required.', 'warning')
    return
  }

  void (async () => {
    try {
      await financeStore.addAccount({
        name: form.name,
        type: form.type,
        balanceMinor: parseAmountToMinor(form.balance),
        currency: userStore.profile.currency || 'RUB',
      })
      appUiStore.pushToast('Account added.', 'success')
      form.balance = ''
      form.name = ''
      form.type = 'bank'
    } catch (error) {
      const message = error instanceof Error ? error.message : 'Account creation failed.'
      appUiStore.pushToast(message, 'warning')
    }
  })()
}

onMounted(async () => {
  await financeStore.fetchAccounts()
})
</script>

<template>
  <PageContainer>
    <div class="page-header">
      <h1>Accounts</h1>
      <p class="muted">Your financial accounts, all in one place.</p>
    </div>

    <!-- Account list -->
    <section v-if="financeStore.accounts.length" class="section-card">
      <div class="acct-list">
        <div v-for="account in financeStore.accounts" :key="account.id" class="acct-row">
          <div class="acct-icon">{{ typeIcon[account.type] ?? '🏦' }}</div>
          <div class="acct-body">
            <span class="acct-name">{{ account.name }}</span>
            <span class="acct-meta">{{ account.type }} · {{ account.updatedAt.slice(0, 10) }}</span>
          </div>
          <strong class="acct-balance">{{ formatMoney(account.balanceMinor, account.currency) }}</strong>
        </div>
      </div>
    </section>

    <EmptyState
      v-else
      icon="🏦"
      title="No accounts yet"
      description="Add your first account to start tracking balances and transactions."
    />

    <!-- Add account form -->
    <section class="section-card">
      <h2 style="margin:0 0 16px;font-size:1.0625rem;font-weight:600;letter-spacing:-0.01em">Add account</h2>
      <form class="stack" style="gap:14px" @submit.prevent="submit">
        <div class="field">
          <label for="accountName">Account name</label>
          <input id="accountName" v-model="form.name" placeholder="e.g. Daily card" />
        </div>

        <div class="grid grid--two">
          <div class="field">
            <label for="accountType">Type</label>
            <select id="accountType" v-model="form.type">
              <option value="bank">🏦 Bank</option>
              <option value="cash">💵 Cash</option>
              <option value="savings">🎯 Savings</option>
            </select>
          </div>

          <div class="field">
            <label for="balance">Opening balance</label>
            <input id="balance" v-model="form.balance" inputmode="decimal" placeholder="0.00" />
          </div>
        </div>

        <button class="button button--primary button--block" type="submit">Create account</button>
      </form>
    </section>

    <!-- Navigation shortcuts -->
    <section class="section-card">
      <p class="muted" style="margin:0 0 12px;font-size:0.875rem">Jump to</p>
      <div class="cta-list">
        <RouterLink class="button button--secondary" to="/transactions">Transactions</RouterLink>
        <RouterLink class="button button--secondary" to="/categories">Categories</RouterLink>
        <RouterLink class="button button--secondary" to="/review">Review</RouterLink>
      </div>
    </section>
  </PageContainer>
</template>

<style scoped>
.acct-list {
  display: flex;
  flex-direction: column;
}

.acct-row {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 0;
  border-bottom: 1px solid var(--separator);
}

.acct-row:first-child { padding-top: 0; }
.acct-row:last-child { border-bottom: none; padding-bottom: 0; }

.acct-icon {
  width: 38px;
  height: 38px;
  border-radius: var(--radius-sm);
  background: var(--surface-secondary);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.125rem;
  flex-shrink: 0;
}

.acct-body {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 1px;
}

.acct-name {
  font-size: 0.9375rem;
  font-weight: 600;
  color: var(--text-primary);
}

.acct-meta {
  font-size: 0.75rem;
  color: var(--text-muted);
  text-transform: capitalize;
}

.acct-balance {
  font-size: 0.9375rem;
  font-weight: 700;
  letter-spacing: -0.02em;
  white-space: nowrap;
}
</style>
