import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable, map } from 'rxjs';

import { API_BASE_URL } from '@core/config/api.config';
import type { ListResponse } from './api.types';
import type { AdminPermission, AdminRole } from './contracts';

@Injectable({ providedIn: 'root' })
export class RolesApiService {
  private readonly base = `${API_BASE_URL}/v1/admin/roles`;
  private readonly permissionsBase = `${API_BASE_URL}/v1/admin/permissions`;

  constructor(private readonly http: HttpClient) {}

  list(): Observable<AdminRole[]> {
    return this.http
      .get<ListResponse<AdminRole>>(this.base)
      .pipe(map((response) => response.items ?? []));
  }

  listPermissions(): Observable<AdminPermission[]> {
    return this.http
      .get<ListResponse<AdminPermission>>(this.permissionsBase)
      .pipe(map((response) => response.items ?? []));
  }
}
