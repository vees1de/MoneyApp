import { CommonModule } from '@angular/common';
import { Component, OnInit, computed, inject, signal } from '@angular/core';
import { RouterLink } from '@angular/router';

import { EnrollmentsApiService } from '@core/api/enrollments-api.service';
import type { Enrollment } from '@entities/enrollment';

interface LearningSummaryItem {
  label: string;
  value: number;
}

interface LearningCardView {
  id: string;
  title: string;
  sourceLabel: string;
  mandatoryLabel: string;
  status: string;
  statusLabel: string;
  statusClass: string;
  deadlineLabel: string;
  deadlineCaption: string;
  deadlineToneClass: string;
  metaLabel: string;
  progressLabel: string | null;
}

@Component({
  selector: 'app-page-my-learning-list',
  standalone: true,
  imports: [CommonModule, RouterLink],
  templateUrl: './list.page.html',
  styleUrl: './list.page.scss',
})
export class MyLearningListPageComponent implements OnInit {
  private readonly api = inject(EnrollmentsApiService);
  private readonly shortDateFormatter = new Intl.DateTimeFormat('ru-RU', {
    day: 'numeric',
    month: 'long',
  });
  private readonly shortDateTimeFormatter = new Intl.DateTimeFormat('ru-RU', {
    day: 'numeric',
    month: 'short',
    hour: '2-digit',
    minute: '2-digit',
  });

  protected readonly loading = signal(true);
  protected readonly error = signal<string | null>(null);
  protected readonly enrollments = signal<Enrollment[]>([]);

  protected readonly orderedEnrollments = computed(() =>
    [...this.enrollments()].sort((left, right) => this.compareEnrollments(left, right)),
  );

  protected readonly cards = computed<LearningCardView[]>(() =>
    this.orderedEnrollments().map((item) => this.toCardView(item)),
  );

  protected readonly summaryItems = computed<LearningSummaryItem[]>(() => {
    const items = this.enrollments();

    return [
      { label: 'Активно', value: items.filter((item) => this.isActive(item)).length },
      { label: 'Скоро срок', value: items.filter((item) => this.isDueSoon(item)).length },
      { label: 'Завершено', value: items.filter((item) => item.status === 'completed').length },
    ];
  });

  protected readonly primaryEnrollment = computed(() => {
    const items = this.orderedEnrollments();
    return items.find((item) => this.isActive(item)) ?? items[0] ?? null;
  });

  protected readonly primaryActionLabel = computed(() => {
    const item = this.primaryEnrollment();
    if (!item) {
      return '';
    }

    return this.isActive(item) ? 'Открыть ближайший курс' : 'Открыть обучение';
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

  protected trackByEnrollment(_index: number, item: LearningCardView): string {
    return item.id;
  }

  private compareEnrollments(left: Enrollment, right: Enrollment): number {
    const statusDelta = this.statusOrder(left.status) - this.statusOrder(right.status);
    if (statusDelta !== 0) {
      return statusDelta;
    }

    const deadlineDelta =
      this.deadlineTimestamp(left.deadline_at) - this.deadlineTimestamp(right.deadline_at);
    if (deadlineDelta !== 0) {
      return deadlineDelta;
    }

    return this.safeTimestamp(right.updated_at) - this.safeTimestamp(left.updated_at);
  }

  private statusOrder(status: string): number {
    switch (status) {
      case 'in_progress':
        return 0;
      case 'enrolled':
        return 1;
      case 'completed':
        return 2;
      case 'canceled':
        return 3;
      default:
        return 4;
    }
  }

  private toCardView(item: Enrollment): LearningCardView {
    return {
      id: item.id,
      title: item.course_title?.trim() || `Курс ${item.course_id.slice(0, 8)}`,
      sourceLabel: this.sourceLabel(item.source),
      mandatoryLabel: item.is_mandatory ? 'Обязательное' : 'По выбору',
      status: item.status,
      statusLabel: this.statusLabel(item.status),
      statusClass: this.statusChipClass(item.status),
      deadlineLabel: this.deadlineLabel(item.deadline_at),
      deadlineCaption: this.deadlineCaption(item),
      deadlineToneClass: this.deadlineToneClass(item),
      metaLabel: this.metaLabel(item),
      progressLabel: this.progressLabel(item),
    };
  }

  private sourceLabel(source: string): string {
    const labels: Record<string, string> = {
      intake: 'Intake',
      manager: 'Manager',
      hr: 'HR',
      self: 'Self',
      catalog: 'Catalog',
    };

    return labels[source] ?? source;
  }

  private statusLabel(status: string): string {
    const labels: Record<string, string> = {
      enrolled: 'Назначено',
      in_progress: 'В процессе',
      completed: 'Завершено',
      canceled: 'Отменено',
    };

    return labels[status] ?? status;
  }

  private statusChipClass(status: string): string {
    switch (status) {
      case 'completed':
        return 'status-pill--done';
      case 'in_progress':
        return 'status-pill--active';
      case 'enrolled':
        return 'status-pill--pending';
      case 'canceled':
        return 'status-pill--canceled';
      default:
        return 'status-pill--default';
    }
  }

  private deadlineToneClass(item: Enrollment): string {
    if (this.isOverdue(item)) {
      return 'learning-card--overdue';
    }
    if (this.isDueSoon(item)) {
      return 'learning-card--soon';
    }
    if (item.status === 'completed') {
      return 'learning-card--done';
    }
    return 'learning-card--default';
  }

  private deadlineLabel(value?: string | null): string {
    const date = this.parseDate(value);
    if (!date) {
      return 'Без дедлайна';
    }
    return this.shortDateFormatter.format(date);
  }

  private deadlineCaption(item: Enrollment): string {
    const date = this.parseDate(item.deadline_at);
    if (!date) {
      return 'Срок не задан';
    }
    if (item.status === 'completed') {
      return 'Срок закрыт';
    }

    const today = Date.now();
    const diff = date.getTime() - today;
    const days = Math.ceil(diff / 86_400_000);

    if (days < 0) {
      return 'Просрочено';
    }
    if (days === 0) {
      return 'Сегодня';
    }
    if (days === 1) {
      return '1 день';
    }
    if (days <= 4) {
      return `${days} дня`;
    }
    return `${days} дней`;
  }

  private metaLabel(item: Enrollment): string {
    if (item.status === 'completed' && item.completed_at) {
      return `Завершено ${this.formatDateTime(item.completed_at)}`;
    }
    if (item.last_activity_at) {
      return `Активность ${this.formatDateTime(item.last_activity_at)}`;
    }
    if (item.started_at) {
      return `Начато ${this.formatDateTime(item.started_at)}`;
    }
    return `Назначено ${this.formatDateTime(item.enrolled_at)}`;
  }

  private progressLabel(item: Enrollment): string | null {
    if (item.status === 'completed') {
      return '100%';
    }

    const progress = Number(item.completion_percent);
    if (Number.isNaN(progress) || progress <= 0) {
      return null;
    }

    return `${Math.round(progress)}%`;
  }

  private formatDateTime(value?: string | null): string {
    const date = this.parseDate(value);
    if (!date) {
      return '—';
    }
    return this.shortDateTimeFormatter.format(date).replace('.', '');
  }

  private parseDate(value?: string | null): Date | null {
    if (!value) {
      return null;
    }

    const date = new Date(value);
    return Number.isNaN(date.getTime()) ? null : date;
  }

  private safeTimestamp(value?: string | null): number {
    return this.parseDate(value)?.getTime() ?? 0;
  }

  private deadlineTimestamp(value?: string | null): number {
    return this.parseDate(value)?.getTime() ?? Number.POSITIVE_INFINITY;
  }

  private isActive(item: Enrollment): boolean {
    return item.status === 'enrolled' || item.status === 'in_progress';
  }

  private isOverdue(item: Enrollment): boolean {
    const deadline = this.parseDate(item.deadline_at);
    if (!deadline || !this.isActive(item)) {
      return false;
    }
    return deadline.getTime() < Date.now();
  }

  private isDueSoon(item: Enrollment): boolean {
    const deadline = this.parseDate(item.deadline_at);
    if (!deadline || !this.isActive(item)) {
      return false;
    }

    const diff = deadline.getTime() - Date.now();
    const days = Math.ceil(diff / 86_400_000);
    return days >= 0 && days <= 7;
  }
}
