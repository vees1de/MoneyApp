import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable, map } from 'rxjs';

import { API_BASE_URL } from '@core/config/api.config';
import type {
  CreateDevelopmentTeamRequest,
  DevelopmentTeam,
  EmployeesLearningStatsQuery,
  EmployeesLearningStatsResponse,
  EmployeeProfileResponse,
  ProfileMeResponse,
  ProfileRole,
  UpdateUserProfileRequest,
} from './contracts';
import type { ListQuery, ListResponse } from './api.types';
import type { IdentityUserView } from '@core/auth/auth.types';
import { toHttpParams } from './http-params.util';

@Injectable({ providedIn: 'root' })
export class UsersApiService {
  private readonly base = `${API_BASE_URL}/v1/users`;
  private readonly adminBase = `${API_BASE_URL}/v1/admin/users`;

  constructor(private readonly http: HttpClient) {}

  me(): Observable<ProfileMeResponse> {
    return this.http.get<ProfileMeResponse>(`${this.base}/me`);
  }

  updateMe(payload: UpdateUserProfileRequest): Observable<ProfileMeResponse> {
    return this.http.patch<ProfileMeResponse>(`${this.base}/me`, payload);
  }

  uploadAvatar(file: File): Observable<ProfileMeResponse> {
    const formData = new FormData();
    formData.append('file', file);
    return this.http.post<ProfileMeResponse>(`${this.base}/me/avatar`, formData);
  }

  listProfileRoles(): Observable<ProfileRole[]> {
    return this.http
      .get<ListResponse<ProfileRole>>(`${this.base}/profile-roles`)
      .pipe(map((response) => response.items ?? []));
  }

  listDevelopmentTeams(): Observable<DevelopmentTeam[]> {
    return this.http
      .get<ListResponse<DevelopmentTeam>>(`${this.base}/development-teams`)
      .pipe(map((response) => response.items ?? []));
  }

  listAvailableDevelopmentTeams(): Observable<DevelopmentTeam[]> {
    return this.http
      .get<ListResponse<DevelopmentTeam>>(`${this.base}/development-teams/available`)
      .pipe(map((response) => response.items ?? []));
  }

  createDevelopmentTeam(payload: CreateDevelopmentTeamRequest): Observable<ProfileMeResponse> {
    return this.http.post<ProfileMeResponse>(`${this.base}/development-teams`, payload);
  }

  joinDevelopmentTeam(id: string): Observable<ProfileMeResponse> {
    return this.http.post<ProfileMeResponse>(`${this.base}/development-teams/${id}/join`, {});
  }

  leaveCurrentDevelopmentTeam(): Observable<ProfileMeResponse> {
    return this.http.post<ProfileMeResponse>(`${this.base}/development-teams/current/leave`, {});
  }

  getEmployeeProfile(userId: string): Observable<EmployeeProfileResponse> {
    return this.http.get<EmployeeProfileResponse>(`${API_BASE_URL}/v1/employees/${userId}`);
  }

  getEmployeesLearningStats(
    query?: EmployeesLearningStatsQuery,
  ): Observable<EmployeesLearningStatsResponse> {
    return this.http.get<EmployeesLearningStatsResponse>(
      `${API_BASE_URL}/v1/employees/learning-stats`,
      {
        params: toHttpParams(query as ListQuery),
      },
    );
  }

  listAdminUsers(): Observable<IdentityUserView[]> {
    return this.http
      .get<ListResponse<IdentityUserView>>(this.adminBase)
      .pipe(map((response) => response.items ?? []));
  }

  createAdminUser(payload: Record<string, unknown>): Observable<IdentityUserView> {
    return this.http.post<IdentityUserView>(this.adminBase, payload);
  }

  updateAdminUser(id: string, payload: Record<string, unknown>): Observable<IdentityUserView> {
    return this.http.patch<IdentityUserView>(`${this.adminBase}/${id}`, payload);
  }

  assignRole(id: string, payload: Record<string, unknown>): Observable<void> {
    return this.http.post<void>(`${this.adminBase}/${id}/roles`, payload);
  }

  removeRole(id: string, roleId: string): Observable<void> {
    return this.http.delete<void>(`${this.adminBase}/${id}/roles/${roleId}`);
  }
}
