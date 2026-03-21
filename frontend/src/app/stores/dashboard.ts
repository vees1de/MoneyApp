import { ref } from 'vue'
import { defineStore } from 'pinia'

import type { TopCategorySummary } from '@/entities/dashboard/model/types'
import { mapApiTopCategory, mapApiWeeklyReview } from '@/shared/api/mappers'
import { fetchFinanceDashboard } from '@/shared/api/services/dashboard'

interface DashboardSnapshot {
  currentBalanceMinor: number
  inflowMinor: number
  outflowMinor: number
  reviewStatus: string
  safeToSpendMinor: number
  savedThisMonthMinor: number
  topCategories: TopCategorySummary[]
  insights: string[]
}

function createEmptyDashboard(): DashboardSnapshot {
  return {
    currentBalanceMinor: 0,
    inflowMinor: 0,
    outflowMinor: 0,
    reviewStatus: 'pending',
    safeToSpendMinor: 0,
    savedThisMonthMinor: 0,
    topCategories: [],
    insights: [],
  }
}

export const useDashboardStore = defineStore('dashboard', () => {
  const period = ref<'week' | 'month'>('month')
  const dismissedCtas = ref<string[]>([])
  const snapshot = ref<DashboardSnapshot>(createEmptyDashboard())
  const loading = ref(false)

  async function fetchSnapshot() {
    loading.value = true

    try {
      const response = await fetchFinanceDashboard()
      snapshot.value = {
        currentBalanceMinor: Math.round(Number(response.current_balance) * 100),
        inflowMinor: Math.round(Number(response.monthly_income) * 100),
        outflowMinor: Math.round(Number(response.monthly_expense) * 100),
        reviewStatus: mapApiWeeklyReview(response.weekly_review).status,
        safeToSpendMinor: Math.round(Number(response.safe_to_spend) * 100),
        savedThisMonthMinor: Math.round(Number(response.saved_this_month) * 100),
        topCategories: response.top_categories.map(mapApiTopCategory),
        insights: response.insights,
      }

      return snapshot.value
    } finally {
      loading.value = false
    }
  }

  function setPeriod(nextPeriod: 'week' | 'month') {
    period.value = nextPeriod
  }

  function dismissCta(id: string) {
    if (!dismissedCtas.value.includes(id)) {
      dismissedCtas.value = [...dismissedCtas.value, id]
    }
  }

  function reset() {
    snapshot.value = createEmptyDashboard()
  }

  return {
    dismissCta,
    dismissedCtas,
    fetchSnapshot,
    loading,
    period,
    reset,
    setPeriod,
    snapshot,
  }
})
