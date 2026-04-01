import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';

import { API_BASE_URL } from '@core/config/api.config';
import type { SmartExportRequest, SourcesResponse } from '@entities/smart-export';

@Injectable({ providedIn: 'root' })
export class ReportsApiService {
  private readonly base = `${API_BASE_URL}/v1/reports`;

  constructor(private readonly http: HttpClient) {}

  getSources(): Observable<SourcesResponse> {
    return this.http.get<SourcesResponse>(`${this.base}/sources`);
  }

  smartExport(request: SmartExportRequest): Observable<Blob> {
    return this.http.post(`${this.base}/smart-export`, request, {
      responseType: 'blob',
    });
  }

  exportExcel(): Observable<Blob> {
    return this.http.get(`${this.base}/export/excel`, { responseType: 'blob' });
  }

  exportPdf(): Observable<Blob> {
    return this.http.get(`${this.base}/export/pdf`, { responseType: 'blob' });
  }
}
