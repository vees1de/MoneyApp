import type {
  ApiSavingsGoal,
  ApiSavingsGoalsResponse,
  ApiSavingsSummary,
} from '@/shared/api/contracts'
import { apiRequest } from '@/shared/api/http/client'

export function listSavingsGoals() {
  return apiRequest<ApiSavingsGoalsResponse>('/savings/goals')
}

export function createSavingsGoal(input: {
  currency: string
  currentAmount: string
  priority: 'low' | 'medium' | 'high'
  targetAmount: string
  targetDate: string | null
  title: string
}) {
  return apiRequest<ApiSavingsGoal>('/savings/goals', {
    method: 'POST',
    body: JSON.stringify({
      title: input.title,
      target_amount: input.targetAmount,
      current_amount: input.currentAmount,
      currency: input.currency,
      target_date: input.targetDate,
      priority: input.priority,
    }),
  })
}

export function fetchSavingsSummary() {
  return apiRequest<ApiSavingsSummary>('/savings/summary')
}
