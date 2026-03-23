<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'

import { useAppUiStore } from '@/app/stores/app-ui'
import { useFinanceStore } from '@/app/stores/finance'
import type { Transaction } from '@/entities/transaction/model/types'
import { useI18n } from '@/shared/i18n'
import { formatDate } from '@/shared/lib/date'
import { formatMoney } from '@/shared/lib/money'
import EmptyState from '@/shared/ui/EmptyState.vue'
import PageContainer from '@/shared/ui/PageContainer.vue'

const financeStore = useFinanceStore()
const appUiStore = useAppUiStore()
const route = useRoute()
const { t } = useI18n()

const canSaveCurrentFilter = computed(
  () =>
    Boolean(financeStore.filters.searchQuery.trim()) ||
    Boolean(financeStore.filters.selectedAccountId) ||
    financeStore.filters.transactionKind !== 'all',
)

function handleAccountFilterChange(event: Event) {
  const target = event.target as HTMLSelectElement | null
  financeStore.setSelectedAccount(target?.value || null)
}

function handleKindFilterChange(event: Event) {
  const target = event.target as HTMLSelectElement | null
  const nextKind = target?.value

  if (
    nextKind === 'income' ||
    nextKind === 'expense' ||
    nextKind === 'transfer' ||
    nextKind === 'all'
  ) {
    financeStore.setTransactionKindFilter(nextKind)
  }
}

function saveCurrentFilter() {
  financeStore.saveCurrentFilter()
  appUiStore.pushToast(t('transactions.filterSaved'), 'success')
}

function transactionColor(transaction: Transaction) {
  if (transaction.type === 'transfer') {
    return 'var(--brand)'
  }

  return financeStore.getCategoryById(transaction.categoryId)?.color ?? '#AEAEB2'
}

function transactionLabel(transaction: Transaction) {
  if (transaction.type === 'transfer') {
    if (transaction.title) {
      return transaction.title
    }

    const fromName = financeStore.getAccountById(transaction.accountId)?.name ?? t('common.account')
    const toName =
      financeStore.getAccountById(transaction.transferAccountId)?.name ?? t('common.account')

    return t('transactions.transferBetween', { from: fromName, to: toName })
  }

  return transaction.title || (financeStore.getCategoryById(transaction.categoryId)?.name ?? t('common.category'))
}

function transactionMeta(transaction: Transaction) {
  if (transaction.type === 'transfer') {
    const fromName = financeStore.getAccountById(transaction.accountId)?.name ?? t('common.account')
    const toName =
      financeStore.getAccountById(transaction.transferAccountId)?.name ?? t('common.account')
    const suffix = transaction.note ? ` · ${transaction.note}` : ''
    return `${formatDate(transaction.occurredAt)} · ${fromName} → ${toName}${suffix}`
  }

  return `${formatDate(transaction.occurredAt)}${transaction.note ? ` · ${transaction.note}` : ''}`
}

function amountText(transaction: Transaction) {
  if (transaction.type === 'transfer') {
    return formatMoney(transaction.amountMinor, transaction.currency)
  }

  return `${transaction.type === 'income' ? '+' : '−'}${formatMoney(transaction.amountMinor, transaction.currency)}`
}

function amountClass(transaction: Transaction) {
  if (transaction.type === 'transfer') {
    return 'txn-amount--transfer'
  }

  return transaction.type === 'income' ? 'txn-amount--income' : 'txn-amount--expense'
}

onMounted(async () => {
  if (typeof route.query.account === 'string') {
    financeStore.setSelectedAccount(route.query.account)
  }

  await Promise.all([financeStore.fetchAccounts(), financeStore.fetchCategories(), financeStore.fetchTransactions()])
})
</script>

<template>
  <PageContainer>
    <div class="page-header">
      <h1>{{ t('transactions.title') }}</h1>
      <p class="muted">{{ t('transactions.subtitle') }}</p>
    </div>

    <section class="section-card action-strip">
      <RouterLink class="action-pill action-pill--expense" to="/transactions/new?kind=expense">
        {{ t('dashboard.quickExpense') }}
      </RouterLink>
      <RouterLink class="action-pill action-pill--income" to="/transactions/new?kind=income">
        {{ t('dashboard.quickIncome') }}
      </RouterLink>
      <RouterLink class="action-pill" to="/transactions/transfer">
        {{ t('common.transfer') }}
      </RouterLink>
    </section>

    <section class="section-card">
      <div class="field" style="margin-bottom:12px">
        <label for="transactionSearch">{{ t('common.search') }}</label>
        <input
          id="transactionSearch"
          :value="financeStore.filters.searchQuery"
          :placeholder="t('transactions.searchPlaceholder')"
          @input="financeStore.setSearchQuery(($event.target as HTMLInputElement).value)"
        />
      </div>

      <div class="txn-filter-header">
        <h2 style="margin:0;font-size:1.0625rem;font-weight:600;letter-spacing:-0.01em">{{ t('common.filter') }}</h2>
        <div class="txn-filter-actions">
          <button
            class="filter-link"
            :disabled="!canSaveCurrentFilter"
            type="button"
            @click="saveCurrentFilter"
          >
            {{ t('transactions.saveFilter') }}
          </button>
          <RouterLink class="filter-link" to="/categories">{{ t('route.categories') }}</RouterLink>
        </div>
      </div>

      <div class="grid grid--two" style="margin-top:12px">
        <div class="field">
          <label for="accountFilter">{{ t('common.account') }}</label>
          <select
            id="accountFilter"
            :value="financeStore.filters.selectedAccountId ?? ''"
            @change="handleAccountFilterChange"
          >
            <option value="">{{ t('common.allAccounts') }}</option>
            <option v-for="account in financeStore.accounts" :key="account.id" :value="account.id">
              {{ account.name }}
            </option>
          </select>
        </div>

        <div class="field">
          <label for="kindFilter">{{ t('common.type') }}</label>
          <select
            id="kindFilter"
            :value="financeStore.filters.transactionKind"
            @change="handleKindFilterChange"
          >
            <option value="all">{{ t('transactions.kindAll') }}</option>
            <option value="income">{{ t('transactions.kindIncome') }}</option>
            <option value="expense">{{ t('transactions.kindExpense') }}</option>
            <option value="transfer">{{ t('transactions.kindTransfer') }}</option>
          </select>
        </div>
      </div>
    </section>

    <section v-if="financeStore.savedFilters.length" class="section-card">
      <div class="txn-filter-header" style="margin-bottom:12px">
        <h2 style="margin:0;font-size:1.0625rem;font-weight:600;letter-spacing:-0.01em">{{ t('transactions.savedFilters') }}</h2>
        <button
          v-if="financeStore.filters.searchQuery"
          class="filter-link"
          type="button"
          @click="financeStore.clearSearch()"
        >
          {{ t('transactions.clearSearch') }}
        </button>
      </div>

      <div class="saved-filter-list">
        <div v-for="savedFilter in financeStore.savedFilters" :key="savedFilter.id" class="saved-filter-chip">
          <button class="saved-filter-chip__label" type="button" @click="financeStore.applySavedFilter(savedFilter.id)">
            {{ savedFilter.label }}
          </button>
          <button
            class="saved-filter-chip__remove"
            type="button"
            @click="financeStore.removeSavedFilter(savedFilter.id)"
          >
            ×
          </button>
        </div>
      </div>
    </section>

    <section v-if="financeStore.filteredTransactions.length" class="section-card">
      <div class="txn-list">
        <RouterLink
          v-for="transaction in financeStore.filteredTransactions"
          :key="transaction.id"
          class="txn-row"
          :to="`/transactions/${transaction.id}`"
        >
          <div class="txn-icon" :style="{ background: `${transactionColor(transaction)}18` }">
            <div class="txn-dot" :style="{ background: transactionColor(transaction) }" />
          </div>
          <div class="txn-info">
            <span class="txn-name">{{ transactionLabel(transaction) }}</span>
            <span class="txn-meta">{{ transactionMeta(transaction) }}</span>
          </div>
          <strong class="txn-amount" :class="amountClass(transaction)">
            {{ amountText(transaction) }}
          </strong>
        </RouterLink>
      </div>
    </section>

    <EmptyState
      v-else
      icon="💸"
      :title="t('transactions.emptyTitle')"
      :description="t('transactions.emptyDescription')"
    />
  </PageContainer>
</template>

<style scoped>
.action-strip {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.action-pill {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-height: 40px;
  padding: 0 16px;
  border-radius: var(--radius-pill);
  background: var(--surface-secondary);
  color: var(--text-primary);
  font-size: 0.875rem;
  font-weight: 700;
}

.action-pill--expense {
  background: var(--danger-soft);
  color: var(--expense);
}

.action-pill--income {
  background: var(--income-soft);
  color: var(--income);
}

.txn-filter-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.txn-filter-actions {
  display: flex;
  align-items: center;
  gap: 12px;
}

.filter-link {
  font-size: 0.875rem;
  font-weight: 600;
  color: var(--brand);
  border: none;
  background: transparent;
  padding: 0;
}

.filter-link:disabled {
  color: var(--text-muted);
}

.saved-filter-list {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.saved-filter-chip {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  min-height: 36px;
  padding: 0 8px 0 14px;
  border-radius: var(--radius-pill);
  background: var(--surface-secondary);
}

.saved-filter-chip__label,
.saved-filter-chip__remove {
  border: none;
  background: transparent;
  padding: 0;
}

.saved-filter-chip__label {
  color: var(--text-primary);
  font-size: 0.8125rem;
  font-weight: 700;
}

.saved-filter-chip__remove {
  color: var(--text-muted);
  font-size: 1rem;
  line-height: 1;
}

.txn-list {
  display: flex;
  flex-direction: column;
}

.txn-row {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 0;
  border-bottom: 1px solid var(--separator);
}

.txn-row:first-child { padding-top: 0; }
.txn-row:last-child { border-bottom: none; padding-bottom: 0; }

.txn-icon {
  width: 36px;
  height: 36px;
  border-radius: var(--radius-sm);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.txn-dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
}

.txn-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 1px;
  min-width: 0;
}

.txn-name {
  font-size: 0.9375rem;
  font-weight: 600;
  color: var(--text-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.txn-meta {
  font-size: 0.75rem;
  color: var(--text-muted);
}

.txn-amount {
  font-size: 0.9375rem;
  font-weight: 700;
  letter-spacing: -0.02em;
  white-space: nowrap;
}

.txn-amount--income { color: var(--income); }
.txn-amount--expense { color: var(--expense); }
.txn-amount--transfer { color: var(--brand); }
</style>
