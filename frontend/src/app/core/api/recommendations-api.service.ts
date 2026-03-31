import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';

import { API_BASE_URL } from '@core/config/api.config';
import type { Course } from '@entities/course';

@Injectable({ providedIn: 'root' })
export class RecommendationsApiService {
  private readonly base = `${API_BASE_URL}/v1/recommendations/courses`;

  constructor(private readonly http: HttpClient) {}

  list(query?: { limit?: number; offset?: number }): Observable<Course[]> {
    const params = new URLSearchParams();
    if (typeof query?.limit === 'number') params.set('limit', String(query.limit));
    if (typeof query?.offset === 'number') params.set('offset', String(query.offset));
    const suffix = params.toString() ? `?${params.toString()}` : '';
    return this.http.get<Course[]>(`${this.base}${suffix}`);
  }
}