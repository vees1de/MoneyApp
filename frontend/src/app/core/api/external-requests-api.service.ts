import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';

import { API_BASE_URL } from '@core/config/api.config';
import type { ListQuery } from './api.types';
import type { ExternalRequest, PendingApprovalItem } from './contracts';
import { toHttpParams } from './http-params.util';

@Injectable({ providedIn: 'root' })
export class ExternalRequestsApiService {
  private readonly base = `${API_BASE_URL}/v1/external-requests`;

  constructor(private readonly http: HttpClient) {}

  list(query?: ListQuery): Observable<ExternalRequest[]> {
    return this.http.get<ExternalRequest[]>(this.base, { params: toHttpParams(query) });
  }

  listMy(): Observable<ExternalRequest[]> {
    return this.http.get<ExternalRequest[]>(`${this.base}/my`);
  }

  listPendingApprovals(): Observable<PendingApprovalItem[]> {
    return this.http.get<PendingApprovalItem[]>(`${this.base}/pending-approvals`);
  }

  getById(id: string): Observable<ExternalRequest> {
    return this.http.get<ExternalRequest>(`${this.base}/${id}`);
  }

  create(payload: Record<string, unknown>): Observable<ExternalRequest> {
    return this.http.post<ExternalRequest>(this.base, payload);
  }

  update(id: string, payload: Record<string, unknown>): Observable<ExternalRequest> {
    return this.http.patch<ExternalRequest>(`${this.base}/${id}`, payload);
  }

  submit(id: string): Observable<ExternalRequest> {
    return this.http.post<ExternalRequest>(`${this.base}/${id}/submit`, {});
  }

  approve(id: string, payload: Record<string, unknown>): Observable<ExternalRequest> {
    return this.http.post<ExternalRequest>(`${this.base}/${id}/approve`, payload);
  }

  reject(id: string, payload: Record<string, unknown>): Observable<ExternalRequest> {
    return this.http.post<ExternalRequest>(`${this.base}/${id}/reject`, payload);
  }

  requestRevision(id: string, payload: Record<string, unknown>): Observable<ExternalRequest> {
    return this.http.post<ExternalRequest>(`${this.base}/${id}/request-revision`, payload);
  }

  uploadCertificate(id: string, payload: Record<string, unknown>): Observable<ExternalRequest> {
    return this.http.post<ExternalRequest>(`${this.base}/${id}/upload-certificate`, payload);
  }
}
