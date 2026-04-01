import { CommonModule } from '@angular/common';
import { Component, OnInit, inject, signal } from '@angular/core';
import { RouterLink } from '@angular/router';
import { MatCardModule } from '@angular/material/card';
import { MatProgressBarModule } from '@angular/material/progress-bar';

import { EnrollmentsApiService } from '@core/api/enrollments-api.service';
import type { Enrollment } from '@entities/enrollment';

@Component({
  selector: 'app-page-my-learning-list',
  standalone: true,
  imports: [CommonModule, RouterLink, MatCardModule, MatProgressBarModule],
  templateUrl: './list.page.html',
  styleUrl: './list.page.scss',
})
export class MyLearningListPageComponent implements OnInit {
  private readonly api = inject(EnrollmentsApiService);

  protected readonly loading = signal(true);
  protected readonly error = signal<string | null>(null);
  protected readonly enrollments = signal<Enrollment[]>([]);

  ngOnInit(): void {
    this.api.listMy().subscribe({
      next: (enrollments) => {
        this.enrollments.set(enrollments ?? []);
        this.loading.set(false);
      },
      error: () => {
        this.error.set('Не удалось загрузить обучение');
        this.loading.set(false);
      },
    });
  }

  protected progressValue(item: Enrollment): number {
    const value = Number(item.completion_percent);
    return Number.isNaN(value) ? 0 : Math.max(0, Math.min(100, value));
  }

  protected trackByEnrollment(_index: number, item: Enrollment): string {
    return item.id;
  }

  protected shortId(value: string): string {
    if (!value || value.length < 14) {
      return value;
    }

    return `${value.slice(0, 8)}…${value.slice(-4)}`;
  }

  protected statusLabel(status: string): string {
    const labels: Record<string, string> = {
      enrolled: 'Назначено',
      in_progress: 'В процессе',
      completed: 'Завершено',
      canceled: 'Отменено',
    };

    return labels[status] ?? status;
  }

  protected statusChipClass(status: string): string {
    switch (status) {
      case 'completed':
        return 'status-chip--done';
      case 'in_progress':
        return 'status-chip--active';
      case 'enrolled':
        return 'status-chip--pending';
      case 'canceled':
        return 'status-chip--canceled';
      default:
        return 'status-chip--default';
    }
  }

  protected sourceLabel(source: string): string {
    const labels: Record<string, string> = {
      intake: 'Intake',
      manager: 'Manager',
      hr: 'HR',
      self: 'Self',
      catalog: 'Catalog',
    };

    return labels[source] ?? source;
  }

  protected mandatoryLabel(item: Enrollment): string {
    return item.is_mandatory ? 'Обязательное' : 'По выбору';
  }

  protected progressToneClass(item: Enrollment): string {
    if (item.status === 'completed') {
      return 'progress-block--done';
    }

    const progress = this.progressValue(item);
    if (progress >= 70) {
      return 'progress-block--high';
    }

    if (progress >= 35) {
      return 'progress-block--medium';
    }

    return 'progress-block--low';
  }
}
