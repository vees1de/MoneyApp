import { computed, ref } from 'vue'
import { defineStore } from 'pinia'

import type { TopCategorySummary } from '@/entities/dashboard/model/types'
import { useFinanceStore } from '@/app/stores/finance'
import { useReviewStore } from '@/app/stores/review'
import { useSavingsStore } from '@/app/stores/savings'
import { addDays, endOfCurrentWeek, isDateWithinRange, startOfCurrentWeek, toIsoDate } from '@/shared/lib/date'

type DashboardPeriod = 'week' | 'month'

export const useDashboardStore = defineStore('dashboard', () => {
  const period = ref<DashboardPeriod>('week')
  const dismissedCtas = ref<string[]>([])

  const financeStore = useFinanceStore()
  const reviewStore = useReviewStore()
  const savingsStore = useSavingsStore()

  const range = computed(() => {
    if (period.value === 'week') {
      return {
        start: toIsoDate(startOfCurrentWeek()),
        end: toIsoDate(endOfCurrentWeek()),
      }
    }

    return {
      start: toIsoDate(addDays(new Date(), -29)),
      end: toIsoDate(new Date()),
    }
  })

  const periodTransactions = computed(() =>
    financeStore.transactions.filter((transaction) =>
      isDateWithinRange(transaction.occurredAt, range.value.start, range.value.end),
    ),
  )

  const inflowMinor = computed(() =>
    periodTransactions.value
      .filter((transaction) => transaction.kind === 'income')
      .reduce((sum, transaction) => sum + transaction.amountMinor, 0),
  )

  const outflowMinor = computed(() =>
    periodTransactions.value
      .filter((transaction) => transaction.kind === 'expense')
      .reduce((sum, transaction) => sum + transaction.amountMinor, 0),
  )

  const safeToSpendMinor = computed(() => {
    const reserveMinor = 45_000
    return Math.max(financeStore.totalBalanceMinor - reserveMinor, 0)
  })

  const topCategories = computed<TopCategorySummary[]>(() => {
    const groups = new Map<string, TopCategorySummary>()

    periodTransactions.value
      .filter((transaction) => transaction.kind === 'expense')
      .forEach((transaction) => {
        const category = financeStore.categories.find((item) => item.id === transaction.categoryId)

        if (!category) {
          return
        }

        const current = groups.get(category.id)

        if (current) {
          current.amountMinor += transaction.amountMinor
          return
        }

        groups.set(category.id, {
          categoryId: category.id,
          label: category.name,
          amountMinor: transaction.amountMinor,
        })
      })

    return [...groups.values()]
      .sort((left, right) => right.amountMinor - left.amountMinor)
      .slice(0, 4)
  })

  const reviewState = computed(() => reviewStore.review.status)
  const goalsSummary = computed(() => ({
    total: savingsStore.goals.length,
    active: savingsStore.goals.filter((goal) => !goal.isCompleted).length,
  }))

  function setPeriod(nextPeriod: DashboardPeriod) {
    period.value = nextPeriod
  }

  function dismissCta(id: string) {
    if (!dismissedCtas.value.includes(id)) {
      dismissedCtas.value = [...dismissedCtas.value, id]
    }
  }

  return {
    dismissCta,
    dismissedCtas,
    goalsSummary,
    inflowMinor,
    outflowMinor,
    period,
    reviewState,
    safeToSpendMinor,
    setPeriod,
    topCategories,
  }
})
