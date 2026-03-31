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

    const values = Array.isArray(value) ? value : [value];
    values.forEach((item) => {
      params = params.append(key, String(item));
    });
  });
  return params;
}
