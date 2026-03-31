import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';

import { API_BASE_URL } from '@core/config/api.config';
import type { CalendarEvent } from '@entities/calendar-event';

@Injectable({ providedIn: 'root' })
export class CalendarApiService {
  private readonly base = `${API_BASE_URL}/v1/calendar`;

  constructor(private readonly http: HttpClient) {}

  getUpcoming(limit?: number): Observable<CalendarEvent[]> {
    const suffix = typeof limit === 'number' ? `?limit=${limit}` : '';
    return this.http.get<CalendarEvent[]>(`${this.base}/events/upcoming${suffix}`);
  }
}