import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';

import { API_BASE_URL } from '@core/config/api.config';
import type { BoardSummary } from './contracts';

@Injectable({ providedIn: 'root' })
export class IntegrationsApiService {
  private readonly base = `${API_BASE_URL}/v1/integrations`;

  constructor(private readonly http: HttpClient) {}

  getOutlookConnectLink(): Observable<Record<string, unknown>> {
    return this.http.get<Record<string, unknown>>(`${this.base}/outlook/connect`);
  }

  getOutlookStatus(): Observable<Record<string, unknown>> {
    return this.http.get<Record<string, unknown>>(`${this.base}/outlook/status`);
  }

  syncOutlook(): Observable<Record<string, unknown>> {
    return this.http.post<Record<string, unknown>>(`${this.base}/outlook/sync`, {});
  }

  disconnectOutlook(): Observable<Record<string, unknown>> {
    return this.http.post<Record<string, unknown>>(`${this.base}/outlook/disconnect`, {});
  }

  getJiraBoardSummary(query?: {
    connection_id?: string;
    board_id?: string;
  }): Observable<BoardSummary> {
    const params = new URLSearchParams();
    if (query?.connection_id) params.set('connection_id', query.connection_id);
    if (query?.board_id) params.set('board_id', query.board_id);
    const suffix = params.toString() ? `?${params.toString()}` : '';
    return this.http.get<BoardSummary>(`${API_BASE_URL}/v1/jira/board-summary${suffix}`);
  }
}
