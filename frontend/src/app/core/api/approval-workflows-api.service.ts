import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';

import { API_BASE_URL } from '@core/config/api.config';
import type { ApprovalStep } from '@entities/approval-step';

@Injectable({ providedIn: 'root' })
export class ApprovalWorkflowsApiService {
  private readonly base = `${API_BASE_URL}/v1/approval-workflows`;

  constructor(private readonly http: HttpClient) {}

  list(): Observable<ApprovalStep[]> {
    return this.http.get<ApprovalStep[]>(this.base);
  }

  create(payload: Record<string, unknown>): Observable<ApprovalStep> {
    return this.http.post<ApprovalStep>(this.base, payload);
  }
}