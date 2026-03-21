import { env } from '@/shared/config/env'
import { clearSession, readSession, writeSession } from '@/shared/api/session'
import type { ApiAuthResponse, ApiErrorPayload } from '@/shared/api/contracts'
import { translate } from '@/shared/i18n'

export interface ApiError {
  code: string
  details?: unknown
  message: string
  status: number
}

export interface RequestOptions extends RequestInit {
  auth?: boolean
  query?: Record<string, string | number | boolean | undefined | null>
  retryOnUnauthorized?: boolean
}

let refreshRequest: Promise<string | null> | null = null

function buildUrl(path: string, query?: RequestOptions['query']) {
  const normalizedPath = path.startsWith('http') ? path : `${env.apiBaseUrl}${path}`
  const url = new URL(normalizedPath, window.location.origin)

  Object.entries(query ?? {}).forEach(([key, value]) => {
    if (value === undefined || value === null || value === '') {
      return
    }

    url.searchParams.set(key, String(value))
  })

  return url.toString()
}

async function parseError(response: Response): Promise<ApiError> {
  let payload: ApiErrorPayload | null = null

  try {
    payload = (await response.json()) as ApiErrorPayload
  } catch {
    payload = null
  }

  return {
    code: payload?.error.code ?? 'http_error',
    details: payload?.error.details,
    message: payload?.error.message ?? translate('errors.requestFailed'),
    status: response.status,
  }
}

async function refreshAccessToken() {
  const session = readSession()

  if (!session?.refreshToken) {
    return null
  }

  if (!refreshRequest) {
    refreshRequest = fetch(buildUrl('/auth/refresh'), {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        refresh_token: session.refreshToken,
      }),
    })
      .then(async (response) => {
        if (!response.ok) {
          clearSession()
          return null
        }

        const payload = (await response.json()) as ApiAuthResponse
        const expiresAt = new Date(Date.now() + payload.tokens.expires_in * 1000).toISOString()

        writeSession({
          accessToken: payload.tokens.access_token,
          refreshToken: payload.tokens.refresh_token,
          expiresAt,
          provider: session.provider,
        })

        return payload.tokens.access_token
      })
      .finally(() => {
        refreshRequest = null
      })
  }

  return refreshRequest
}

export async function apiRequest<T>(path: string, options: RequestOptions = {}) {
  const session = readSession()
  const headers = new Headers(options.headers ?? {})
  const needsJsonBody = options.body !== undefined && !headers.has('Content-Type')

  if (needsJsonBody) {
    headers.set('Content-Type', 'application/json')
  }

  if (options.auth !== false && session?.accessToken) {
    headers.set('Authorization', `Bearer ${session.accessToken}`)
  }

  let response = await fetch(buildUrl(path, options.query), {
    ...options,
    headers,
  })

  if (response.status === 401 && options.auth !== false && options.retryOnUnauthorized !== false) {
    const nextAccessToken = await refreshAccessToken()

    if (nextAccessToken) {
      headers.set('Authorization', `Bearer ${nextAccessToken}`)

      response = await fetch(buildUrl(path, options.query), {
        ...options,
        headers,
      })
    }
  }

  if (!response.ok) {
    throw await parseError(response)
  }

  if (response.status === 204) {
    return undefined as T
  }

  return (await response.json()) as T
}
