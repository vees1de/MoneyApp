import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';

import { API_BASE_URL } from '@core/config/api.config';
import type { Certificate } from '@entities/certificate';

@Injectable({ providedIn: 'root' })
export class CertificatesApiService {
  private readonly base = `${API_BASE_URL}/v1/certificates`;

  constructor(private readonly http: HttpClient) {}

  listMy(): Observable<Certificate[]> {
    return this.http.get<Certificate[]>(`${this.base}/my`);
  }

  upload(payload: Record<string, unknown>): Observable<Certificate> {
    return this.http.post<Certificate>(`${this.base}/upload`, payload);
  }

  verify(id: string, payload: Record<string, unknown>): Observable<Certificate> {
    return this.http.post<Certificate>(`${this.base}/${id}/verify`, payload);
  }

  reject(id: string, payload: Record<string, unknown>): Observable<Certificate> {
    return this.http.post<Certificate>(`${this.base}/${id}/reject`, payload);
  }
}