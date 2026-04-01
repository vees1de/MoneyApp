import { CommonModule } from '@angular/common';
import { Component, OnInit, computed, inject, signal } from '@angular/core';
import { RouterLink } from '@angular/router';
import { MatCardModule } from '@angular/material/card';

import { EnrollmentsApiService } from '@core/api/enrollments-api.service';
import type { Enrollment } from '@entities/enrollment';

interface LearningStat {
  label: string;
  value: number;
  tone: 'primary' | 'success' | 'warning' | 'neutral';
}

@Component({
  selector: 'app-page-my-learning-list',
  standalone: true,
  imports: [CommonModule, RouterLink, MatCardModule],
  templateUrl: './list.page.html',
  styleUrl: './list.page.scss',
})
export class MyLearningListPageComponent implements OnInit {
  private readonly api = inject(EnrollmentsApiService);

  protected readonly loading = signal(true);
  protected readonly error = signal<string | null>(null);
  protected readonly enrollments = signal<Enrollment[]>([]);
  protected readonly summaryItems = computed<LearningStat[]>(() => {
    const items = this.enrollments();

    return [
      { label: 'Всего', value: items.length, tone: 'primary' },
      {
        label: 'В процессе',
        value: items.filter((item) => item.status === 'in_progress').length,
        tone: 'warning',
      },
      {
        label: 'Назначено',
        value: items.filter((item) => item.status === 'enrolled').length,
        tone: 'neutral',
      },
      {
        label: 'Завершено',
        value: items.filter((item) => item.status === 'completed').length,
        tone: 'success',
      },
    ];
  });

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

  protected cardToneClass(status: string): string {
    switch (status) {
      case 'completed':
        return 'my-learning-card--done';
      case 'in_progress':
        return 'my-learning-card--active';
      case 'enrolled':
        return 'my-learning-card--pending';
      case 'canceled':
        return 'my-learning-card--canceled';
      default:
        return 'my-learning-card--default';
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

  protected cardNote(item: Enrollment): string {
    switch (item.status) {
      case 'completed':
        return 'Сертификат подтверждён, обучение закрыто автоматически.';
      case 'in_progress':
        return 'Загрузите сертификат на странице курса. После апрува HR обучение завершится автоматически.';
      case 'enrolled':
        return 'Курс назначен. Откройте карточку и загрузите сертификат, когда будете готовы.';
      case 'canceled':
        return 'Назначение отменено.';
      default:
        return 'Откройте карточку обучения для деталей.';
    }
  }
}
