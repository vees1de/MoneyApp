import type { ApiFinanceDashboard } from '@/shared/api/contracts'
import { apiRequest } from '@/shared/api/http/client'

export function fetchFinanceDashboard() {
  return apiRequest<ApiFinanceDashboard>('/dashboard/finance')
}
