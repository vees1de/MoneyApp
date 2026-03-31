import { Injectable } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { Observable, map } from 'rxjs';

import { API_BASE_URL } from '@core/config/api.config';
import type { CourseSuggestion, SuggestionOpenIntakeResponse } from './contracts';
import { unwrapListResponse } from './list-response.util';

@Injectable({ providedIn: 'root' })
export class CourseSuggestionsApiService {
  private readonly base = `${API_BASE_URL}/v1/suggestions`;

  constructor(private readonly http: HttpClient) {}

  list(status?: string): Observable<CourseSuggestion[]> {
    let params = new HttpParams();
    if (status) {
      params = params.set('status', status);
    }

    return this.http
      .get<CourseSuggestion[] | { items: CourseSuggestion[] }>(this.base, { params })
      .pipe(map((response) => unwrapListResponse(response)));
  }

  listMy(): Observable<CourseSuggestion[]> {
    return this.http
      .get<CourseSuggestion[] | { items: CourseSuggestion[] }>(`${this.base}/my`)
      .pipe(map((response) => unwrapListResponse(response)));
  }

  getById(id: string): Observable<CourseSuggestion> {
    return this.http.get<CourseSuggestion>(`${this.base}/${id}`);
  }

  create(payload: Record<string, unknown>): Observable<CourseSuggestion> {
    return this.http.post<CourseSuggestion>(this.base, payload);
  }

  approve(id: string, payload: Record<string, unknown>): Observable<CourseSuggestion> {
    return this.http.post<CourseSuggestion>(`${this.base}/${id}/approve`, payload);
  }

  reject(id: string, payload: Record<string, unknown>): Observable<CourseSuggestion> {
    return this.http.post<CourseSuggestion>(`${this.base}/${id}/reject`, payload);
  }

  openIntake(
    id: string,
    payload: Record<string, unknown>,
  ): Observable<SuggestionOpenIntakeResponse> {
    return this.http.post<SuggestionOpenIntakeResponse>(`${this.base}/${id}/open-intake`, payload);
  }
}
