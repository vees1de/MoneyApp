import type { ApiMeResponse } from '@/shared/api/contracts'
import { apiRequest } from '@/shared/api/http/client'

export function updateUserPreferences(input: {
  baseCurrency?: string
  onboardingCompleted?: boolean
  timezone?: string
  weeklyReviewHour?: number
  weeklyReviewWeekday?: number
}) {
  return apiRequest<ApiMeResponse>('/users/preferences', {
    method: 'PATCH',
    body: JSON.stringify({
      timezone: input.timezone,
      base_currency: input.baseCurrency,
      onboarding_completed: input.onboardingCompleted,
      weekly_review_hour: input.weeklyReviewHour,
      weekly_review_weekday: input.weeklyReviewWeekday,
    }),
  })
}
