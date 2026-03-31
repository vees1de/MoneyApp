import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';

import { API_BASE_URL } from '@core/config/api.config';

@Injectable({ providedIn: 'root' })
export class AnalyticsApiService {
  private readonly base = `${API_BASE_URL}/v1/analytics`;

  constructor(private readonly http: HttpClient) {}

  getHrDashboard(): Observable<Record<string, unknown>> {
    return this.http.get<Record<string, unknown>>(`${this.base}/dashboard/hr`);
  }

  getManagerDashboard(): Observable<Record<string, unknown>> {
    return this.http.get<Record<string, unknown>>(`${this.base}/dashboard/manager`);
  }

  getBudget(): Observable<Record<string, unknown>> {
    return this.http.get<Record<string, unknown>>(`${this.base}/budget`);
  }

  getCompliance(): Observable<Record<string, unknown>> {
    return this.http.get<Record<string, unknown>>(`${this.base}/compliance`);
  }

  getExternalRequestsAnalytics(): Observable<Record<string, unknown>> {
    return this.http.get<Record<string, unknown>>(`${this.base}/external-requests`);
  }

  getTrainersAnalytics(): Observable<Record<string, unknown>> {
    return this.http.get<Record<string, unknown>>(`${this.base}/trainers`);
  }
}