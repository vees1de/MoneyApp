<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'

import { useAppUiStore } from '@/app/stores/app-ui'
import { useFinanceStore } from '@/app/stores/finance'
import { useReviewStore } from '@/app/stores/review'
import ReviewActions from '@/features/weekly-review/ReviewActions.vue'
import ReviewDeltaCard from '@/widgets/review/ReviewDeltaCard.vue'
import { useI18n } from '@/shared/i18n'
import { isDateWithinRange, formatWeekRange } from '@/shared/lib/date'
import { formatMoney, parseAmountToMinor } from '@/shared/lib/money'
import PageContainer from '@/shared/ui/PageContainer.vue'

const reviewStore = useReviewStore()
const appUiStore = useAppUiStore()
const financeStore = useFinanceStore()
const { t } = useI18n()

const actualBalanceInput = ref(
  reviewStore.review.actualBalanceMinor === null
    ? ''
    : (reviewStore.review.actualBalanceMinor / 100).toString(),
)

const canResolve = computed(() => reviewStore.review.actualBalanceMinor !== null)

const reviewRange = computed(() =>
  reviewStore.review.periodStart && reviewStore.review.periodEnd
    ? formatWeekRange(reviewStore.review.periodStart, reviewStore.review.periodEnd)
    : t('common.currentWeek'),
)

const isResolved = computed(
  () => reviewStore.review.status === 'matched' || reviewStore.review.status === 'resolved',
)

const periodTransactions = computed(() =>
  financeStore.sortedTransactions.filter((transaction) =>
    reviewStore.review.periodStart && reviewStore.review.periodEnd
      ? isDateWithinRange(transaction.occurredAt, reviewStore.review.periodStart, reviewStore.review.periodEnd)
      : false,
  ),
)

const incomeMinor = computed(() =>
  periodTransactions.value
    .filter((t) => t.kind === 'income')
    .reduce((sum, t) => sum + t.amountMinor, 0),
)

const expenseMinor = computed(() =>
  periodTransactions.value
    .filter((t) => t.kind === 'expense')
    .reduce((sum, t) => sum + t.amountMinor, 0),
)

async function submitActualBalance() {
  try {
    await reviewStore.submitActualBalance(parseAmountToMinor(actualBalanceInput.value))
    appUiStore.pushToast(t('review.balanceSubmitted'), 'success')
  } catch (error) {
    const message = error instanceof Error ? error.message : t('review.submitFailed')
    appUiStore.pushToast(message, 'warning')
  }
}

async function resolve() {
  try {
    await reviewStore.resolve(reviewStore.review.deltaMinor === 0 ? 'matched' : 'manual-confirmation')
    appUiStore.pushToast(t('review.completed'), 'success')
  } catch (error) {
    const message = error instanceof Error ? error.message : t('review.resolveFailed')
    appUiStore.pushToast(message, 'warning')
  }
}

async function skip() {
  try {
    await reviewStore.skip()
    appUiStore.pushToast(t('review.skipped'), 'warning')
  } catch (error) {
    const message = error instanceof Error ? error.message : t('review.skipFailed')
    appUiStore.pushToast(message, 'warning')
  }
}

function transactionLabel(transaction: (typeof periodTransactions.value)[number]) {
  if (transaction.type === 'transfer') {
    const fromName = financeStore.getAccountById(transaction.accountId)?.name ?? t('common.account')
    const toName =
      financeStore.getAccountById(transaction.transferAccountId)?.name ?? t('common.account')

    return transaction.title || t('transactions.transferBetween', { from: fromName, to: toName })
  }

  if (transaction.title) {
    return transaction.title
  }

  if (transaction.note) {
    return transaction.note
  }

  return financeStore.getCategoryById(transaction.categoryId)?.name ?? t('common.category')
}

function transactionAmountClass(transaction: (typeof periodTransactions.value)[number]) {
  if (transaction.type === 'transfer') {
    return 'review-txn-amount--transfer'
  }

  return transaction.type === 'income' ? 'review-txn-amount--income' : 'review-txn-amount--expense'
}

function transactionAmountText(transaction: (typeof periodTransactions.value)[number]) {
  if (transaction.type === 'transfer') {
    return formatMoney(transaction.amountMinor, transaction.currency)
  }

  return `${transaction.type === 'income' ? '+' : '−'}${formatMoney(transaction.amountMinor, transaction.currency)}`
}

onMounted(async () => {
  await Promise.all([
    financeStore.fetchTransactions({ accountId: null, type: 'all' }),
    reviewStore.fetchCurrent(),
  ])
  actualBalanceInput.value =
    reviewStore.review.actualBalanceMinor === null
      ? ''
      : (reviewStore.review.actualBalanceMinor / 100).toString()
})
</script>

<template>
  <PageContainer>
    <!-- Header -->
    <div class="review-page-header">
      <div>
        <h1 class="review-title">{{ t('review.title') }}</h1>
        <p class="review-period">{{ reviewRange }}</p>
      </div>
      <span
        class="review-status-badge"
        :class="isResolved ? 'review-status-badge--done' : 'review-status-badge--pending'"
      >
        {{ isResolved ? `✓ ${t('review.done')}` : `◔ ${t('review.pending')}` }}
      </span>
    </div>

    <!-- Period metrics -->
    <div class="review-metrics">
      <div class="review-metric">
        <span class="review-metric__label">{{ t('common.expected') }}</span>
        <strong class="review-metric__value">{{ formatMoney(reviewStore.review.expectedBalanceMinor) }}</strong>
      </div>
      <div class="review-metric review-metric--income">
        <span class="review-metric__label">{{ t('common.income') }}</span>
        <strong class="review-metric__value">+{{ formatMoney(incomeMinor) }}</strong>
      </div>
      <div class="review-metric review-metric--expense">
        <span class="review-metric__label">{{ t('common.expenses') }}</span>
        <strong class="review-metric__value">−{{ formatMoney(expenseMinor) }}</strong>
      </div>
    </div>

    <!-- Delta card -->
    <ReviewDeltaCard
      :actual-balance-minor="reviewStore.review.actualBalanceMinor"
      :delta-hint="reviewStore.deltaHint"
      :delta-minor="reviewStore.review.deltaMinor"
      :expected-balance-minor="reviewStore.review.expectedBalanceMinor"
    />

    <!-- Actual balance input -->
    <section class="section-card">
      <h2 style="margin:0 0 4px;font-size:1.0625rem;font-weight:600;letter-spacing:-0.01em">{{ t('review.enterBalanceTitle') }}</h2>
      <p class="muted" style="margin:0 0 16px;font-size:0.875rem">{{ t('review.enterBalanceBody') }}</p>
      <form class="stack" style="gap:12px" @submit.prevent="submitActualBalance">
        <div class="field">
          <label for="actualBalance">{{ t('common.actual') }}</label>
          <input
            id="actualBalance"
            v-model="actualBalanceInput"
            inputmode="decimal"
            :placeholder="t('transactionForm.placeholderAmount')"
            style="font-size:1.25rem;font-weight:600;letter-spacing:-0.02em"
          />
        </div>
        <button class="button button--primary button--block" type="submit">{{ t('review.compareBalances') }}</button>
      </form>
    </section>

    <!-- Actions -->
    <ReviewActions
      :can-resolve="canResolve"
      :delta-minor="reviewStore.review.deltaMinor"
      @resolve="resolve"
      @skip="skip"
    />

    <!-- Period transactions -->
    <section v-if="periodTransactions.length" class="section-card">
      <h2 style="margin:0 0 12px;font-size:1.0625rem;font-weight:600;letter-spacing:-0.01em">
        {{ t('common.periodEntries', { count: periodTransactions.length }) }}
      </h2>
      <div class="review-txn-list">
        <div
          v-for="transaction in periodTransactions"
          :key="transaction.id"
          class="review-txn-row"
        >
          <span class="review-txn-name">{{ transactionLabel(transaction) }}</span>
          <strong class="review-txn-amount" :class="transactionAmountClass(transaction)">
            {{ transactionAmountText(transaction) }}
          </strong>
        </div>
      </div>
    </section>
  </PageContainer>
</template>

<style scoped>
.review-page-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
  padding: 4px 0;
}

.review-title {
  margin: 0;
  font-size: 1.75rem;
  font-weight: 700;
  letter-spacing: -0.03em;
  line-height: 1.1;
}

.review-period {
  margin: 4px 0 0;
  font-size: 0.875rem;
  color: var(--text-muted);
}

.review-status-badge {
  flex-shrink: 0;
  font-size: 0.8125rem;
  font-weight: 700;
  padding: 6px 12px;
  border-radius: var(--radius-pill);
  margin-top: 4px;
}

.review-status-badge--done {
  background: var(--income-soft);
  color: var(--income);
}

.review-status-badge--pending {
  background: var(--warning-soft);
  color: var(--warning);
}

.review-metrics {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 10px;
}

.review-metric {
  display: flex;
  flex-direction: column;
  gap: 4px;
  padding: 14px 14px 12px;
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-xs);
}

.review-metric__label {
  font-size: 0.6875rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--text-muted);
}

.review-metric__value {
  font-size: 1rem;
  font-weight: 700;
  letter-spacing: -0.025em;
  color: var(--text-primary);
  line-height: 1.1;
}

.review-metric--income .review-metric__value { color: var(--income); }
.review-metric--expense .review-metric__value { color: var(--expense); }

.review-txn-list {
  display: flex;
  flex-direction: column;
}

.review-txn-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 11px 0;
  border-bottom: 1px solid var(--separator);
}

.review-txn-row:first-child { padding-top: 0; }
.review-txn-row:last-child { border-bottom: none; padding-bottom: 0; }

.review-txn-name {
  font-size: 0.9375rem;
  color: var(--text-primary);
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.review-txn-amount {
  font-size: 0.9375rem;
  font-weight: 700;
  letter-spacing: -0.02em;
  white-space: nowrap;
}

.review-txn-amount--income { color: var(--income); }
.review-txn-amount--expense { color: var(--expense); }
.review-txn-amount--transfer { color: var(--brand); }
</style>
