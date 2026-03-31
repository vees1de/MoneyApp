export interface YougileConnection {
  id: string;
  company_id?: string | null;
  title?: string | null;
  api_base_url: string;
  yougile_company_id: string;
  api_key_last4?: string | null;
  status: string;
  created_by: string;
  last_sync_at?: string | null;
  last_success_sync_at?: string | null;
  last_error?: string | null;
  created_at: string;
  updated_at: string;
}

export interface YougileCreateConnectionRequest {
  title?: string | null;
  apiBaseUrl: string;
  authMode: 'api_key';
  yougileCompanyId: string;
  apiKey: string;
}

export interface YougileCreateKeyRequest {
  login: string;
  password: string;
  companyId: string;
}

export interface YougileDiscoverCompaniesRequest {
  login: string;
  password: string;
  name?: string | null;
}

export interface YougileDiscoverCompaniesResponse {
  paging: {
    count: number;
    limit: number;
    offset: number;
    next: boolean;
  };
  content: YougileCompanyOption[];
}

export interface YougileCompanyOption {
  id: string;
  name: string;
  isAdmin: boolean;
}

export interface YougileConnectRequest {
  login: string;
  password: string;
  companyId: string;
  companyName?: string | null;
}

export interface YougileCreateKeyResponse {
  companyId: string;
  apiKey: string;
  warning: string;
}

export interface YougileTestKeyRequest {
  apiBaseUrl: string;
  apiKey: string;
}

export interface YougileTestConnectionResponse {
  ok: boolean;
  companyAccessible: boolean;
  rateLimitMode: string;
  message: string;
}

export interface YougileImportUsersResponse {
  imported: number;
  updated: number;
  failed: number;
}

export interface YougileImportStructureResponse {
  projectsImported: number;
  boardsImported: number;
  columnsImported: number;
}

export interface YougileImportedUser {
  id: string;
  connection_id: string;
  yougile_user_id: string;
  email?: string | null;
  real_name?: string | null;
  is_admin: boolean;
  status?: string | null;
  last_activity_at?: string | null;
  created_at: string;
  updated_at: string;
}

export interface YougileProject {
  id: string;
  connection_id: string;
  yougile_project_id: string;
  title: string;
  deleted: boolean;
  created_at: string;
  updated_at: string;
}

export interface YougileBoard {
  id: string;
  connection_id: string;
  yougile_board_id: string;
  yougile_project_id: string;
  title: string;
  deleted: boolean;
  created_at: string;
  updated_at: string;
}

export interface YougileColumn {
  id: string;
  connection_id: string;
  yougile_column_id: string;
  yougile_board_id: string;
  title: string;
  color?: number | null;
  deleted: boolean;
  created_at: string;
  updated_at: string;
}

export interface YougileMappingView {
  id: string;
  employee_user_id: string;
  employee_name: string;
  employee_email: string;
  yougile_user_id: string;
  yougile_real_name?: string | null;
  yougile_email?: string | null;
  match_source: string;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface YougileAutoMatchRequest {
  strategy: 'email';
}

export interface YougileAutoMatchResponse {
  matched: number;
  unmatchedInternal: number;
  unmatchedYougile: number;
}

export interface YougileSyncRequest {
  mode: 'incremental' | 'full';
  includeUsers: boolean;
  includeStructure: boolean;
  includeTasks: boolean;
  taskFilters?: Record<string, unknown>;
}

export interface YougileStartSyncResponse {
  jobId: string;
  status: string;
}

export interface YougileSyncJob {
  id: string;
  connection_id: string;
  job_type: string;
  status: string;
  cursor?: Record<string, unknown>;
  progress: Record<string, unknown>;
  started_at?: string | null;
  finished_at?: string | null;
  attempt: number;
  next_retry_at?: string | null;
  error_text?: string | null;
  created_at: string;
  updated_at: string;
}
