import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable, map } from 'rxjs';

import { API_BASE_URL } from '@core/config/api.config';
import type { CourseApplication } from './contracts';
import { unwrapListResponse } from './list-response.util';

@Injectable({ providedIn: 'root' })
export class CourseApplicationsApiService {
  private readonly base = `${API_BASE_URL}/v1/applications`;

  constructor(private readonly http: HttpClient) {}

  apply(payload: Record<string, unknown>): Observable<CourseApplication> {
    return this.http.post<CourseApplication>(this.base, payload);
  }

  listMy(): Observable<CourseApplication[]> {
    return this.http
      .get<CourseApplication[] | { items: CourseApplication[] }>(`${this.base}/my`)
      .pipe(map((response) => unwrapListResponse(response)));
  }

  listPendingManager(): Observable<CourseApplication[]> {
    return this.http
      .get<CourseApplication[] | { items: CourseApplication[] }>(`${this.base}/pending-manager`)
      .pipe(map((response) => unwrapListResponse(response)));
  }

  getById(id: string): Observable<CourseApplication> {
    return this.http.get<CourseApplication>(`${this.base}/${id}`);
  }

  approveManager(id: string, payload: Record<string, unknown>): Observable<CourseApplication> {
    return this.http.post<CourseApplication>(`${this.base}/${id}/approve-manager`, payload);
  }

  rejectManager(id: string, payload: Record<string, unknown>): Observable<CourseApplication> {
    return this.http.post<CourseApplication>(`${this.base}/${id}/reject-manager`, payload);
  }

  approveHr(id: string, payload: Record<string, unknown>): Observable<CourseApplication> {
    return this.http.post<CourseApplication>(`${this.base}/${id}/approve-hr`, payload);
  }

  rejectHr(id: string, payload: Record<string, unknown>): Observable<CourseApplication> {
    return this.http.post<CourseApplication>(`${this.base}/${id}/reject-hr`, payload);
  }

  withdraw(id: string): Observable<CourseApplication> {
    return this.http.post<CourseApplication>(`${this.base}/${id}/withdraw`, {});
  }

  enroll(id: string): Observable<CourseApplication> {
    return this.http.post<CourseApplication>(`${this.base}/${id}/enroll`, {});
  }
}
