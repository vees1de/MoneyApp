import { computed, ref } from 'vue'
import { defineStore } from 'pinia'

import type { SavingsGoal } from '@/entities/savings-goal/model/types'
import { useUserStore } from '@/app/stores/user'
import { mapApiSavingsGoal } from '@/shared/api/mappers'
import { createSavingsGoal, fetchSavingsSummary, listSavingsGoals } from '@/shared/api/services/savings'
import { minorToMoneyString } from '@/shared/lib/money'

interface SavingsSummaryState {
  reservedThisMonthMinor: number
  safeToSpendMinor: number
  totalCurrentMinor: number
  totalTargetMinor: number
}

function createEmptySummary(): SavingsSummaryState {
  return {
    reservedThisMonthMinor: 0,
    safeToSpendMinor: 0,
    totalCurrentMinor: 0,
    totalTargetMinor: 0,
  }
}

export const useSavingsStore = defineStore('savings', () => {
  const goals = ref<SavingsGoal[]>([])
  const showCompleted = ref(false)
  const summary = ref<SavingsSummaryState>(createEmptySummary())
  const loading = ref(false)
  const userStore = useUserStore()

  const visibleGoals = computed(() =>
    goals.value.filter((goal) => showCompleted.value || !goal.isCompleted),
  )

  const progressRatio = computed(() => {
    if (!summary.value.totalTargetMinor) {
      return 0
    }

    return summary.value.totalCurrentMinor / summary.value.totalTargetMinor
  })

  async function fetchGoals() {
    const response = await listSavingsGoals()
    goals.value = response.items.map(mapApiSavingsGoal)
    return goals.value
  }

  async function fetchSummary() {
    const response = await fetchSavingsSummary()
    summary.value = {
      totalTargetMinor: Math.round(Number(response.total_target) * 100),
      totalCurrentMinor: Math.round(Number(response.total_current) * 100),
      reservedThisMonthMinor: Math.round(Number(response.reserved_this_month) * 100),
      safeToSpendMinor: Math.round(Number(response.safe_to_spend) * 100),
    }
    return summary.value
  }

  async function hydrateSavings() {
    loading.value = true

    try {
      await Promise.all([fetchGoals(), fetchSummary()])
    } finally {
      loading.value = false
    }
  }

  async function addGoal(input: { name: string; targetMinor: number; targetDate: string | null }) {
    const goal = await createSavingsGoal({
      title: input.name.trim(),
      targetAmount: minorToMoneyString(input.targetMinor),
      currentAmount: '0.00',
      currency: userStore.profile.currency || 'RUB',
      targetDate: input.targetDate ? new Date(input.targetDate).toISOString() : null,
      priority: 'medium',
    })

    const mapped = mapApiSavingsGoal(goal)
    goals.value = [mapped, ...goals.value]
    await fetchSummary()
    return mapped
  }

  function toggleShowCompleted() {
    showCompleted.value = !showCompleted.value
  }

  function reset() {
    goals.value = []
    showCompleted.value = false
    summary.value = createEmptySummary()
  }

  return {
    addGoal,
    fetchGoals,
    fetchSummary,
    goals,
    hydrateSavings,
    loading,
    progressRatio,
    reset,
    showCompleted,
    summary,
    toggleShowCompleted,
    visibleGoals,
  }
})
