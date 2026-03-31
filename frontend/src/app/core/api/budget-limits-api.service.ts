import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable, map } from 'rxjs';

import { API_BASE_URL } from '@core/config/api.config';
import type { ListResponse } from './api.types';
import type { BudgetLimitRecord } from './contracts';

@Injectable({ providedIn: 'root' })
export class BudgetLimitsApiService {
  private readonly base = `${API_BASE_URL}/v1/budget-limits`;

  constructor(private readonly http: HttpClient) {}

  list(): Observable<BudgetLimitRecord[]> {
    return this.http
      .get<ListResponse<BudgetLimitRecord>>(this.base)
      .pipe(map((response) => response.items ?? []));
  }

  create(payload: Record<string, unknown>): Observable<BudgetLimitRecord> {
    return this.http.post<BudgetLimitRecord>(this.base, payload);
  }
}
