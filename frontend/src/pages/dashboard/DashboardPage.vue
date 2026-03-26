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
import { useI18n } from '@/shared/i18n'
import { formatMoney } from '@/shared/lib/money'
import PageContainer from '@/shared/ui/PageContainer.vue'

const dashboardStore = useDashboardStore()
const financeStore = useFinanceStore()
const savingsStore = useSavingsStore()
const reviewStore = useReviewStore()
const { t } = useI18n()

const reviewTone = computed(() =>
  reviewStore.review.status === 'matched' || reviewStore.review.status === 'resolved'
    ? 'success'
    : 'warning',
)

const reviewBody = computed(() => {
  if (reviewStore.review.status === 'matched' || reviewStore.review.status === 'resolved') {
    return t('dashboard.reviewHealthy')
  }
  return t('dashboard.reviewPending')
})

const reviewActionLabel = computed(() =>
  reviewStore.review.status === 'matched' || reviewStore.review.status === 'resolved'
    ? t('dashboard.openReviewLog')
    : t('dashboard.resolveReview'),
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
    <DashboardHero
      :balance="formatMoney(dashboardStore.snapshot.currentBalanceMinor)"
      :review-status="reviewStore.review.status"
      :safe-to-spend="formatMoney(dashboardStore.snapshot.safeToSpendMinor)"
    />

    <CashflowMetrics
      :inflow="formatMoney(dashboardStore.snapshot.inflowMinor)"
      :outflow="formatMoney(dashboardStore.snapshot.outflowMinor)"
      :safe-to-spend="formatMoney(dashboardStore.snapshot.safeToSpendMinor)"
    />

    <div class="quick-actions">
      <RouterLink class="qa-btn qa-btn--expense" to="/transactions/new?kind=expense">
        <span class="qa-btn__icon">−</span>
        <span class="qa-btn__label">{{ t('dashboard.quickExpense') }}</span>
      </RouterLink>
      <RouterLink class="qa-btn qa-btn--income" to="/transactions/new?kind=income">
        <span class="qa-btn__icon">+</span>
        <span class="qa-btn__label">{{ t('dashboard.quickIncome') }}</span>
      </RouterLink>
      <RouterLink class="qa-btn" to="/review">
        <span class="qa-btn__icon">↻</span>
        <span class="qa-btn__label">{{ t('dashboard.quickReview') }}</span>
      </RouterLink>
      <RouterLink class="qa-btn" to="/savings">
        <span class="qa-btn__icon qa-btn__icon--target">●</span>
        <span class="qa-btn__label">{{ t('dashboard.quickGoal') }}</span>
      </RouterLink>
    </div>

    <ReviewPromptCard
      :action-label="reviewActionLabel"
      :body="reviewBody"
      :tone="reviewTone"
    />

    <div class="grid grid--two">
      <TopCategoriesCard :categories="dashboardStore.snapshot.topCategories" />
      <SavingsGoalsCard
        :goals="savingsStore.visibleGoals.slice(0, 3)"
        :progress-ratio="savingsStore.progressRatio"
      />
    </div>

    <AccountsSnapshotCard :accounts="financeStore.accounts" />

    <TransactionFeedCard
      :accounts="financeStore.accounts"
      :categories="financeStore.categories"
      :transactions="financeStore.sortedTransactions.slice(0, 6)"
    />
  </PageContainer>
</template>

<style scoped>
.quick-actions {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 10px;
}

.qa-btn {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  padding: 16px 8px 14px;
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-xs);
  text-decoration: none;
  transition: transform var(--duration-fast) var(--ease-out),
              box-shadow var(--duration-base) var(--ease-out),
              background-color var(--duration-base) var(--ease-out);
}

.qa-btn:hover {
  background: var(--surface-muted);
  box-shadow: var(--shadow-sm);
  transform: translateY(-1px);
}

.qa-btn:active {
  transform: scale(0.95);
}

.qa-btn__icon {
  font-size: 1.375rem;
  font-weight: 700;
  line-height: 1;
  color: var(--text-primary);
  width: 36px;
  height: 36px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: var(--radius-sm);
  background: var(--surface-secondary);
}

.qa-btn__icon--target {
  font-size: 0.75rem;
  color: var(--brand);
}

.qa-btn__label {
  font-size: 0.6875rem;
  font-weight: 600;
  color: var(--text-secondary);
  letter-spacing: 0.01em;
  text-align: center;
  line-height: 1.3;
}

.qa-btn--expense .qa-btn__icon {
  color: var(--expense);
  background: var(--danger-soft);
}

.qa-btn--income .qa-btn__icon {
  color: var(--income);
  background: var(--income-soft);
}
</style>
