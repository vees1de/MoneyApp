import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';

import { API_BASE_URL } from '@core/config/api.config';
import type { Request } from '@entities/request';
import type { ListQuery } from './api.types';
import { toHttpParams } from './http-params.util';

@Injectable({ providedIn: 'root' })
export class ExternalRequestsApiService {
  private readonly base = `${API_BASE_URL}/v1/external-requests`;

  constructor(private readonly http: HttpClient) {}

  list(query?: ListQuery): Observable<Request[]> {
    return this.http.get<Request[]>(this.base, { params: toHttpParams(query) });
  }

  listMy(): Observable<Request[]> {
    return this.http.get<Request[]>(`${this.base}/my`);
  }

  listPendingApprovals(): Observable<Record<string, unknown>[]> {
    return this.http.get<Record<string, unknown>[]>(`${this.base}/pending-approvals`);
  }

  getById(id: string): Observable<Request> {
    return this.http.get<Request>(`${this.base}/${id}`);
  }

  create(payload: Record<string, unknown>): Observable<Request> {
    return this.http.post<Request>(this.base, payload);
  }

  update(id: string, payload: Record<string, unknown>): Observable<Request> {
    return this.http.patch<Request>(`${this.base}/${id}`, payload);
  }

  submit(id: string): Observable<Request> {
    return this.http.post<Request>(`${this.base}/${id}/submit`, {});
  }

  approve(id: string, payload: Record<string, unknown>): Observable<Request> {
    return this.http.post<Request>(`${this.base}/${id}/approve`, payload);
  }

  reject(id: string, payload: Record<string, unknown>): Observable<Request> {
    return this.http.post<Request>(`${this.base}/${id}/reject`, payload);
  }

  requestRevision(id: string, payload: Record<string, unknown>): Observable<Request> {
    return this.http.post<Request>(`${this.base}/${id}/request-revision`, payload);
  }

  uploadCertificate(id: string, payload: Record<string, unknown>): Observable<Request> {
    return this.http.post<Request>(`${this.base}/${id}/upload-certificate`, payload);
  }
}