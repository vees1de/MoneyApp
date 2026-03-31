import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';

import { API_BASE_URL } from '@core/config/api.config';
import type { Enrollment } from '@entities/enrollment';

@Injectable({ providedIn: 'root' })
export class EnrollmentsApiService {
  private readonly base = `${API_BASE_URL}/v1/enrollments`;

  constructor(private readonly http: HttpClient) {}

  listMy(): Observable<Enrollment[]> {
    return this.http.get<Enrollment[]>(`${this.base}/my`);
  }

  getById(id: string): Observable<Enrollment> {
    return this.http.get<Enrollment>(`${this.base}/${id}`);
  }

  start(id: string): Observable<Enrollment> {
    return this.http.post<Enrollment>(`${this.base}/${id}/start`, {});
  }

  complete(id: string, payload: Record<string, unknown>): Observable<Enrollment> {
    return this.http.post<Enrollment>(`${this.base}/${id}/complete`, payload);
  }

  updateProgress(id: string, payload: Record<string, unknown>): Observable<Enrollment> {
    return this.http.post<Enrollment>(`${this.base}/${id}/progress`, payload);
  }
}