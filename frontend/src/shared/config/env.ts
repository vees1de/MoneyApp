const fallbackApiBaseUrl = '/api/v1'

export const env = {
  appName: 'Personal Life OS',
  productName: 'Finance Core',
  apiBaseUrl: import.meta.env.VITE_API_BASE_URL ?? fallbackApiBaseUrl,
}
