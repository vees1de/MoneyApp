<script setup lang="ts">
import type { Account } from '@/entities/account/model/types'
import { translateAccountType, useI18n } from '@/shared/i18n'
import { formatMoney } from '@/shared/lib/money'

defineProps<{
  accounts: Account[]
}>()

const { t } = useI18n()

const typeIcon: Record<string, string> = {
  bank: '🏦',
  cash: '💵',
  savings: '🎯',
}
</script>

<template>
  <section class="section-card">
    <div class="acct-header">
      <h3>{{ t('dashboard.accountsTitle') }}</h3>
      <RouterLink class="acct-link" to="/accounts">{{ t('common.manage') }}</RouterLink>
    </div>

    <div class="acct-list">
      <div v-for="account in accounts" :key="account.id" class="acct-row">
        <div class="acct-icon">{{ typeIcon[account.type] ?? '🏦' }}</div>
        <div class="acct-info">
          <span class="acct-name">{{ account.name }}</span>
          <span class="acct-type">{{ translateAccountType(account.type) }}</span>
        </div>
        <strong class="acct-balance">{{ formatMoney(account.balanceMinor, account.currency) }}</strong>
      </div>
    </div>
  </section>
</template>

<style scoped>
.acct-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 14px;
}

.acct-header h3 {
  margin: 0;
  font-size: 1.0625rem;
  font-weight: 600;
  letter-spacing: -0.015em;
}

.acct-link {
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--brand);
  transition: opacity var(--duration-fast) var(--ease-out);
}

.acct-link:hover {
  opacity: 0.7;
}

.acct-list {
  display: flex;
  flex-direction: column;
}

.acct-row {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 0;
  border-bottom: 0.5px solid var(--separator);
}

.acct-row:first-child {
  padding-top: 0;
}

.acct-row:last-child {
  border-bottom: none;
  padding-bottom: 0;
}

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

.acct-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 2px;
  min-width: 0;
}

.acct-name {
  font-size: 0.9375rem;
  font-weight: 600;
  color: var(--text-primary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  letter-spacing: -0.01em;
}

.acct-type {
  font-size: 0.75rem;
  color: var(--text-muted);
  text-transform: capitalize;
}

.acct-balance {
  font-size: 0.9375rem;
  font-weight: 700;
  letter-spacing: -0.02em;
  color: var(--text-primary);
  white-space: nowrap;
  font-variant-numeric: tabular-nums;
}
</style>
