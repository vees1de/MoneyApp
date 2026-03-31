import { HttpParams } from '@angular/common/http';

import type { ListQuery } from './api.types';

export function toHttpParams(query?: ListQuery): HttpParams {
  if (!query) {
    return new HttpParams();
  }

  let params = new HttpParams();
  Object.entries(query).forEach(([key, value]) => {
    if (value === undefined || value === null) {
      return;
    }
    params = params.set(key, String(value));
  });
  return params;
}