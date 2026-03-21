import type { ApiAuthResponse, ApiMeResponse } from '@/shared/api/contracts'
import { apiRequest } from '@/shared/api/http/client'

export function loginWithTelegram(providerUserId: string) {
  return apiRequest<ApiAuthResponse>('/auth/telegram', {
    method: 'POST',
    auth: false,
    body: JSON.stringify({
      provider_user_id: providerUserId,
      username: 'localdev',
      first_name: 'Local',
      last_name: 'User',
      auth_date: Math.floor(Date.now() / 1000),
      hash: 'dev-mode',
    }),
  })
}

export function loginWithYandex(providerUserId: string) {
  return apiRequest<ApiAuthResponse>('/auth/yandex', {
    method: 'POST',
    auth: false,
    body: JSON.stringify({
      provider_user_id: providerUserId,
      email: `${providerUserId}@local.dev`,
      display_name: 'Local Dev',
      avatar_url: null,
    }),
  })
}

export function fetchMe() {
  return apiRequest<ApiMeResponse>('/auth/me')
}

export function logout(refreshToken: string) {
  return apiRequest<void>('/auth/logout', {
    method: 'POST',
    auth: false,
    retryOnUnauthorized: false,
    body: JSON.stringify({
      refresh_token: refreshToken,
    }),
  })
}
