import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';

import { API_BASE_URL } from '@core/config/api.config';

@Injectable({ providedIn: 'root' })
export class TestsApiService {
  private readonly base = `${API_BASE_URL}/v1/tests`;

  constructor(private readonly http: HttpClient) {}

  getById(id: string): Observable<Record<string, unknown>> {
    return this.http.get<Record<string, unknown>>(`${this.base}/${id}`);
  }

  getResults(id: string): Observable<Record<string, unknown>[]> {
    return this.http.get<Record<string, unknown>[]>(`${this.base}/${id}/results`);
  }

  startAttempt(testId: string, payload: Record<string, unknown>): Observable<Record<string, unknown>> {
    return this.http.post<Record<string, unknown>>(`${this.base}/${testId}/attempts`, payload);
  }

  submitAnswers(attemptId: string, payload: Record<string, unknown>): Observable<Record<string, unknown>> {
    return this.http.post<Record<string, unknown>>(`${API_BASE_URL}/v1/test-attempts/${attemptId}/answers`, payload);
  }

  submitAttempt(attemptId: string): Observable<Record<string, unknown>> {
    return this.http.post<Record<string, unknown>>(`${API_BASE_URL}/v1/test-attempts/${attemptId}/submit`, {});
  }
}