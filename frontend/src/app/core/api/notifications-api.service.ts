import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';

import { API_BASE_URL } from '@core/config/api.config';
import type { Notification } from '@entities/notification';

@Injectable({ providedIn: 'root' })
export class NotificationsApiService {
  private readonly base = `${API_BASE_URL}/v1/notifications`;

  constructor(private readonly http: HttpClient) {}

  list(): Observable<Notification[]> {
    return this.http.get<Notification[]>(this.base);
  }

  markRead(id: string): Observable<Notification> {
    return this.http.post<Notification>(`${this.base}/${id}/read`, {});
  }

  markAllRead(): Observable<Record<string, unknown>> {
    return this.http.post<Record<string, unknown>>(`${this.base}/read-all`, {});
  }
}