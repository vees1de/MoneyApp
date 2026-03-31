import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';

import { API_BASE_URL } from '@core/config/api.config';
import type { User } from '@entities/user';

@Injectable({ providedIn: 'root' })
export class UsersApiService {
  private readonly base = `${API_BASE_URL}/v1/users`;
  private readonly adminBase = `${API_BASE_URL}/v1/admin/users`;

  constructor(private readonly http: HttpClient) {}

  me(): Observable<{ user: User }> {
    return this.http.get<{ user: User }>(`${this.base}/me`);
  }

  listAdminUsers(): Observable<User[]> {
    return this.http.get<User[]>(this.adminBase);
  }

  createAdminUser(payload: Record<string, unknown>): Observable<User> {
    return this.http.post<User>(this.adminBase, payload);
  }

  updateAdminUser(id: string, payload: Record<string, unknown>): Observable<User> {
    return this.http.patch<User>(`${this.adminBase}/${id}`, payload);
  }

  assignRole(id: string, payload: Record<string, unknown>): Observable<User> {
    return this.http.post<User>(`${this.adminBase}/${id}/roles`, payload);
  }
}