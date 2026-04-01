import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import type {
  BoardSummary,
  OutlookConnectResponse,
  OutlookEventRecord,
  OutlookIntegrationStatus,
  OutlookManualConnectRequest,
  OutlookMessageRecord,
  OutlookSyncResponse,
  OutlookTestEmailRequest,
  OutlookTestEmailResponse,
  OutlookUpdateSettingsRequest,
} from './contracts';
import { API_BASE_URL } from '@core/config/api.config';

import type {
  YougileAutoMatchRequest,
  YougileAutoMatchResponse,
  YougileBoard,
  YougileColumn,
  YougileConnectRequest,
  YougileConnection,
  YougileCreateConnectionRequest,
  YougileDiscoverCompaniesRequest,
  YougileDiscoverCompaniesResponse,
  YougileCreateKeyRequest,
  YougileCreateKeyResponse,
  YougileImportedUser,
  YougileImportStructureResponse,
  YougileImportUsersResponse,
  YougileMappingView,
  YougileProject,
  YougileStartSyncResponse,
  YougileTask,
  YougileTasksResponse,
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

  getOutlookConnectLink(): Observable<OutlookConnectResponse> {
    return this.http.get<OutlookConnectResponse>(`${this.base}/outlook/connect`);
  }

  connectOutlookManual(payload: OutlookManualConnectRequest): Observable<OutlookIntegrationStatus> {
    return this.http.post<OutlookIntegrationStatus>(`${this.base}/outlook/connect/manual`, payload);
  }

  getOutlookStatus(): Observable<OutlookIntegrationStatus> {
    return this.http.get<OutlookIntegrationStatus>(`${this.base}/outlook/status`);
  }

  syncOutlook(): Observable<OutlookSyncResponse> {
    return this.http.post<OutlookSyncResponse>(`${this.base}/outlook/sync`, {});
  }

  listOutlookMessages(limit = 12): Observable<{ items: OutlookMessageRecord[] }> {
    return this.http.get<{ items: OutlookMessageRecord[] }>(
      `${this.base}/outlook/messages?limit=${limit}`,
    );
  }

  listOutlookEvents(limit = 12): Observable<{ items: OutlookEventRecord[] }> {
    return this.http.get<{ items: OutlookEventRecord[] }>(
      `${this.base}/outlook/events?limit=${limit}`,
    );
  }

  updateOutlookSettings(
    payload: OutlookUpdateSettingsRequest,
  ): Observable<OutlookIntegrationStatus> {
    return this.http.post<OutlookIntegrationStatus>(`${this.base}/outlook/settings`, payload);
  }

  sendOutlookTestEmail(
    payload: OutlookTestEmailRequest = {},
  ): Observable<OutlookTestEmailResponse> {
    return this.http.post<OutlookTestEmailResponse>(`${this.base}/outlook/test-email`, payload);
  }

  disconnectOutlook(): Observable<void> {
    return this.http.post<void>(`${this.base}/outlook/disconnect`, {});
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

  connectYougileConnection(payload: YougileConnectRequest): Observable<YougileConnection> {
    return this.http.post<YougileConnection>(`${this.yougileBase}/connections/connect`, payload);
  }

  createYougileKey(payload: YougileCreateKeyRequest): Observable<YougileCreateKeyResponse> {
    return this.http.post<YougileCreateKeyResponse>(
      `${this.yougileBase}/connections/create-key`,
      payload,
    );
  }

  discoverYougileCompanies(
    payload: YougileDiscoverCompaniesRequest,
  ): Observable<YougileDiscoverCompaniesResponse> {
    return this.http.post<YougileDiscoverCompaniesResponse>(
      `${this.yougileBase}/discover-companies`,
      payload,
    );
  }

  deleteYougileConnection(id: string): Observable<void> {
    return this.http.delete<void>(`${this.yougileBase}/connections/${id}`);
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

  listYougileTasks(
    id: string,
    query?: {
      assignedTo?: string;
      columnId?: string;
      includeDeleted?: boolean;
      limit?: number;
      mineOnly?: boolean;
      offset?: number;
      title?: string;
    },
  ): Observable<YougileTasksResponse> {
    const params = new URLSearchParams();
    if (query?.assignedTo) params.set('assignedTo', query.assignedTo);
    if (query?.columnId) params.set('columnId', query.columnId);
    if (typeof query?.includeDeleted === 'boolean') {
      params.set('includeDeleted', String(query.includeDeleted));
    }
    if (typeof query?.mineOnly === 'boolean') {
      params.set('mineOnly', String(query.mineOnly));
    }
    if (typeof query?.limit === 'number') params.set('limit', String(query.limit));
    if (typeof query?.offset === 'number') params.set('offset', String(query.offset));
    if (query?.title) params.set('title', query.title);
    const suffix = params.toString() ? `?${params.toString()}` : '';
    return this.http.get<YougileTasksResponse>(
      `${this.yougileBase}/connections/${id}/tasks${suffix}`,
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
