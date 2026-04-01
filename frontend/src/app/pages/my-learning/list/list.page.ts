import { CommonModule } from '@angular/common';
import { Component, OnInit, computed, inject, signal } from '@angular/core';
import { RouterLink } from '@angular/router';

import { EnrollmentsApiService } from '@core/api/enrollments-api.service';
import type { Enrollment } from '@entities/enrollment';

type LearningFilterKey = 'all' | 'active' | 'attention' | 'completed';

interface LearningFilterItem {
  key: LearningFilterKey;
  label: string;
  count: number;
}

interface LearningCardView {
  id: string;
  title: string;
  sourceLabel: string;
  mandatoryLabel: string;
  statusLabel: string;
  statusClass: string;
  deadlineText: string;
  deadlineHint: string;
  deadlineClass: string;
  progressText: string;
  activityText: string;
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
    day: '2-digit',
    month: 'short',
  });
  private readonly shortDateTimeFormatter = new Intl.DateTimeFormat('ru-RU', {
    day: '2-digit',
    month: 'short',
    hour: '2-digit',
    minute: '2-digit',
  });

  protected readonly loading = signal(true);
  protected readonly error = signal<string | null>(null);
  protected readonly enrollments = signal<Enrollment[]>([]);
  protected readonly selectedFilter = signal<LearningFilterKey>('all');

  protected readonly orderedEnrollments = computed(() =>
    [...this.enrollments()].sort((left, right) => this.compareEnrollments(left, right)),
  );

  protected readonly filterItems = computed<LearningFilterItem[]>(() => {
    const items = this.orderedEnrollments();

    return [
      { key: 'all', label: 'Все', count: items.length },
      {
        key: 'active',
        label: 'Активные',
        count: items.filter((item) => this.isActive(item)).length,
      },
      {
        key: 'attention',
        label: 'Требуют внимания',
        count: items.filter((item) => this.needsAttention(item)).length,
      },
      {
        key: 'completed',
        label: 'Завершённые',
        count: items.filter((item) => item.status === 'completed').length,
      },
    ];
  });

  protected readonly filteredEnrollments = computed(() => {
    const items = this.orderedEnrollments();

    switch (this.selectedFilter()) {
      case 'active':
        return items.filter((item) => this.isActive(item));
      case 'attention':
        return items.filter((item) => this.needsAttention(item));
      case 'completed':
        return items.filter((item) => item.status === 'completed');
      default:
        return items;
    }
  });

  protected readonly filteredCards = computed<LearningCardView[]>(() =>
    this.filteredEnrollments().map((item) => this.toCardView(item)),
  );

  protected readonly resultLabel = computed(() =>
    this.formatCourseCount(this.filteredCards().length),
  );

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

  protected setFilter(filter: LearningFilterKey): void {
    this.selectedFilter.set(filter);
  }

  protected isFilterActive(filter: LearningFilterKey): boolean {
    return this.selectedFilter() === filter;
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
      statusLabel: this.statusLabel(item.status),
      statusClass: this.statusClass(item.status),
      deadlineText: this.deadlineText(item.deadline_at),
      deadlineHint: this.deadlineHint(item),
      deadlineClass: this.deadlineClass(item),
      progressText: this.progressText(item),
      activityText: this.activityText(item),
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

  private statusClass(status: string): string {
    switch (status) {
      case 'completed':
        return 'status-pill--success';
      case 'in_progress':
        return 'status-pill--info';
      case 'enrolled':
        return 'status-pill--warning';
      case 'canceled':
        return 'status-pill--critical';
      default:
        return 'status-pill--default';
    }
  }

  private deadlineText(value?: string | null): string {
    const date = this.parseDate(value);
    if (!date) {
      return 'Не задан';
    }
    return this.shortDateFormatter.format(date).replace('.', '');
  }

  private deadlineHint(item: Enrollment): string {
    const date = this.parseDate(item.deadline_at);
    if (!date) {
      return 'Без ограничения по сроку';
    }
    if (item.status === 'completed') {
      return 'Срок закрыт';
    }

    const days = Math.ceil((date.getTime() - Date.now()) / 86_400_000);
    if (days < 0) {
      return `Просрочено на ${Math.abs(days)} дн.`;
    }
    if (days === 0) {
      return 'Требует действия сегодня';
    }
    if (days <= 7) {
      return `Осталось ${days} дн.`;
    }
    return 'Срок в норме';
  }

  private deadlineClass(item: Enrollment): string {
    if (this.isOverdue(item)) {
      return 'meta-item__value--critical';
    }
    if (this.isDueSoon(item)) {
      return 'meta-item__value--warning';
    }
    if (item.status === 'completed') {
      return 'meta-item__value--success';
    }
    return 'meta-item__value--default';
  }

  private progressText(item: Enrollment): string {
    if (item.status === 'completed') {
      return '100%';
    }

    const progress = Number(item.completion_percent);
    if (!Number.isNaN(progress) && progress > 0) {
      return `${Math.round(progress)}%`;
    }

    if (item.status === 'in_progress') {
      return 'В процессе';
    }
    if (item.status === 'canceled') {
      return 'Остановлено';
    }
    return 'Не начато';
  }

  private activityText(item: Enrollment): string {
    return this.formatDateTime(
      item.last_activity_at ?? item.started_at ?? item.completed_at ?? item.enrolled_at,
    );
  }

  private formatDateTime(value?: string | null): string {
    const date = this.parseDate(value);
    if (!date) {
      return '—';
    }
    return this.shortDateTimeFormatter.format(date).replace('.', '');
  }

  private formatCourseCount(count: number): string {
    const mod10 = count % 10;
    const mod100 = count % 100;

    if (mod10 === 1 && mod100 !== 11) {
      return `${count} курс`;
    }
    if (mod10 >= 2 && mod10 <= 4 && (mod100 < 12 || mod100 > 14)) {
      return `${count} курса`;
    }
    return `${count} курсов`;
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

    const days = Math.ceil((deadline.getTime() - Date.now()) / 86_400_000);
    return days >= 0 && days <= 7;
  }

  private needsAttention(item: Enrollment): boolean {
    return this.isOverdue(item) || this.isDueSoon(item);
  }
}
