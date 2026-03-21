<script setup lang="ts">
import { computed, onMounted } from 'vue'

import { useDashboardStore } from '@/app/stores/dashboard'
import { useFinanceStore } from '@/app/stores/finance'
import { useReviewStore } from '@/app/stores/review'
import { useSavingsStore } from '@/app/stores/savings'
import DashboardHero from '@/widgets/dashboard/DashboardHero.vue'
import CashflowMetrics from '@/widgets/dashboard/CashflowMetrics.vue'
import ReviewPromptCard from '@/widgets/dashboard/ReviewPromptCard.vue'
import TopCategoriesCard from '@/widgets/dashboard/TopCategoriesCard.vue'
import AccountsSnapshotCard from '@/widgets/finance/AccountsSnapshotCard.vue'
import TransactionFeedCard from '@/widgets/finance/TransactionFeedCard.vue'
import SavingsGoalsCard from '@/widgets/savings/SavingsGoalsCard.vue'
import { formatMoney } from '@/shared/lib/money'
import PageContainer from '@/shared/ui/PageContainer.vue'

const dashboardStore = useDashboardStore()
const financeStore = useFinanceStore()
const savingsStore = useSavingsStore()
const reviewStore = useReviewStore()

const reviewTone = computed(() =>
  reviewStore.review.status === 'matched' || reviewStore.review.status === 'resolved'
    ? 'success'
    : 'warning',
)

const reviewBody = computed(() => {
  if (reviewStore.review.status === 'matched' || reviewStore.review.status === 'resolved') {
    return 'Last week closed cleanly. Keep the cadence and stay ahead.'
  }
  return 'A weekly review is waiting. Resolve the delta before uncertainty compounds.'
})

const reviewActionLabel = computed(() =>
  reviewStore.review.status === 'matched' || reviewStore.review.status === 'resolved'
    ? 'Open review log'
    : 'Resolve weekly review',
)

onMounted(async () => {
  await Promise.all([
    financeStore.hydrateFinance(),
    savingsStore.hydrateSavings(),
    reviewStore.fetchCurrent(),
    dashboardStore.fetchSnapshot(),
  ])
})
</script>

<template>
  <PageContainer>
    <!-- Hero balance card -->
    <DashboardHero
      :balance="formatMoney(dashboardStore.snapshot.currentBalanceMinor)"
      :review-status="reviewStore.review.status"
      :safe-to-spend="formatMoney(dashboardStore.snapshot.safeToSpendMinor)"
    />

    <!-- Cashflow metrics row -->
    <CashflowMetrics
      :inflow="formatMoney(dashboardStore.snapshot.inflowMinor)"
      :outflow="formatMoney(dashboardStore.snapshot.outflowMinor)"
      :safe-to-spend="formatMoney(dashboardStore.snapshot.safeToSpendMinor)"
    />

    <!-- Quick actions -->
    <div class="quick-actions">
      <RouterLink class="qa-btn qa-btn--expense" to="/transactions/new?kind=expense">
        <span class="qa-btn__icon">−</span>
        <span class="qa-btn__label">Expense</span>
      </RouterLink>
      <RouterLink class="qa-btn qa-btn--income" to="/transactions/new?kind=income">
        <span class="qa-btn__icon">+</span>
        <span class="qa-btn__label">Income</span>
      </RouterLink>
      <RouterLink class="qa-btn" to="/review">
        <span class="qa-btn__icon">↻</span>
        <span class="qa-btn__label">Review</span>
      </RouterLink>
      <RouterLink class="qa-btn" to="/savings">
        <span class="qa-btn__icon">🎯</span>
        <span class="qa-btn__label">Goal</span>
      </RouterLink>
    </div>

    <!-- Review prompt -->
    <ReviewPromptCard
      :action-label="reviewActionLabel"
      :body="reviewBody"
      :tone="reviewTone"
    />

    <!-- Two-column: top categories + savings -->
    <div class="grid grid--two">
      <TopCategoriesCard :categories="dashboardStore.snapshot.topCategories" />
      <SavingsGoalsCard
        :goals="savingsStore.visibleGoals.slice(0, 3)"
        :progress-ratio="savingsStore.progressRatio"
      />
    </div>

    <!-- Accounts snapshot -->
    <AccountsSnapshotCard :accounts="financeStore.accounts" />

    <!-- Recent transactions feed -->
    <TransactionFeedCard
      :categories="financeStore.categories"
      :transactions="financeStore.sortedTransactions.slice(0, 6)"
    />
  </PageContainer>
</template>

<style scoped>
.quick-actions {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 8px;
}

.qa-btn {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
  padding: 12px 8px 10px;
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-xs);
  text-decoration: none;
  transition: transform var(--duration-fast) ease, opacity var(--duration-fast) ease;
}

.qa-btn:active {
  transform: scale(0.95);
  opacity: 0.80;
}

.qa-btn__icon {
  font-size: 1.25rem;
  font-weight: 700;
  line-height: 1;
  color: var(--text-primary);
}

.qa-btn__label {
  font-size: 0.6875rem;
  font-weight: 600;
  color: var(--text-secondary);
  letter-spacing: 0.01em;
}

.qa-btn--expense .qa-btn__icon {
  color: var(--expense);
}

.qa-btn--income .qa-btn__icon {
  color: var(--income);
}
</style>
