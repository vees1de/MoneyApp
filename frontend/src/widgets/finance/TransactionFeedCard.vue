<script setup lang="ts">
import type { Category } from '@/entities/category/model/types'
import type { Transaction } from '@/entities/transaction/model/types'
import { formatDate } from '@/shared/lib/date'
import { formatMoney } from '@/shared/lib/money'

const props = defineProps<{
  categories: Category[]
  transactions: Transaction[]
}>()

function categoryName(categoryId: string) {
  return props.categories.find((category) => category.id === categoryId)?.name ?? 'Category'
}
</script>

<template>
  <section class="surface-card stack">
    <div class="section-card__header">
      <div>
        <h3>Recent transactions</h3>
        <p class="muted">Latest recorded entries with account-level impact.</p>
      </div>
      <RouterLink class="button button--secondary" to="/transactions">Open ledger</RouterLink>
    </div>

    <div v-if="transactions.length" class="list">
      <div v-for="transaction in transactions" :key="transaction.id" class="list-item">
        <div class="stack" style="gap: 4px;">
          <strong>{{ categoryName(transaction.categoryId) }}</strong>
          <span class="tiny">{{ formatDate(transaction.occurredAt) }} · {{ transaction.note || 'No note' }}</span>
        </div>
        <strong :class="transaction.kind === 'income' ? 'amount--income' : 'amount--expense'">
          {{ transaction.kind === 'income' ? '+' : '-' }}{{ formatMoney(transaction.amountMinor, transaction.currency) }}
        </strong>
      </div>
    </div>
    <p v-else class="tiny">The ledger is empty. Add the first transaction to start building signals.</p>
  </section>
</template>
