import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';

import { API_BASE_URL } from '@core/config/api.config';
import type { MyLearningPlan } from './contracts';

@Injectable({ providedIn: 'root' })
export class LearningPlanApiService {
  private readonly base = `${API_BASE_URL}/v1/learning-plan`;

  constructor(private readonly http: HttpClient) {}

  getMyPlan(): Observable<MyLearningPlan> {
    return this.http.get<MyLearningPlan>(`${this.base}/my`);
  }
}
