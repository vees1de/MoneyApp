import type { ListResponse } from './api.types';

export function unwrapListResponse<T>(response: T[] | ListResponse<T>): T[] {
  if (Array.isArray(response)) {
    return response;
  }

  return response.items ?? [];
}
