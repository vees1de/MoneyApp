import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable, map } from 'rxjs';

import { API_BASE_URL } from '@core/config/api.config';
import type { ExternalRequest, PendingApprovalItem } from './contracts';
import type { ListQuery, ListResponse } from './api.types';
import { toHttpParams } from './http-params.util';

@Injectable({ providedIn: 'root' })
export class ExternalRequestsApiService {
  private readonly base = `${API_BASE_URL}/v1/external-requests`;

  constructor(private readonly http: HttpClient) {}

  list(query?: ListQuery): Observable<ExternalRequest[]> {
    return this.http
      .get<ListResponse<ExternalRequest>>(this.base, { params: toHttpParams(query) })
      .pipe(map((response) => response.items ?? []));
  }

  listFiltered(options: {
    scope?: 'my' | 'team' | 'all';
    statuses?: string[];
    assignee?: string;
  }): Observable<ExternalRequest[]> {
    let params = new HttpParams();

    if (options.scope) {
      params = params.set('scope', options.scope);
    }
    if (options.assignee) {
      params = params.set('assignee', options.assignee);
    }
    (options.statuses ?? []).forEach((status) => {
      params = params.append('status', status);
    });

    return this.http.get<ExternalRequest[]>(this.base, { params });
  }

  listMy(): Observable<ExternalRequest[]> {
    return this.http
      .get<ListResponse<ExternalRequest>>(`${this.base}/my`)
      .pipe(map((response) => response.items ?? []));
  }

  listByScopeAndStatuses(
    scope: 'my' | 'team' | 'all',
    statuses: string[],
  ): Observable<ExternalRequest[]> {
    let params = new HttpParams().set('scope', scope);
    statuses.forEach((status) => {
      params = params.append('status', status);
    });
    return this.http.get<ExternalRequest[]>(this.base, { params });
  }

  listPendingApprovals(query?: ListQuery): Observable<PendingApprovalItem[]> {
    return this.http
      .get<ListResponse<PendingApprovalItem>>(`${this.base}/pending-approvals`, {
        params: toHttpParams(query),
      })
      .pipe(map((response) => response.items ?? []));
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
