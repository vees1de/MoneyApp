import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';

import { API_BASE_URL } from '@core/config/api.config';

export interface AIRecommendation {
  course_id: string;
  title: string;
  short_description?: string | null;
  reason: string;
}

export interface AIDebugLog {
  prompt_sent_to_ai: string;
  ai_raw_response: string;
  ai_model_uri: string;
  tasks_summary: string;
  courses_summary: string;
}

export interface AIRecommendResponse {
  tasks_analyzed: number;
  courses_in_pool: number;
  recommendations: AIRecommendation[];
  debug?: AIDebugLog | null;
}

@Injectable({ providedIn: 'root' })
export class AIApiService {
  private readonly base = `${API_BASE_URL}/v1/ai`;

  constructor(private readonly http: HttpClient) {}

  getRecommendations(): Observable<AIRecommendResponse> {
    return this.http.get<AIRecommendResponse>(`${this.base}/recommendations`);
  }
}
