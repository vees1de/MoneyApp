import { CommonModule } from '@angular/common';
import { Component, OnInit, inject, signal } from '@angular/core';
import { ActivatedRoute, RouterLink } from '@angular/router';
import { MatButtonModule } from '@angular/material/button';
import { MatCardModule } from '@angular/material/card';
import { MatIconModule } from '@angular/material/icon';

import { CoursesApiService } from '@core/api/courses-api.service';
import type { Course } from '@entities/course';

@Component({
  selector: 'app-page-catalog-detail',
  standalone: true,
  imports: [CommonModule, RouterLink, MatButtonModule, MatCardModule, MatIconModule],
  templateUrl: './detail.page.html',
  styleUrl: './detail.page.scss',
})
export class CatalogDetailPageComponent implements OnInit {
  private readonly coursesApi = inject(CoursesApiService);
  private readonly route = inject(ActivatedRoute);

  protected readonly loading = signal(true);
  protected readonly error = signal<string | null>(null);
  protected readonly course = signal<Course | null>(null);

  ngOnInit(): void {
    const id = this.route.snapshot.paramMap.get('courseId');
    if (!id) {
      this.error.set('Не найден идентификатор курса.');
      this.loading.set(false);
      return;
    }

    this.coursesApi.getById(id).subscribe({
      next: (course) => {
        this.course.set(course);
        this.loading.set(false);
      },
      error: () => {
        this.error.set('Не удалось загрузить курс.');
        this.loading.set(false);
      },
    });
  }

  protected formatCourseSubtitle(course: Course): string {
    const sourceType = this.translateSourceType(course.source_type);
    const status = this.normalizeStatus(course.status);

    return [sourceType, status].filter(Boolean).join(' · ');
  }

  protected formatDurationHours(value?: string | null): string {
    const normalized = value?.trim();

    if (!normalized) {
      return '—';
    }

    const numericValue = Number(normalized);

    if (!Number.isFinite(numericValue)) {
      return normalized;
    }

    return String(numericValue);
  }

  protected formatDurationLabel(value?: string | null): string {
    const hours = this.formatDurationHours(value);
    return hours === '—' ? hours : `${hours} ч`;
  }

  protected formatPrice(course: Course): string {
    const price = course.price?.trim();
    const currency = course.price_currency?.trim();

    if (!price) {
      return '—';
    }

    if (currency === 'RUB' && Number(price) === 0) {
      return 'бесплатно';
    }

    return [price, currency].filter(Boolean).join(' ');
  }

  protected suggestionQueryParams(course: Course): Record<string, string | null> {
    return {
      title: course.title,
      external_url: course.external_url?.trim() || null,
      description: course.description?.trim() || course.short_description?.trim() || null,
      price: course.price?.trim() || null,
      price_currency: course.price_currency?.trim() || 'RUB',
      duration_hours: this.normalizeSuggestionDuration(course.duration_hours),
      from: 'catalog',
    };
  }

  private translateSourceType(value?: string | null): string {
    const normalized = value?.trim().toLowerCase();

    switch (normalized) {
      case 'imported':
        return 'Импортирован';
      case 'catalog':
        return 'Каталог';
      case 'internal':
        return 'Внутренний';
      case 'external':
        return 'Внешний';
      default:
        return value?.trim() ?? '';
    }
  }

  private normalizeStatus(value?: string | null): string {
    const normalized = value?.trim().toLowerCase();

    if (!normalized || normalized === 'published') {
      return '';
    }

    switch (normalized) {
      case 'draft':
        return 'Черновик';
      case 'archived':
        return 'Архив';
      default:
        return value?.trim() ?? '';
    }
  }

  private normalizeSuggestionDuration(value?: string | null): string | null {
    const normalized = value?.trim();

    if (!normalized) {
      return null;
    }

    const numericValue = Number(normalized);
    return Number.isFinite(numericValue) ? String(numericValue) : normalized;
  }
}
