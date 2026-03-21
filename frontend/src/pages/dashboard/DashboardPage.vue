<script setup lang="ts">
import { computed } from 'vue'

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

const reviewTone = computed(() => (reviewStore.review.status === 'completed' ? 'success' : 'warning'))
const reviewBody = computed(() => {
  if (reviewStore.review.status === 'completed') {
    return 'The last review is closed cleanly. Keep the weekly cadence and focus on future goals.'
  }

  return 'A weekly review is waiting. Resolve balance delta before it turns into uncertainty.'
})
</script>

<template>
  <PageContainer>
    <div class="hero-grid">
      <DashboardHero
        :balance="formatMoney(financeStore.totalBalanceMinor)"
        :review-status="reviewStore.review.status"
        :safe-to-spend="formatMoney(dashboardStore.safeToSpendMinor)"
      />

      <section class="surface-card stack">
        <div class="page-header">
          <h2>Quick actions</h2>
          <p class="muted">Most-used finance actions stay one tap away.</p>
        </div>

        <div class="cta-list">
          <RouterLink class="button button--primary" to="/transactions/new?kind=expense">Add expense</RouterLink>
          <RouterLink class="button button--secondary" to="/transactions/new?kind=income">Add income</RouterLink>
          <RouterLink class="button button--secondary" to="/review">Run review</RouterLink>
          <RouterLink class="button button--secondary" to="/savings">Create goal</RouterLink>
        </div>
      </section>
    </div>

    <CashflowMetrics
      :inflow="formatMoney(dashboardStore.inflowMinor)"
      :outflow="formatMoney(dashboardStore.outflowMinor)"
      :safe-to-spend="formatMoney(dashboardStore.safeToSpendMinor)"
    />

    <div class="grid grid--two">
      <ReviewPromptCard
        :action-label="reviewStore.review.status === 'completed' ? 'Open review log' : 'Resolve weekly review'"
        :body="reviewBody"
        :tone="reviewTone"
      />
      <TopCategoriesCard :categories="dashboardStore.topCategories" />
    </div>

    <div class="grid grid--two">
      <AccountsSnapshotCard :accounts="financeStore.accounts" />
      <SavingsGoalsCard :goals="savingsStore.visibleGoals.slice(0, 3)" :progress-ratio="savingsStore.progressRatio" />
    </div>

    <TransactionFeedCard
      :categories="financeStore.categories"
      :transactions="financeStore.sortedTransactions.slice(0, 5)"
    />
  </PageContainer>
</template>
