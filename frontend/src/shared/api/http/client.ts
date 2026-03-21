import { env } from '@/shared/config/env'

export interface ApiError {
  code: string
  message: string
  status: number
}

export interface RequestOptions extends RequestInit {
  query?: Record<string, string | number | boolean | undefined | null>
}

function buildUrl(path: string, query?: RequestOptions['query']) {
  const baseUrl = path.startsWith('http') ? path : `${env.apiBaseUrl}${path}`
  const url = new URL(baseUrl, window.location.origin)

  Object.entries(query ?? {}).forEach(([key, value]) => {
    if (value === undefined || value === null || value === '') {
      return
    }

    url.searchParams.set(key, String(value))
  })

  return url.toString()
}

export async function apiRequest<T>(path: string, options: RequestOptions = {}) {
  const response = await fetch(buildUrl(path, options.query), {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...(options.headers ?? {}),
    },
  })

  if (!response.ok) {
    const error: ApiError = {
      code: 'http_error',
      message: 'Request failed',
      status: response.status,
    }

    throw error
  }

  if (response.status === 204) {
    return undefined as T
  }

  return (await response.json()) as T
}
