import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';

import { API_BASE_URL } from '@core/config/api.config';

@Injectable({ providedIn: 'root' })
export class ReportsApiService {
  private readonly base = `${API_BASE_URL}/v1/reports/export`;

  constructor(private readonly http: HttpClient) {}

  exportExcel(): Observable<Blob> {
    return this.http.get(`${this.base}/excel`, { responseType: 'blob' });
  }

  exportPdf(): Observable<Blob> {
    return this.http.get(`${this.base}/pdf`, { responseType: 'blob' });
  }
}