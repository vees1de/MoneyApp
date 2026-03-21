<script setup lang="ts">
import { onMounted, watch } from 'vue'

import { formatDate } from '@/shared/lib/date'
import { formatMoney } from '@/shared/lib/money'
import { useFinanceStore } from '@/app/stores/finance'
import QuickTransactionComposer from '@/features/upsert-transaction/QuickTransactionComposer.vue'
import EmptyState from '@/shared/ui/EmptyState.vue'
import PageContainer from '@/shared/ui/PageContainer.vue'

const financeStore = useFinanceStore()

function categoryName(categoryId: string) {
  return financeStore.categories.find((c) => c.id === categoryId)?.name ?? 'Category'
}

function categoryColor(categoryId: string) {
  return financeStore.categories.find((c) => c.id === categoryId)?.color ?? '#AEAEB2'
}

function handleAccountFilterChange(event: Event) {
  const target = event.target as HTMLSelectElement | null
  financeStore.setSelectedAccount(target?.value || null)
}

function handleKindFilterChange(event: Event) {
  const target = event.target as HTMLSelectElement | null
  const nextKind = target?.value
  if (nextKind === 'income' || nextKind === 'expense' || nextKind === 'all') {
    financeStore.setTransactionKindFilter(nextKind)
  }
}

onMounted(async () => {
  await Promise.all([financeStore.fetchAccounts(), financeStore.fetchCategories(), financeStore.fetchTransactions()])
})

watch(
  () => [financeStore.filters.selectedAccountId, financeStore.filters.transactionKind],
  async () => {
    await financeStore.fetchTransactions()
  },
)
</script>

<template>
  <PageContainer>
    <div class="page-header">
      <h1>Transactions</h1>
      <p class="muted">Log income and expenses. Filter to find what you need.</p>
    </div>

    <!-- Quick add form -->
    <QuickTransactionComposer />

    <!-- Filters -->
    <section class="section-card">
      <div class="txn-filter-header">
        <h2 style="margin:0;font-size:1.0625rem;font-weight:600;letter-spacing:-0.01em">Filter</h2>
        <RouterLink class="filter-link" to="/categories">Categories</RouterLink>
      </div>

      <div class="grid grid--two" style="margin-top:12px">
        <div class="field">
          <label for="accountFilter">Account</label>
          <select
            id="accountFilter"
            :value="financeStore.filters.selectedAccountId ?? ''"
            @change="handleAccountFilterChange"
          >
            <option value="">All accounts</option>
            <option v-for="account in financeStore.accounts" :key="account.id" :value="account.id">
              {{ account.name }}
            </option>
          </select>
        </div>

        <div class="field">
          <label for="kindFilter">Type</label>
          <select
            id="kindFilter"
            :value="financeStore.filters.transactionKind"
            @change="handleKindFilterChange"
          >
            <option value="all">All</option>
            <option value="income">Income</option>
            <option value="expense">Expense</option>
          </select>
        </div>
      </div>
    </section>

    <!-- Transaction list -->
    <section v-if="financeStore.filteredTransactions.length" class="section-card">
      <div class="txn-list">
        <div
          v-for="transaction in financeStore.filteredTransactions"
          :key="transaction.id"
          class="txn-row"
        >
          <div
            class="txn-icon"
            :style="{ background: `${categoryColor(transaction.categoryId)}18` }"
          >
            <div
              class="txn-dot"
              :style="{ background: categoryColor(transaction.categoryId) }"
            />
          </div>
          <div class="txn-info">
            <span class="txn-name">{{ categoryName(transaction.categoryId) }}</span>
            <span class="txn-meta">
              {{ formatDate(transaction.occurredAt) }}<span v-if="transaction.note"> · {{ transaction.note }}</span>
            </span>
          </div>
          <strong
            class="txn-amount"
            :class="transaction.kind === 'income' ? 'txn-amount--income' : 'txn-amount--expense'"
          >
            {{ transaction.kind === 'income' ? '+' : '−' }}{{ formatMoney(transaction.amountMinor, transaction.currency) }}
          </strong>
        </div>
      </div>
    </section>

    <EmptyState
      v-else
      icon="💸"
      title="No transactions found"
      description="Adjust filters or add a new entry above."
    />
  </PageContainer>
</template>

<style scoped>
.txn-filter-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.filter-link {
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--brand);
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
</style>
