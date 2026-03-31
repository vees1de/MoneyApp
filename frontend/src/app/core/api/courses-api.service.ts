import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';

import { API_BASE_URL } from '@core/config/api.config';
import type { Course } from '@entities/course';
import { toHttpParams } from './http-params.util';
import type { ListQuery } from './api.types';

@Injectable({ providedIn: 'root' })
export class CoursesApiService {
  private readonly base = `${API_BASE_URL}/v1/courses`;

  constructor(private readonly http: HttpClient) {}

  list(query?: ListQuery): Observable<Course[]> {
    return this.http.get<Course[]>(this.base, { params: toHttpParams(query) });
  }

  getById(id: string): Observable<Course> {
    return this.http.get<Course>(`${this.base}/${id}`);
  }

  getMaterials(id: string): Observable<Record<string, unknown>[]> {
    return this.http.get<Record<string, unknown>[]>(`${this.base}/${id}/materials`);
  }

  create(payload: Record<string, unknown>): Observable<Course> {
    return this.http.post<Course>(this.base, payload);
  }

  update(id: string, payload: Record<string, unknown>): Observable<Course> {
    return this.http.patch<Course>(`${this.base}/${id}`, payload);
  }

  publish(id: string): Observable<Course> {
    return this.http.post<Course>(`${this.base}/${id}/publish`, {});
  }

  archive(id: string): Observable<Course> {
    return this.http.post<Course>(`${this.base}/${id}/archive`, {});
  }

  addMaterial(id: string, payload: Record<string, unknown>): Observable<Record<string, unknown>> {
    return this.http.post<Record<string, unknown>>(`${this.base}/${id}/materials`, payload);
  }
}
