import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable, map } from 'rxjs';

import type { IdentityUserView, MeResponse } from '@core/auth/auth.types';
import { API_BASE_URL } from '@core/config/api.config';
import type { ListResponse } from './api.types';

@Injectable({ providedIn: 'root' })
export class UsersApiService {
  private readonly base = `${API_BASE_URL}/v1/users`;
  private readonly adminBase = `${API_BASE_URL}/v1/admin/users`;

  constructor(private readonly http: HttpClient) {}

  me(): Observable<MeResponse> {
    return this.http.get<MeResponse>(`${this.base}/me`);
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
