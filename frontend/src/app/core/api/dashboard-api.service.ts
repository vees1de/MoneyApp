import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';

import { API_BASE_URL } from '@core/config/api.config';

@Injectable({ providedIn: 'root' })
export class DashboardApiService {
  private readonly base = `${API_BASE_URL}/v1/dashboard`;

  constructor(private readonly http: HttpClient) {}

  getEmployeeDashboard(): Observable<Record<string, unknown>> {
    return this.http.get<Record<string, unknown>>(`${this.base}/employee`);
  }

  getManagerDashboard(): Observable<Record<string, unknown>> {
    return this.http.get<Record<string, unknown>>(`${this.base}/manager`);
  }
}