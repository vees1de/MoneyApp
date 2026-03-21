<script setup lang="ts">
import { formatDate } from '@/shared/lib/date'
import { formatMoney } from '@/shared/lib/money'
import { useFinanceStore } from '@/app/stores/finance'
import QuickTransactionComposer from '@/features/upsert-transaction/QuickTransactionComposer.vue'
import EmptyState from '@/shared/ui/EmptyState.vue'
import PageContainer from '@/shared/ui/PageContainer.vue'
import SectionCard from '@/shared/ui/SectionCard.vue'

const financeStore = useFinanceStore()

function categoryName(categoryId: string) {
  return financeStore.categories.find((category) => category.id === categoryId)?.name ?? 'Category'
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
</script>

<template>
  <PageContainer>
    <div class="page-header">
      <h1>Transactions</h1>
      <p class="muted">Fast entry first, then filter by account and type without leaving the screen.</p>
    </div>

    <QuickTransactionComposer />

    <SectionCard>
      <div class="split">
        <div class="page-header">
          <h2>Ledger filters</h2>
          <p class="muted">Keep the list narrow on mobile and wide enough for finance review.</p>
        </div>

        <RouterLink class="button button--secondary" to="/categories">Categories</RouterLink>
      </div>

      <div class="grid grid--two">
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
    </SectionCard>

    <SectionCard v-if="financeStore.filteredTransactions.length">
      <div class="list">
        <div v-for="transaction in financeStore.filteredTransactions" :key="transaction.id" class="list-item">
          <div class="stack" style="gap: 4px;">
            <strong>{{ categoryName(transaction.categoryId) }}</strong>
            <span class="tiny">{{ formatDate(transaction.occurredAt) }} · {{ transaction.note || 'No note' }}</span>
          </div>
          <strong :class="transaction.kind === 'income' ? 'amount--income' : 'amount--expense'">
            {{ transaction.kind === 'income' ? '+' : '-' }}{{ formatMoney(transaction.amountMinor, transaction.currency) }}
          </strong>
        </div>
      </div>
    </SectionCard>

    <EmptyState
      v-else
      title="No transactions for these filters"
      description="Change the active filters or record a new entry to repopulate the ledger."
    />
  </PageContainer>
</template>
