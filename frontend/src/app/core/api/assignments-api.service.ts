import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable, map } from 'rxjs';

import { API_BASE_URL } from '@core/config/api.config';
import type { ListResponse } from './api.types';
import type { CourseAssignment } from './contracts';

@Injectable({ providedIn: 'root' })
export class AssignmentsApiService {
  private readonly base = `${API_BASE_URL}/v1/assignments`;

  constructor(private readonly http: HttpClient) {}

  list(): Observable<CourseAssignment[]> {
    return this.http
      .get<ListResponse<CourseAssignment>>(this.base)
      .pipe(map((response) => response.items ?? []));
  }

  create(payload: Record<string, unknown>): Observable<CourseAssignment> {
    return this.http.post<CourseAssignment>(this.base, payload);
  }
}
