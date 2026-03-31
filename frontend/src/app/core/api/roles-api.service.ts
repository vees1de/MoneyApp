import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';

import { API_BASE_URL } from '@core/config/api.config';

@Injectable({ providedIn: 'root' })
export class RolesApiService {
  private readonly base = `${API_BASE_URL}/v1/admin/roles`;

  constructor(private readonly http: HttpClient) {}

  list(): Observable<Record<string, unknown>[]> {
    return this.http.get<Record<string, unknown>[]>(this.base);
  }
}