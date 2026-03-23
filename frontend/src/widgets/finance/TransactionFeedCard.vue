<script setup lang="ts">
import type { Account } from '@/entities/account/model/types'
import type { Category } from '@/entities/category/model/types'
import type { Transaction } from '@/entities/transaction/model/types'
import { useI18n } from '@/shared/i18n'
import { formatDate } from '@/shared/lib/date'
import { formatMoney } from '@/shared/lib/money'

const props = defineProps<{
  accounts: Account[]
  categories: Category[]
  transactions: Transaction[]
}>()

const { t } = useI18n()

function category(categoryId: string | null) {
  if (!categoryId) {
    return null
  }

  return props.categories.find((c) => c.id === categoryId)
}

function accountName(accountId: string | null) {
  if (!accountId) {
    return t('common.account')
  }

  return props.accounts.find((account) => account.id === accountId)?.name ?? t('common.account')
}

function categoryName(categoryId: string | null) {
  return category(categoryId)?.name ?? t('common.category')
}

function categoryColor(categoryId: string | null) {
  return category(categoryId)?.color ?? '#AEAEB2'
}

function transactionColor(transaction: Transaction) {
  return transaction.type === 'transfer' ? 'var(--brand)' : categoryColor(transaction.categoryId)
}

function transactionLabel(transaction: Transaction) {
  if (transaction.type === 'transfer') {
    return transaction.title || t('transactions.transferBetween', {
      from: accountName(transaction.accountId),
      to: accountName(transaction.transferAccountId),
    })
  }

  return transaction.title || categoryName(transaction.categoryId)
}

function transactionMeta(transaction: Transaction) {
  if (transaction.type === 'transfer') {
    const suffix = transaction.note ? ` · ${transaction.note}` : ''
    return `${formatDate(transaction.occurredAt)} · ${accountName(transaction.accountId)} → ${accountName(transaction.transferAccountId)}${suffix}`
  }

  return `${formatDate(transaction.occurredAt)}${transaction.note ? ` · ${transaction.note}` : ''}`
}

function amountClass(transaction: Transaction) {
  if (transaction.type === 'transfer') {
    return 'feed-amount--transfer'
  }

  return transaction.type === 'income' ? 'feed-amount--income' : 'feed-amount--expense'
}

function amountText(transaction: Transaction) {
  if (transaction.type === 'transfer') {
    return formatMoney(transaction.amountMinor, transaction.currency)
  }

  return `${transaction.type === 'income' ? '+' : '−'}${formatMoney(transaction.amountMinor, transaction.currency)}`
}
</script>

<template>
  <section class="section-card">
    <div class="feed-header">
      <h3>{{ t('dashboard.recentTransactions') }}</h3>
      <RouterLink class="feed-link" to="/transactions">{{ t('common.seeAll') }}</RouterLink>
    </div>

    <div v-if="transactions.length" class="feed-list">
      <RouterLink v-for="transaction in transactions" :key="transaction.id" class="feed-row" :to="`/transactions/${transaction.id}`">
        <div class="feed-icon" :style="{ background: `${transactionColor(transaction)}18` }">
          <div class="feed-dot" :style="{ background: transactionColor(transaction) }" />
        </div>
        <div class="feed-info">
          <span class="feed-name">{{ transactionLabel(transaction) }}</span>
          <span class="feed-meta">{{ transactionMeta(transaction) }}</span>
        </div>
        <strong class="feed-amount" :class="amountClass(transaction)">
          {{ amountText(transaction) }}
        </strong>
      </RouterLink>
    </div>
    <p v-else class="tiny" style="margin:0">{{ t('dashboard.transactionsEmpty') }}</p>
  </section>
</template>

<style scoped>
.feed-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
}

.feed-header h3 {
  margin: 0;
  font-size: 1.0625rem;
  font-weight: 600;
  letter-spacing: -0.01em;
}

.feed-link {
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--brand);
}

.feed-list {
  display: flex;
  flex-direction: column;
}

.feed-row {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 11px 0;
  border-bottom: 1px solid var(--separator);
}

.feed-row:first-child { padding-top: 0; }
.feed-row:last-child { border-bottom: none; padding-bottom: 0; }

.feed-icon {
  width: 36px;
  height: 36px;
  border-radius: var(--radius-sm);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.feed-dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
}

.feed-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 1px;
  min-width: 0;
}

.feed-name {
  font-size: 0.9375rem;
  font-weight: 600;
  color: var(--text-primary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.feed-meta {
  font-size: 0.75rem;
  color: var(--text-muted);
}

.feed-amount {
  font-size: 0.9375rem;
  font-weight: 700;
  letter-spacing: -0.02em;
  white-space: nowrap;
}

.feed-amount--income { color: var(--income); }
.feed-amount--expense { color: var(--expense); }
.feed-amount--transfer { color: var(--brand); }
</style>
