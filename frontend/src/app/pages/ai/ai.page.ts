import { Component, DestroyRef, OnDestroy, OnInit, inject, signal } from '@angular/core';
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
import { Subscription } from 'rxjs';

import {
  AIApiService,
  type AIRecommendationJob,
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
export class AIPageComponent implements OnInit, OnDestroy {
  private readonly api = inject(AIApiService);
  private readonly destroyRef = inject(DestroyRef);
  private currentJobSubscription?: Subscription;

  protected readonly loading = signal(false);
  protected readonly error = signal<string | null>(null);
  protected readonly data = signal<AIRecommendResponse | null>(null);
  protected readonly jobStatus = signal<AIRecommendationJobStatus | null>(null);
  protected readonly jobId = signal<string | null>(null);
  protected readonly history = signal<AIRecommendationJob[]>([]);
  protected readonly historyLoading = signal(false);
  protected readonly historyError = signal<string | null>(null);
  protected readonly deletingJobId = signal<string | null>(null);

  ngOnInit(): void {
    this.loadHistory();
  }

  ngOnDestroy(): void {
    this.cancelCurrentJobWatch();
  }

  protected startRecommendations(): void {
    if (this.loading()) {
      return;
    }

    this.cancelCurrentJobWatch();
    this.loading.set(true);
    this.error.set(null);
    this.data.set(null);
    this.jobId.set(null);
    this.jobStatus.set('pending');

    this.api
      .startRecommendationJob()
      .pipe(takeUntilDestroyed(this.destroyRef))
      .subscribe({
        next: (job) => {
          this.upsertHistory(job);
          this.observeJob(job.id);
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

  protected openJob(jobId: string): void {
    if (!jobId) {
      return;
    }

    this.observeJob(jobId);
  }

  protected deleteJob(job: AIRecommendationJob, event?: MouseEvent): void {
    event?.stopPropagation();

    if (this.deletingJobId() === job.id) {
      return;
    }

    const confirmed = typeof window === 'undefined' ? true : window.confirm('Удалить этот запуск AI-рекомендаций?');
    if (!confirmed) {
      return;
    }

    const wasCurrent = this.jobId() === job.id;
    if (wasCurrent) {
      this.cancelCurrentJobWatch();
      this.loading.set(false);
      this.jobStatus.set(null);
    }

    this.deletingJobId.set(job.id);

    this.api
      .deleteRecommendationJob(job.id)
      .pipe(takeUntilDestroyed(this.destroyRef))
      .subscribe({
        next: () => {
          this.deletingJobId.set(null);
          this.history.set(this.history().filter((item) => item.id !== job.id));
          this.error.set(null);
          if (wasCurrent) {
            this.data.set(null);
            this.jobId.set(null);
            this.jobStatus.set(null);
          }
          this.loadHistory();
        },
        error: (err) => {
          this.deletingJobId.set(null);
          this.error.set(this.resolveErrorMessage(err));
          if (wasCurrent) {
            this.loadHistory();
          }
        },
      });
  }

  protected loadHistory(): void {
    this.historyLoading.set(true);
    this.historyError.set(null);

    this.api
      .listRecommendationJobs(20)
      .pipe(takeUntilDestroyed(this.destroyRef))
      .subscribe({
        next: (jobs) => {
          this.history.set(jobs);
          this.historyLoading.set(false);

          if (this.currentJobSubscription) {
            return;
          }

          const activeJob = jobs.find((job) => this.isActiveJob(job.status));
          const targetJob = activeJob ?? jobs[0];

          if (targetJob) {
            this.observeJob(targetJob.id);
          } else {
            this.cancelCurrentJobWatch();
            this.loading.set(false);
            this.jobId.set(null);
            this.jobStatus.set(null);
            this.data.set(null);
          }
        },
        error: (err) => {
          this.historyLoading.set(false);
          this.historyError.set(this.resolveHistoryErrorMessage(err));
          if (!this.data()) {
            this.loading.set(false);
          }
        },
      });
  }

  private observeJob(jobId: string): void {
    this.cancelCurrentJobWatch();
    this.loading.set(true);
    this.error.set(null);
    this.data.set(null);
    this.jobId.set(jobId);
    this.jobStatus.set('pending');

    this.currentJobSubscription = this.api.watchRecommendationJob(jobId).subscribe({
      next: (job) => {
        this.upsertHistory(job);
        this.jobId.set(job.id);
        this.jobStatus.set(job.status);

        if (job.status === 'done') {
          if (job.result) {
            this.data.set(job.result);
          } else {
            this.error.set('Рекомендации завершились без результата');
          }
          this.loading.set(false);
          this.currentJobSubscription = undefined;
          return;
        }

        if (job.status === 'failed') {
          this.error.set(job.last_error || 'Не удалось получить рекомендации');
          this.loading.set(false);
          this.currentJobSubscription = undefined;
        }
      },
      error: (err) => {
        this.error.set(this.resolveErrorMessage(err));
        this.jobStatus.set(null);
        this.loading.set(false);
        this.currentJobSubscription = undefined;
      },
      complete: () => {
        this.currentJobSubscription = undefined;
      },
    });
  }

  private upsertHistory(job: AIRecommendationJob): void {
    const history = this.history();
    const nextHistory = [...history];
    const index = nextHistory.findIndex((item) => item.id === job.id);
    if (index >= 0) {
      nextHistory[index] = job;
    } else {
      nextHistory.unshift(job);
    }
    this.history.set(nextHistory);
  }

  private cancelCurrentJobWatch(): void {
    this.currentJobSubscription?.unsubscribe();
    this.currentJobSubscription = undefined;
  }

  private isActiveJob(status: AIRecommendationJobStatus): boolean {
    return status === 'pending' || status === 'processing' || status === 'retry';
  }

  private resolveErrorMessage(err: unknown): string {
    const httpError = err as HttpErrorResponse | undefined;
    const status = httpError?.status ?? 0;

    if (status === 404) {
      return 'Запуск был удалён.';
    }

    if (status === 502 || status === 503 || status === 504 || status === 0) {
      return 'AI-сервис временно недоступен. Запрос повторялся несколько раз, но gateway не вернул ответ.';
    }

    return (
      (httpError?.error as { message?: string } | undefined)?.message ||
      httpError?.message ||
      'Не удалось получить рекомендации'
    );
  }

  private resolveHistoryErrorMessage(err: unknown): string {
    const httpError = err as HttpErrorResponse | undefined;
    const status = httpError?.status ?? 0;

    if (status === 502 || status === 503 || status === 504 || status === 0) {
      return 'Не удалось загрузить историю запусков AI. Gateway временно недоступен.';
    }

    return (
      (httpError?.error as { message?: string } | undefined)?.message ||
      httpError?.message ||
      'Не удалось загрузить историю запусков'
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
