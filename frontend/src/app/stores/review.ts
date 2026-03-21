import { computed, ref } from 'vue'
import { defineStore } from 'pinia'

import type { WeeklyReview } from '@/entities/review/model/types'
import { mapApiWeeklyReview } from '@/shared/api/mappers'
import {
  fetchCurrentWeeklyReview,
  resolveWeeklyReview,
  skipWeeklyReview,
  submitWeeklyReviewBalance,
} from '@/shared/api/services/review'
import { translate } from '@/shared/i18n'
import { minorToMoneyString } from '@/shared/lib/money'

function createEmptyReview(): WeeklyReview {
  return {
    id: '',
    periodStart: '',
    periodEnd: '',
    expectedBalanceMinor: 0,
    actualBalanceMinor: null,
    deltaMinor: null,
    resolutionNote: null,
    status: 'pending',
    resolvedAt: null,
  }
}

export const useReviewStore = defineStore('review', () => {
  const review = ref<WeeklyReview>(createEmptyReview())
  const loading = ref(false)

  const deltaHint = computed(() => {
    if (!review.value.id) {
      return translate('review.hintNoFetch')
    }

    if (review.value.deltaMinor === null) {
      return translate('review.hintEnterBalance')
    }

    if (review.value.deltaMinor === 0) {
      return translate('review.hintMatched')
    }

    if (review.value.deltaMinor < 0) {
      return translate('review.hintNegative')
    }

    return translate('review.hintPositive')
  })

  async function fetchCurrent() {
    loading.value = true
    try {
      const response = await fetchCurrentWeeklyReview()
      review.value = mapApiWeeklyReview(response)
      return review.value
    } finally {
      loading.value = false
    }
  }

  async function submitActualBalance(valueMinor: number) {
    const response = await submitWeeklyReviewBalance(review.value.id, minorToMoneyString(valueMinor))
    review.value = mapApiWeeklyReview(response)
    return review.value
  }

  async function resolve(reason = 'Resolved from frontend') {
    const response = await resolveWeeklyReview(review.value.id, reason)
    review.value = mapApiWeeklyReview(response)
    return review.value
  }

  async function skip() {
    const response = await skipWeeklyReview(review.value.id)
    review.value = mapApiWeeklyReview(response)
    return review.value
  }

  function reset() {
    review.value = createEmptyReview()
  }

  return {
    deltaHint,
    fetchCurrent,
    loading,
    reset,
    resolve,
    review,
    skip,
    submitActualBalance,
  }
})
