<script setup lang="ts">
import { computed, ref } from 'vue'

import { useAppUiStore } from '@/app/stores/app-ui'
import { useReviewStore } from '@/app/stores/review'
import ReviewActions from '@/features/weekly-review/ReviewActions.vue'
import ReviewDeltaCard from '@/widgets/review/ReviewDeltaCard.vue'
import { formatWeekRange } from '@/shared/lib/date'
import { formatMoney, parseAmountToMinor } from '@/shared/lib/money'
import PageContainer from '@/shared/ui/PageContainer.vue'
import SectionCard from '@/shared/ui/SectionCard.vue'
import StatCard from '@/shared/ui/StatCard.vue'

const reviewStore = useReviewStore()
const appUiStore = useAppUiStore()

const actualBalanceInput = ref(
  reviewStore.review.actualBalanceMinor === null ? '' : (reviewStore.review.actualBalanceMinor / 100).toString(),
)

const canResolve = computed(() => reviewStore.review.actualBalanceMinor !== null)

function submitActualBalance() {
  reviewStore.submitActualBalance(parseAmountToMinor(actualBalanceInput.value))
  appUiStore.pushToast('Actual balance submitted.', 'success')
}

function resolve() {
  reviewStore.resolve(reviewStore.deltaMinor === 0 ? 'matched' : 'manual-confirmation')
  appUiStore.pushToast('Weekly review completed.', 'success')
}

function skip() {
  reviewStore.skip()
  appUiStore.pushToast('Weekly review skipped.', 'warning')
}
</script>

<template>
  <PageContainer>
    <div class="page-header">
      <h1>Weekly review</h1>
      <p class="muted">
        {{ formatWeekRange(reviewStore.review.periodStart, reviewStore.review.periodEnd) }} · expected balance is
        recalculated from opening balance, income, and expenses.
      </p>
      <span class="pill" :class="reviewStore.review.status === 'completed' ? 'pill--success' : 'pill--warning'">
        {{ reviewStore.review.status }}
      </span>
    </div>

    <div class="metrics-grid">
      <StatCard label="Opening balance" :value="formatMoney(reviewStore.review.openingBalanceMinor)" />
      <StatCard label="Income in period" :value="formatMoney(reviewStore.incomeMinor)" />
      <StatCard label="Expenses in period" :value="formatMoney(reviewStore.expenseMinor)" />
    </div>

    <ReviewDeltaCard
      :actual-balance-minor="reviewStore.review.actualBalanceMinor"
      :delta-hint="reviewStore.deltaHint"
      :delta-minor="reviewStore.deltaMinor"
      :expected-balance-minor="reviewStore.expectedBalanceMinor"
    />

    <SectionCard>
      <form class="stack" @submit.prevent="submitActualBalance">
        <div class="page-header">
          <h2>Submit actual balance</h2>
          <p class="muted">Use the real balance from the account statement or banking app.</p>
        </div>

        <div class="field">
          <label for="actualBalance">Actual balance</label>
          <input id="actualBalance" v-model="actualBalanceInput" inputmode="decimal" placeholder="0.00" />
        </div>

        <button class="button button--primary" type="submit">Compare balances</button>
      </form>
    </SectionCard>

    <ReviewActions :can-resolve="canResolve" :delta-minor="reviewStore.deltaMinor" @resolve="resolve" @skip="skip" />

    <SectionCard v-if="reviewStore.periodTransactions.length">
      <div class="page-header">
        <h2>Transactions in this review window</h2>
        <p class="muted">Use this list to quickly sanity-check suspicious entries before resolving the week.</p>
      </div>

      <div class="list">
        <div v-for="transaction in reviewStore.periodTransactions" :key="transaction.id" class="list-item">
          <span>{{ transaction.note || transaction.categoryId }}</span>
          <strong :class="transaction.kind === 'income' ? 'amount--income' : 'amount--expense'">
            {{ transaction.kind === 'income' ? '+' : '-' }}{{ formatMoney(transaction.amountMinor, transaction.currency) }}
          </strong>
        </div>
      </div>
    </SectionCard>
  </PageContainer>
</template>
