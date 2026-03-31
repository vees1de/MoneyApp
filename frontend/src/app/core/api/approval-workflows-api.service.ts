import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable, map } from 'rxjs';

import { API_BASE_URL } from '@core/config/api.config';
import type { ListResponse } from './api.types';
import type { ApprovalWorkflow } from './contracts';

@Injectable({ providedIn: 'root' })
export class ApprovalWorkflowsApiService {
  private readonly base = `${API_BASE_URL}/v1/approval-workflows`;

  constructor(private readonly http: HttpClient) {}

  list(): Observable<ApprovalWorkflow[]> {
    return this.http
      .get<ListResponse<ApprovalWorkflow>>(this.base)
      .pipe(map((response) => response.items ?? []));
  }

  create(payload: Record<string, unknown>): Observable<ApprovalWorkflow> {
    return this.http.post<ApprovalWorkflow>(this.base, payload);
  }
}
