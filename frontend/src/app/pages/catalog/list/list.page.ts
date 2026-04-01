import { CommonModule } from '@angular/common';
import { Component, OnDestroy, OnInit, inject, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule } from '@angular/forms';
import { ActivatedRoute, ParamMap, Router, RouterLink } from '@angular/router';
import { MatButtonModule } from '@angular/material/button';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatIconModule } from '@angular/material/icon';
import { MatSelectModule } from '@angular/material/select';
import { MatTabsModule } from '@angular/material/tabs';
import { Subscription } from 'rxjs';

import { CatalogDeadlinesWidgetComponent } from '@app/widgets/catalog-deadlines-widget/catalog-deadlines-widget.component';
import { CatalogFocusWidgetComponent } from '@app/widgets/catalog-focus-widget/catalog-focus-widget.component';
import { CatalogTracksWidgetComponent } from '@app/widgets/catalog-tracks-widget/catalog-tracks-widget.component';
import { CoursesApiService } from '@core/api/courses-api.service';
import type { Course } from '@entities/course';

const DEFAULT_LIMIT = 50;
const PAGE_SIZE_OPTIONS = [5, 10, 20, 50];
const CATALOG_TAB_PRESETS = [
  { label: 'Все курсы', status: '', source_type: '' },
  { label: 'Внутренние', status: '', source_type: 'internal' },
  { label: 'Внешние', status: '', source_type: 'external' },
  { label: 'Архив', status: 'archived', source_type: '' },
] as const;

type CatalogQueryState = {
  status: string;
  source_type: string;
  level: string;
  limit: number;
  offset: number;
};

@Component({
  selector: 'app-page-catalog-list',
  standalone: true,
  imports: [
    CommonModule,
    ReactiveFormsModule,
    RouterLink,
    MatButtonModule,
    MatFormFieldModule,
    MatIconModule,
    MatSelectModule,
    MatTabsModule,
    CatalogFocusWidgetComponent,
    CatalogTracksWidgetComponent,
    CatalogDeadlinesWidgetComponent,
  ],
  templateUrl: './list.page.html',
  styleUrl: './list.page.scss',
})
export class CatalogListPageComponent implements OnInit, OnDestroy {
  private readonly api = inject(CoursesApiService);
  private readonly fb = inject(FormBuilder);
  private readonly route = inject(ActivatedRoute);
  private readonly router = inject(Router);
  private readonly subscriptions = new Subscription();

  protected readonly loading = signal(true);
  protected readonly error = signal<string | null>(null);
  protected readonly courses = signal<Course[]>([]);
  protected readonly pageSizeOptions = PAGE_SIZE_OPTIONS;
  protected readonly tabPresets = CATALOG_TAB_PRESETS;
  protected readonly activeTabIndex = signal(0);

  protected readonly form = this.fb.group({
    status: [''],
    source_type: [''],
    level: [''],
    limit: [DEFAULT_LIMIT],
    offset: [0],
  });

  protected readonly statusOptions = [
    { value: '', label: 'Все статусы' },
    { value: 'draft', label: 'Черновик' },
    { value: 'published', label: 'Опубликован' },
    { value: 'archived', label: 'Архив' },
  ];

  protected readonly sourceTypeOptions = [
    { value: '', label: 'Все источники' },
    { value: 'internal', label: 'Внутренний' },
    { value: 'external', label: 'Внешний' },
  ];

  protected readonly levelOptions = [
    { value: '', label: 'Все уровни' },
    { value: 'junior', label: 'Junior' },
    { value: 'middle', label: 'Middle' },
    { value: 'senior', label: 'Senior' },
    { value: 'lead', label: 'Lead' },
  ];

  ngOnInit(): void {
    this.subscriptions.add(
      this.route.queryParamMap.subscribe((params) => {
        const state = this.readStateFromQuery(params);

        this.form.patchValue(state, { emitEvent: false });
        this.activeTabIndex.set(this.resolveTabPresetIndex(state));
        this.loadCourses(state);
      }),
    );
  }

  ngOnDestroy(): void {
    this.subscriptions.unsubscribe();
  }

  protected submitFilters(): void {
    const state = this.readStateFromForm();
    this.updateQueryParams({ ...state, offset: 0 });
  }

  protected resetFilters(): void {
    this.updateQueryParams(this.getDefaultState());
  }

  protected selectTabPreset(index: number): void {
    const preset = this.tabPresets[index];

    if (!preset) {
      return;
    }

    const state = this.readStateFromForm();
    this.updateQueryParams({
      ...state,
      status: preset.status,
      source_type: preset.source_type,
      offset: 0,
    });
  }

  protected trackByCourse(_: number, course: Course): string {
    return course.id;
  }

  protected changeLimit(limit: number): void {
    const state = this.readStateFromForm();
    this.updateQueryParams({
      ...state,
      limit: this.normalizeLimit(limit),
      offset: 0,
    });
  }

  protected previousPage(): void {
    const state = this.readStateFromForm();

    if (state.offset <= 0) {
      return;
    }

    this.updateQueryParams({
      ...state,
      offset: Math.max(0, state.offset - state.limit),
    });
  }

  protected nextPage(): void {
    const state = this.readStateFromForm();

    if (!this.canGoNext()) {
      return;
    }

    this.updateQueryParams({
      ...state,
      offset: state.offset + state.limit,
    });
  }

  protected currentPage(): number {
    const state = this.readStateFromForm();
    return Math.floor(state.offset / state.limit) + 1;
  }

  protected canGoPrevious(): boolean {
    return this.readStateFromForm().offset > 0;
  }

  protected canGoNext(): boolean {
    return this.courses().length === this.readStateFromForm().limit;
  }

  protected pageRangeLabel(): string {
    const state = this.readStateFromForm();
    const count = this.courses().length;

    if (count === 0) {
      return 'Курсы не найдены';
    }

    return `${state.offset + 1}-${state.offset + count}`;
  }

  protected formatCourseSubtitle(course: Course): string {
    const sourceType = this.translateSourceType(course.source_type);
    const status = this.normalizeStatus(course.status);

    return [sourceType, status].filter(Boolean).join(' · ');
  }

  protected formatMetaValue(value?: string | null): string {
    const normalized = value?.trim();
    return normalized || '—';
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

  private loadCourses(state: CatalogQueryState): void {
    this.loading.set(true);
    this.error.set(null);

    const query: Record<string, string | number> = {
      limit: state.limit,
      offset: state.offset,
    };

    if (state.status) query['status'] = state.status;
    if (state.source_type) query['source_type'] = state.source_type;
    if (state.level) query['level'] = state.level;

    this.api.list(query).subscribe({
      next: (courses) => {
        this.courses.set(courses ?? []);
        this.loading.set(false);
      },
      error: () => {
        this.error.set('Не удалось загрузить курсы');
        this.loading.set(false);
      },
    });
  }

  private readStateFromForm(): CatalogQueryState {
    const values = this.form.getRawValue();

    return {
      status: values.status ?? '',
      source_type: values.source_type ?? '',
      level: values.level ?? '',
      limit: this.normalizeLimit(values.limit),
      offset: this.normalizeOffset(values.offset),
    };
  }

  private readStateFromQuery(params: ParamMap): CatalogQueryState {
    return {
      status: params.get('status') ?? '',
      source_type: params.get('source_type') ?? '',
      level: params.get('level') ?? '',
      limit: this.normalizeLimit(params.get('limit')),
      offset: this.normalizeOffset(params.get('offset')),
    };
  }

  private updateQueryParams(state: CatalogQueryState): void {
    void this.router.navigate([], {
      relativeTo: this.route,
      queryParams: {
        status: state.status || null,
        source_type: state.source_type || null,
        level: state.level || null,
        limit: state.limit,
        offset: state.offset,
      },
    });
  }

  private getDefaultState(): CatalogQueryState {
    return {
      status: '',
      source_type: '',
      level: '',
      limit: DEFAULT_LIMIT,
      offset: 0,
    };
  }

  private normalizeLimit(value: string | number | null | undefined): number {
    const limit = Math.floor(Number(value));

    if (!Number.isFinite(limit) || limit <= 0) {
      return DEFAULT_LIMIT;
    }

    return Math.min(200, limit);
  }

  private normalizeOffset(value: string | number | null | undefined): number {
    const offset = Math.floor(Number(value));

    if (!Number.isFinite(offset) || offset < 0) {
      return 0;
    }

    return offset;
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

  private resolveTabPresetIndex(state: CatalogQueryState): number {
    const index = this.tabPresets.findIndex((preset) => {
      return preset.status === state.status && preset.source_type === state.source_type;
    });

    return index >= 0 ? index : 0;
  }
}
