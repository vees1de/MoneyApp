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

  exportIntakeApplicationsExcel(intakeId: string): Observable<Blob> {
    const request: SmartExportRequest = {
      source: 'applications',
      format: 'xlsx',
      columns: [
        'employee_name',
        'employee_email',
        'department',
        'position',
        'intake_title',
        'course_title',
        'max_price',
        'current_price',
        'status',
        'payment_status',
        'enrollment_status',
        'certificate_status',
        'certificate_uploaded_at',
        'motivation',
        'manager_approver',
        'manager_comment',
        'manager_decided_at',
        'hr_approver',
        'hr_comment',
        'hr_decided_at',
        'applied_at',
        'updated_at',
      ],
      filters: {
        intake_ids: [intakeId],
      },
      sort_by: 'applied_at',
      sort_dir: 'desc',
    };

    return this.smartExport(request);
  }

  exportExcel(): Observable<Blob> {
    return this.http.get(`${this.base}/export/excel`, { responseType: 'blob' });
  }

  exportPdf(): Observable<Blob> {
    return this.http.get(`${this.base}/export/pdf`, { responseType: 'blob' });
  }
}
