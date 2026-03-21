import { computed, ref, watch } from 'vue'
import { defineStore } from 'pinia'

import type { WeeklyReview } from '@/entities/review/model/types'
import { useFinanceStore } from '@/app/stores/finance'
import { demoWeeklyReview } from '@/shared/mocks/demo'
import { isDateWithinRange } from '@/shared/lib/date'
import { readStorage, writeStorage } from '@/shared/lib/storage'

const REVIEW_STORAGE_KEY = 'plos-review-current'

function cloneReview(): WeeklyReview {
  return { ...demoWeeklyReview }
}

export const useReviewStore = defineStore('review', () => {
  const review = ref<WeeklyReview>(cloneReview())
  const resolutionReason = ref<string | null>(null)
  const hydrated = ref(false)

  function bootstrap() {
    review.value = readStorage<WeeklyReview>(REVIEW_STORAGE_KEY, cloneReview())
    hydrated.value = true
  }

  const financeStore = useFinanceStore()

  const periodTransactions = computed(() =>
    financeStore.transactions.filter((transaction) =>
      isDateWithinRange(transaction.occurredAt, review.value.periodStart, review.value.periodEnd),
    ),
  )

  const incomeMinor = computed(() =>
    periodTransactions.value
      .filter((transaction) => transaction.kind === 'income')
      .reduce((sum, transaction) => sum + transaction.amountMinor, 0),
  )

  const expenseMinor = computed(() =>
    periodTransactions.value
      .filter((transaction) => transaction.kind === 'expense')
      .reduce((sum, transaction) => sum + transaction.amountMinor, 0),
  )

  const expectedBalanceMinor = computed(
    () => review.value.openingBalanceMinor + incomeMinor.value - expenseMinor.value,
  )

  const deltaMinor = computed(() => {
    if (review.value.actualBalanceMinor === null) {
      return null
    }

    return review.value.actualBalanceMinor - expectedBalanceMinor.value
  })

  const deltaHint = computed(() => {
    if (deltaMinor.value === null) {
      return 'Enter the real balance to compare it with recorded transactions.'
    }

    if (deltaMinor.value === 0) {
      return 'Everything matches. You can complete the weekly review.'
    }

    if (deltaMinor.value < 0) {
      return 'Negative delta usually means a missing expense or a wrong account.'
    }

    return 'Positive delta often means a missed income or an expense that never happened.'
  })

  function submitActualBalance(value: number) {
    review.value.actualBalanceMinor = value
    review.value.status = 'pending'
  }

  function resolve(reason = 'confirmed') {
    resolutionReason.value = reason
    review.value.status = 'completed'
    review.value.resolvedAt = new Date().toISOString()
  }

  function skip() {
    review.value.status = 'skipped'
    review.value.resolvedAt = new Date().toISOString()
  }

  function reopen() {
    review.value.status = 'pending'
    review.value.resolvedAt = null
  }

  function reset() {
    review.value = cloneReview()
    resolutionReason.value = null
  }

  watch(
    review,
    (nextReview) => {
      if (hydrated.value) {
        writeStorage(REVIEW_STORAGE_KEY, nextReview)
      }
    },
    { deep: true },
  )

  return {
    bootstrap,
    deltaHint,
    deltaMinor,
    expectedBalanceMinor,
    expenseMinor,
    incomeMinor,
    periodTransactions,
    reopen,
    resolutionReason,
    resolve,
    reset,
    review,
    skip,
    submitActualBalance,
  }
})
