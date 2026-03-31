import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';

import { API_BASE_URL } from '@core/config/api.config';
import type { Enrollment } from '@entities/enrollment';

@Injectable({ providedIn: 'root' })
export class AssignmentsApiService {
  private readonly base = `${API_BASE_URL}/v1/assignments`;

  constructor(private readonly http: HttpClient) {}

  list(): Observable<Enrollment[]> {
    return this.http.get<Enrollment[]>(this.base);
  }

  create(payload: Record<string, unknown>): Observable<Enrollment> {
    return this.http.post<Enrollment>(this.base, payload);
  }
}