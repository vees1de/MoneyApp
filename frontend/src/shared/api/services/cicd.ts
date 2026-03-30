import type { ApiCicdSmokeResponse } from '@/shared/api/contracts'
import { apiRequest } from '@/shared/api/http/client'

export function runCicdSmoke() {
  return apiRequest<ApiCicdSmokeResponse>('/cicd/smoke', {
    method: 'POST',
  })
}
