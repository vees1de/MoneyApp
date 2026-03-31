import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';

import { API_BASE_URL } from '@core/config/api.config';
import type { LearningDirection } from '@entities/learning-direction';

@Injectable({ providedIn: 'root' })
export class ProgramsApiService {
  private readonly base = `${API_BASE_URL}/v1/programs`;

  constructor(private readonly http: HttpClient) {}

  list(): Observable<LearningDirection[]> {
    return this.http.get<LearningDirection[]>(this.base);
  }

  getById(id: string): Observable<LearningDirection> {
    return this.http.get<LearningDirection>(`${this.base}/${id}`);
  }

  create(payload: Record<string, unknown>): Observable<LearningDirection> {
    return this.http.post<LearningDirection>(this.base, payload);
  }

  createGroup(programId: string, payload: Record<string, unknown>): Observable<Record<string, unknown>> {
    return this.http.post<Record<string, unknown>>(`${this.base}/${programId}/groups`, payload);
  }
}