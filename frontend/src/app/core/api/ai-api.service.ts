import { Injectable } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { EMPTY, Observable, expand, retry, switchMap, throwError, timer } from 'rxjs';

import { API_BASE_URL } from '@core/config/api.config';

export interface AIRecommendation {
  course_id: string;
  title: string;
  short_description?: string | null;
  reason: string;
}

export interface AIIntakeRecommendation {
  intake_id: string;
  course_id?: string | null;
  title: string;
  description?: string | null;
  reason: string;
  start_date?: string | null;
  application_deadline?: string | null;
}

export interface AIDebugLog {
  prompt_sent_to_ai: string;
  ai_raw_response: string;
  ai_model_uri: string;
  ai_request_duration_ms?: number | null;
  tasks_summary: string;
  courses_summary: string;
  intakes_summary: string;
  courses_source?: string | null;
  intakes_source?: string | null;
  courses_error?: string | null;
  intakes_error?: string | null;
}

export interface AIRecommendResponse {
  tasks_analyzed: number;
  courses_in_pool: number;
  intakes_in_pool: number;
  recommendations: AIRecommendation[];
  intake_recommendations: AIIntakeRecommendation[];
  debug?: AIDebugLog | null;
}

export type AIRecommendationJobStatus = 'pending' | 'processing' | 'retry' | 'done' | 'failed';

export interface AIRecommendationJob {
  id: string;
  user_id: string;
  status: AIRecommendationJobStatus;
  attempt: number;
  result?: AIRecommendResponse | null;
  last_error?: string | null;
  started_at?: string | null;
  finished_at?: string | null;
  created_at: string;
  updated_at: string;
}

@Injectable({ providedIn: 'root' })
export class AIApiService {
  private readonly base = `${API_BASE_URL}/v1/ai`;

  constructor(private readonly http: HttpClient) {}

  getRecommendations(): Observable<AIRecommendResponse> {
    return this.http.get<AIRecommendResponse>(`${this.base}/recommendations`);
  }

  startRecommendationJob(): Observable<AIRecommendationJob> {
    return this.http.post<AIRecommendationJob>(`${this.base}/recommendations`, {});
  }

  listRecommendationJobs(limit = 10): Observable<AIRecommendationJob[]> {
    const params = new HttpParams().set('limit', String(limit));
    return this.http.get<AIRecommendationJob[]>(`${this.base}/recommendations/jobs`, { params });
  }

  getRecommendationJob(jobId: string): Observable<AIRecommendationJob> {
    return this.http.get<AIRecommendationJob>(`${this.base}/recommendations/${jobId}`).pipe(
      retry({
        count: 3,
        delay: (error, retryCount) => {
          if (!this.isRetriableGatewayError(error)) {
            return throwError(() => error);
          }

          const delayMs = Math.min(500 * retryCount, 2000);
          return timer(delayMs);
        },
      }),
    );
  }

  watchRecommendationJob(jobId: string): Observable<AIRecommendationJob> {
    const startedAt = Date.now();

    return this.getRecommendationJob(jobId).pipe(
      expand((job) => {
        if (this.isTerminalJob(job.status)) {
          return EMPTY;
        }

        return timer(this.getRecommendationDelayMs(Date.now() - startedAt)).pipe(
          switchMap(() => this.getRecommendationJob(jobId)),
        );
      }),
    );
  }

  runRecommendations(): Observable<AIRecommendationJob> {
    return this.startRecommendationJob().pipe(switchMap((job) => this.watchRecommendationJob(job.id)));
  }

  private isTerminalJob(status: AIRecommendationJobStatus): boolean {
    return status === 'done' || status === 'failed';
  }

  private isRetriableGatewayError(error: unknown): boolean {
    const status = (error as { status?: number } | null | undefined)?.status ?? 0;
    return status === 0 || status === 502 || status === 503 || status === 504;
  }

  private getRecommendationDelayMs(elapsedMs: number): number {
    if (elapsedMs < 10_000) {
      return 2000;
    }

    if (elapsedMs < 30_000) {
      return 4000;
    }

    return 8000;
  }
}
