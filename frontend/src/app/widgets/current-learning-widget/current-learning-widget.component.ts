import { CommonModule } from '@angular/common';
import { Component, OnInit, computed, inject, signal } from '@angular/core';
import { RouterLink } from '@angular/router';
import { MatButtonModule } from '@angular/material/button';
import { forkJoin } from 'rxjs';

import { CoursesApiService } from '@core/api/courses-api.service';
import { EnrollmentsApiService } from '@core/api/enrollments-api.service';
import type { Course } from '@entities/course';
import type { Enrollment } from '@entities/enrollment';
import { WidgetShellComponent } from '@app/widgets/widget-shell/widget-shell.component';

interface ProcessItem {
  id: string;
  title: string;
  status: string;
  deadlineText: string;
  overdue: boolean;
}

@Component({
  selector: 'app-current-learning-widget',
  standalone: true,
  imports: [CommonModule, RouterLink, MatButtonModule, WidgetShellComponent],
  templateUrl: './current-learning-widget.component.html',
  styleUrl: './current-learning-widget.component.scss',
})
export class CurrentLearningWidgetComponent implements OnInit {
  private readonly enrollmentsApi = inject(EnrollmentsApiService);
  private readonly coursesApi = inject(CoursesApiService);

  protected readonly loading = signal(true);
  protected readonly error = signal<string | null>(null);
  protected readonly items = signal<ProcessItem[]>([]);
  protected readonly selectedTab = signal<string>('all');
  protected readonly tabs = [
    { key: 'all', label: 'Все курсы' },
    { key: 'in_progress', label: 'В процессе' },
    { key: 'overdue', label: 'Просроченные' },
    { key: 'completed', label: 'Завершенные' },
  ];
  protected readonly filteredItems = computed(() => {
    const status = this.selectedTab();
    if (status === 'all') {
      return this.items();
    }
    if (status === 'overdue') {
      return this.items().filter((item) => item.overdue);
    }
    return this.items().filter((item) => item.status === status);
  });

  ngOnInit(): void {
    forkJoin({
      enrollments: this.enrollmentsApi.listMy(),
      courses: this.coursesApi.list({ limit: 300, offset: 0 }),
    }).subscribe({
      next: ({ enrollments, courses }) => {
        this.items.set(this.mapToProcessItems(enrollments ?? [], courses ?? []).slice(0, 8));
        this.loading.set(false);
      },
      error: () => {
        this.error.set('failed');
        this.loading.set(false);
      },
    });
  }

  protected setTab(tab: string): void {
    this.selectedTab.set(tab);
  }

  protected tabCount(tab: string): number {
    if (tab === 'all') {
      return this.items().length;
    }
    if (tab === 'overdue') {
      return this.items().filter((item) => item.overdue).length;
    }
    return this.items().filter((item) => item.status === tab).length;
  }

  private mapToProcessItems(enrollments: Enrollment[], courses: Course[]): ProcessItem[] {
    const courseMap = new Map(courses.map((course) => [course.id, course.title]));

    return enrollments.map((item) => {
      const deadline = this.deadlineText(item.deadline_at);
      const isOverdue =
        item.status !== 'completed' &&
        !!item.deadline_at &&
        !Number.isNaN(new Date(item.deadline_at).getTime()) &&
        new Date(item.deadline_at).getTime() < Date.now();

      return {
        id: item.id,
        title: courseMap.get(item.course_id) || `Курс ${item.course_id.slice(0, 8)}`,
        status: item.status,
        deadlineText: deadline,
        overdue: isOverdue,
      };
    });
  }

  private deadlineText(date: string | null | undefined): string {
    if (!date) {
      return 'Дедлайн не задан';
    }
    const parsed = new Date(date);
    if (Number.isNaN(parsed.getTime())) {
      return 'Дедлайн не задан';
    }
    return `Дедлайн: ${parsed.toLocaleDateString('ru-RU', { day: 'numeric', month: 'long' })}`;
  }
}
