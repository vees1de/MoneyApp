import type { ApiWeeklyReview } from '@/shared/api/contracts'
import { apiRequest } from '@/shared/api/http/client'

export function fetchCurrentWeeklyReview() {
  return apiRequest<ApiWeeklyReview>('/reviews/weekly/current')
}

export function submitWeeklyReviewBalance(reviewId: string, actualBalance: string) {
  return apiRequest<ApiWeeklyReview>(`/reviews/weekly/${reviewId}/submit-balance`, {
    method: 'POST',
    body: JSON.stringify({
      actual_balance: actualBalance,
    }),
  })
}

export function resolveWeeklyReview(reviewId: string, resolutionNote?: string) {
  return apiRequest<ApiWeeklyReview>(`/reviews/weekly/${reviewId}/resolve`, {
    method: 'POST',
    body: JSON.stringify({
      resolution_note: resolutionNote ?? null,
    }),
  })
}

export function skipWeeklyReview(reviewId: string) {
  return apiRequest<ApiWeeklyReview>(`/reviews/weekly/${reviewId}/skip`, {
    method: 'POST',
  })
}
