import { Component, DestroyRef, inject, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterLink } from '@angular/router';
import { MatCardModule } from '@angular/material/card';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatIconModule } from '@angular/material/icon';
import { MatButtonModule } from '@angular/material/button';
import { MatChipsModule } from '@angular/material/chips';
import { MatExpansionModule } from '@angular/material/expansion';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { HttpErrorResponse } from '@angular/common/http';

import {
  AIApiService,
  type AIRecommendationJobStatus,
  type AIRecommendResponse,
} from '@core/api/ai-api.service';

@Component({
  selector: 'app-page-ai',
  standalone: true,
  imports: [
    CommonModule,
    RouterLink,
    MatCardModule,
    MatProgressSpinnerModule,
    MatIconModule,
    MatButtonModule,
    MatChipsModule,
    MatExpansionModule,
  ],
  templateUrl: './ai.page.html',
  styleUrl: './ai.page.scss',
})
export class AIPageComponent {
  private readonly api = inject(AIApiService);
  private readonly destroyRef = inject(DestroyRef);

  protected readonly loading = signal(false);
  protected readonly error = signal<string | null>(null);
  protected readonly data = signal<AIRecommendResponse | null>(null);
  protected readonly jobStatus = signal<AIRecommendationJobStatus | null>(null);
  protected readonly jobId = signal<string | null>(null);

  protected startRecommendations(): void {
    if (this.loading()) {
      return;
    }

    this.loading.set(true);
    this.error.set(null);
    this.data.set(null);
    this.jobId.set(null);
    this.jobStatus.set('pending');

    this.api
      .runRecommendations()
      .pipe(takeUntilDestroyed(this.destroyRef))
      .subscribe({
        next: (job) => {
          this.jobId.set(job.id);
          this.jobStatus.set(job.status);

          if (job.status === 'done') {
            if (job.result) {
              this.data.set(job.result);
            } else {
              this.error.set('Рекомендации завершились без результата');
            }
            this.loading.set(false);
            return;
          }

          if (job.status === 'failed') {
            this.error.set(job.last_error || 'Не удалось получить рекомендации');
            this.loading.set(false);
            return;
          }
        },
        error: (err) => {
          const message = this.resolveErrorMessage(err);
          this.error.set(message);
          this.jobStatus.set(null);
          this.jobId.set(null);
          this.loading.set(false);
        },
      });
  }

  private resolveErrorMessage(err: unknown): string {
    const httpError = err as HttpErrorResponse | undefined;
    const status = httpError?.status ?? 0;

    if (status === 502 || status === 503 || status === 504 || status === 0) {
      return 'AI-сервис временно недоступен. Запрос повторялся несколько раз, но gateway не вернул ответ.';
    }

    return (
      (httpError?.error as { message?: string } | undefined)?.message ||
      httpError?.message ||
      'Не удалось получить рекомендации'
    );
  }

  protected describeJobStatus(status: AIRecommendationJobStatus | null): string {
    switch (status) {
      case 'pending':
        return 'ожидает запуска';
      case 'processing':
        return 'обрабатывается';
      case 'retry':
        return 'повторная попытка';
      case 'done':
        return 'готово';
      case 'failed':
        return 'ошибка';
      default:
        return 'не запущено';
    }
  }
}
