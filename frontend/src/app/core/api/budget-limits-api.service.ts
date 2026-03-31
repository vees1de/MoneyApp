import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';

import { API_BASE_URL } from '@core/config/api.config';
import type { BudgetLimit } from '@entities/budget-limit';

@Injectable({ providedIn: 'root' })
export class BudgetLimitsApiService {
  private readonly base = `${API_BASE_URL}/v1/budget-limits`;

  constructor(private readonly http: HttpClient) {}

  list(): Observable<BudgetLimit[]> {
    return this.http.get<BudgetLimit[]>(this.base);
  }

  create(payload: Record<string, unknown>): Observable<BudgetLimit> {
    return this.http.post<BudgetLimit>(this.base, payload);
  }
}