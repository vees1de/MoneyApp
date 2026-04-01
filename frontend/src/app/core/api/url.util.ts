import { API_BASE_URL } from '@core/config/api.config';

const API_ORIGIN = API_BASE_URL.replace(/\/api$/, '');

export function resolveApiUrl(value?: string | null): string | null {
  const trimmed = value?.trim();
  if (!trimmed) {
    return null;
  }

  if (/^[a-zA-Z][a-zA-Z\d+\-.]*:/.test(trimmed) || trimmed.startsWith('//')) {
    return trimmed;
  }

  if (trimmed.startsWith('/api/')) {
    return `${API_ORIGIN}${trimmed}`;
  }

  if (trimmed.startsWith('/')) {
    return `${API_BASE_URL}${trimmed}`;
  }

  return `${API_BASE_URL}/${trimmed}`;
}
