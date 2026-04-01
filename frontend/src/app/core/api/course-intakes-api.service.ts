import { Injectable } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { Observable, map } from 'rxjs';

import { API_BASE_URL } from '@core/config/api.config';
import type { CourseApplication, CourseIntake } from './contracts';
import { unwrapListResponse } from './list-response.util';

@Injectable({ providedIn: 'root' })
export class CourseIntakesApiService {
  private readonly base = `${API_BASE_URL}/v1/intakes`;

  constructor(private readonly http: HttpClient) {}

  list(status?: string): Observable<CourseIntake[]> {
    let params = new HttpParams();
    if (status) {
      params = params.set('status', status);
    }

    return this.http
      .get<CourseIntake[] | { items: CourseIntake[] }>(this.base, { params })
      .pipe(map((response) => unwrapListResponse(response)));
  }

  getById(id: string): Observable<CourseIntake> {
    return this.http.get<CourseIntake>(`${this.base}/${id}`);
  }

  create(payload: Record<string, unknown>): Observable<CourseIntake> {
    return this.http.post<CourseIntake>(this.base, payload);
  }

  update(id: string, payload: Record<string, unknown>): Observable<CourseIntake> {
    return this.http.patch<CourseIntake>(`${this.base}/${id}`, payload);
  }

  delete(id: string): Observable<void> {
    return this.http.delete<void>(`${this.base}/${id}`);
  }

  close(id: string): Observable<CourseIntake> {
    return this.http.post<CourseIntake>(`${this.base}/${id}/close`, {});
  }

  listApplications(intakeId: string): Observable<CourseApplication[]> {
    return this.http
      .get<
        CourseApplication[] | { items: CourseApplication[] }
      >(`${this.base}/${intakeId}/applications`)
      .pipe(map((response) => unwrapListResponse(response)));
  }

  updatePaymentStatus(id: string, status: 'paid' | 'unpaid'): Observable<CourseApplication[]> {
    return this.http
      .post<CourseApplication[] | { items: CourseApplication[] }>(
        `${this.base}/${id}/payment-status`,
        {
          status,
        },
      )
      .pipe(map((response) => unwrapListResponse(response)));
  }

  startCourse(id: string): Observable<CourseApplication[]> {
    return this.http
      .post<
        CourseApplication[] | { items: CourseApplication[] }
      >(`${this.base}/${id}/start-course`, {})
      .pipe(map((response) => unwrapListResponse(response)));
  }
}
