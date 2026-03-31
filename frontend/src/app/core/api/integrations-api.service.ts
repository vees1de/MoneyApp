import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import type { BoardSummary } from './contracts';
import { API_BASE_URL } from '@core/config/api.config';

import type {
  YougileAutoMatchRequest,
  YougileAutoMatchResponse,
  YougileBoard,
  YougileColumn,
  YougileConnection,
  YougileCreateConnectionRequest,
  YougileImportedUser,
  YougileImportStructureResponse,
  YougileImportUsersResponse,
  YougileMappingView,
  YougileProject,
  YougileStartSyncResponse,
  YougileSyncJob,
  YougileSyncRequest,
  YougileTestConnectionResponse,
  YougileTestKeyRequest,
} from '@entities/yougile';

@Injectable({ providedIn: 'root' })
export class IntegrationsApiService {
  private readonly base = `${API_BASE_URL}/v1/integrations`;
  private readonly yougileBase = `${this.base}/yougile`;

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

  getAgileBoardSummary(query?: {
    connection_id?: string;
    board_id?: string;
  }): Observable<BoardSummary> {
    const params = new URLSearchParams();
    if (query?.connection_id) params.set('connection_id', query.connection_id);
    if (query?.board_id) params.set('board_id', query.board_id);
    const suffix = params.toString() ? `?${params.toString()}` : '';
    return this.http.get<BoardSummary>(`${API_BASE_URL}/v1/jira/board-summary${suffix}`);
  }

  getJiraBoardSummary(query?: {
    connection_id?: string;
    board_id?: string;
  }): Observable<BoardSummary> {
    return this.getAgileBoardSummary(query);
  }

  listYougileConnections(): Observable<{ items: YougileConnection[] }> {
    return this.http.get<{ items: YougileConnection[] }>(`${this.yougileBase}/connections`);
  }

  testYougileKey(payload: YougileTestKeyRequest): Observable<YougileTestConnectionResponse> {
    return this.http.post<YougileTestConnectionResponse>(
      `${this.yougileBase}/connections/test-key`,
      payload,
    );
  }

  createYougileConnection(payload: YougileCreateConnectionRequest): Observable<YougileConnection> {
    return this.http.post<YougileConnection>(`${this.yougileBase}/connections`, payload);
  }

  importYougileUsers(id: string): Observable<YougileImportUsersResponse> {
    return this.http.post<YougileImportUsersResponse>(
      `${this.yougileBase}/connections/${id}/import/users`,
      {},
    );
  }

  importYougileStructure(id: string): Observable<YougileImportStructureResponse> {
    return this.http.post<YougileImportStructureResponse>(
      `${this.yougileBase}/connections/${id}/import/structure`,
      {},
    );
  }

  listYougileUsers(id: string): Observable<{ items: YougileImportedUser[] }> {
    return this.http.get<{ items: YougileImportedUser[] }>(
      `${this.yougileBase}/connections/${id}/users`,
    );
  }

  listYougileProjects(id: string): Observable<{ items: YougileProject[] }> {
    return this.http.get<{ items: YougileProject[] }>(
      `${this.yougileBase}/connections/${id}/projects`,
    );
  }

  listYougileBoards(id: string): Observable<{ items: YougileBoard[] }> {
    return this.http.get<{ items: YougileBoard[] }>(`${this.yougileBase}/connections/${id}/boards`);
  }

  listYougileColumns(id: string): Observable<{ items: YougileColumn[] }> {
    return this.http.get<{ items: YougileColumn[] }>(
      `${this.yougileBase}/connections/${id}/columns`,
    );
  }

  listYougileMappings(id: string): Observable<{ content: YougileMappingView[] }> {
    return this.http.get<{ content: YougileMappingView[] }>(
      `${this.yougileBase}/connections/${id}/mappings`,
    );
  }

  autoMatchYougileMappings(
    id: string,
    payload: YougileAutoMatchRequest,
  ): Observable<YougileAutoMatchResponse> {
    return this.http.post<YougileAutoMatchResponse>(
      `${this.yougileBase}/connections/${id}/mappings/auto-match`,
      payload,
    );
  }

  startYougileSync(id: string, payload: YougileSyncRequest): Observable<YougileStartSyncResponse> {
    return this.http.post<YougileStartSyncResponse>(
      `${this.yougileBase}/connections/${id}/sync`,
      payload,
    );
  }

  getYougileSyncJob(jobId: string): Observable<YougileSyncJob> {
    return this.http.get<YougileSyncJob>(`${this.yougileBase}/sync-jobs/${jobId}`);
  }
}
