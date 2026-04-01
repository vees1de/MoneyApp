import { CommonModule } from '@angular/common';
import { Component, computed, inject, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { Router, RouterLink } from '@angular/router';
import { MatButtonModule } from '@angular/material/button';
import { MatButtonToggleModule } from '@angular/material/button-toggle';
import { MatCardModule } from '@angular/material/card';
import { MatDialog } from '@angular/material/dialog';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatIconModule } from '@angular/material/icon';
import { MatInputModule } from '@angular/material/input';

import { CourseIntakesApiService } from '@core/api/course-intakes-api.service';
import {
  buildIntakeSchedulePayload,
  defaultDeadlineFromStartDate,
  normalizeText,
  resolveEndDateFromWeeks,
  toPositiveNumber,
} from '@core/domain/course-intake-form.util';
import type { Course } from '@entities/course';

import { CoursePickerDialogComponent } from '../course-picker-dialog/course-picker-dialog.component';

@Component({
  selector: 'app-page-intakes-new',
  standalone: true,
  imports: [
    CommonModule,
    ReactiveFormsModule,
    RouterLink,
    MatButtonModule,
    MatButtonToggleModule,
    MatCardModule,
    MatFormFieldModule,
    MatIconModule,
    MatInputModule,
  ],
  templateUrl: './new.page.html',
  styleUrl: './new.page.scss',
})
export class IntakesNewPageComponent {
  private readonly intakesApi = inject(CourseIntakesApiService);
  private readonly fb = inject(FormBuilder);
  private readonly router = inject(Router);
  private readonly dialog = inject(MatDialog);

  protected readonly submitting = signal(false);
  protected readonly error = signal<string | null>(null);
  protected readonly selectedCourse = signal<Course | null>(null);
  protected readonly standardParticipants = 10;
  protected readonly defaultStartDate = this.buildDefaultStartDate();
  protected readonly defaultStartDateLabel = this.formatDate(this.defaultStartDate);
  private autoDeadline = defaultDeadlineFromStartDate(this.defaultStartDate) ?? '';

  protected readonly form = this.fb.group({
    source_mode: ['catalog', [Validators.required]],
    course_id: [''],
    title: ['', [Validators.required]],
    description: [''],
    max_participants: [
      String(this.standardParticipants),
      [Validators.required, Validators.min(1)],
    ],
    price: [''],
    price_currency: ['RUB'],
    start_date: [this.defaultStartDate, [Validators.required]],
    schedule_mode: ['weeks', [Validators.required]],
    duration_weeks: [''],
    end_date: [''],
    application_deadline: [this.autoDeadline],
  });

  protected readonly isCatalogMode = computed(
    () => this.form.controls.source_mode.value === 'catalog',
  );
  protected readonly isWeeksMode = computed(
    () => this.form.controls.schedule_mode.value === 'weeks',
  );
  protected readonly calculatedEndDate = computed(() =>
    resolveEndDateFromWeeks(
      this.form.controls.start_date.value,
      toPositiveNumber(this.form.controls.duration_weeks.value) ?? null,
    ),
  );
  protected readonly summary = computed(() => {
    const participants =
      toPositiveNumber(this.form.controls.max_participants.value) ?? this.standardParticipants;
    const startDate = normalizeText(this.form.controls.start_date.value) ?? this.defaultStartDate;
    const endDate = this.isWeeksMode()
      ? this.calculatedEndDate() ?? 'будет рассчитана после ввода длительности'
      : normalizeText(this.form.controls.end_date.value) ?? 'не указана';
    const applicationDeadline =
      normalizeText(this.form.controls.application_deadline.value) ??
      'автоматически: за 3 дня до старта';

    return {
      participants,
      startDate,
      endDate,
      applicationDeadline,
    };
  });

  constructor() {
    this.form.controls.source_mode.valueChanges.subscribe((mode) => {
      if (mode === 'custom') {
        this.selectedCourse.set(null);
        this.form.patchValue({ course_id: '' }, { emitEvent: false });
      }
    });

    this.form.controls.start_date.valueChanges.subscribe((startDate) => {
      const calculatedDeadline = defaultDeadlineFromStartDate(startDate) ?? '';
      const currentDeadline = this.form.controls.application_deadline.value ?? '';

      if (!normalizeText(currentDeadline) || currentDeadline === this.autoDeadline) {
        this.form.controls.application_deadline.setValue(calculatedDeadline, { emitEvent: false });
      }

      this.autoDeadline = calculatedDeadline;
    });
  }

  protected formatDate(value: string | null | undefined): string {
    const normalized = normalizeText(value);
    if (!normalized) {
      return '—';
    }
    if (!/^\d{4}-\d{2}-\d{2}$/.test(normalized)) {
      return normalized;
    }

    const date = new Date(`${normalized}T00:00:00`);
    if (Number.isNaN(date.getTime())) {
      return normalized;
    }

    return new Intl.DateTimeFormat('ru-RU', {
      day: '2-digit',
      month: '2-digit',
      year: 'numeric',
    }).format(date);
  }

  protected formatDateTime(value: string | null | undefined): string {
    const normalized = normalizeText(value);
    if (!normalized) {
      return '—';
    }

    const date = new Date(normalized);
    if (Number.isNaN(date.getTime())) {
      return normalized;
    }

    return new Intl.DateTimeFormat('ru-RU', {
      day: '2-digit',
      month: '2-digit',
      year: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    }).format(date);
  }

  protected openCatalogDialog(): void {
    const dialogRef = this.dialog.open(CoursePickerDialogComponent, {
      width: '960px',
      maxWidth: '96vw',
      data: {
        selectedCourseId: this.selectedCourse()?.id ?? this.form.controls.course_id.value,
      },
    });

    dialogRef.afterClosed().subscribe((course) => {
      if (!course) {
        return;
      }

      this.selectedCourse.set(course);
      this.form.patchValue({
        course_id: course.id,
        title: course.title,
        description: course.description ?? course.short_description ?? '',
        price: course.price ?? '',
        price_currency: course.price_currency ?? 'RUB',
      });
    });
  }

  protected clearCatalogSelection(): void {
    this.selectedCourse.set(null);
    this.form.patchValue({
      course_id: '',
      title: '',
      description: '',
      price: '',
      price_currency: 'RUB',
    });
  }

  protected async submit(): Promise<void> {
    if (this.form.invalid || this.submitting()) {
      this.form.markAllAsTouched();
      return;
    }

    const values = this.form.getRawValue();
    const selectedCourse = this.selectedCourse();
    const title = normalizeText(values.title) ?? selectedCourse?.title;

    if (this.isCatalogMode() && !normalizeText(values.course_id)) {
      this.error.set('Выберите курс из каталога.');
      return;
    }
    if (!title) {
      this.error.set('Укажите название набора.');
      return;
    }
    if (this.isWeeksMode() && !toPositiveNumber(values.duration_weeks)) {
      this.error.set('Укажите количество недель.');
      return;
    }
    if (!this.isWeeksMode() && !normalizeText(values.end_date)) {
      this.error.set('Укажите дату окончания курса.');
      return;
    }

    this.submitting.set(true);
    this.error.set(null);

    const payload: Record<string, unknown> = {
      course_id: this.isCatalogMode() ? normalizeText(values.course_id) : undefined,
      title,
      description: normalizeText(values.description),
      max_participants: toPositiveNumber(values.max_participants),
      price: values.price ?? '',
      price_currency: normalizeText(values.price_currency) ?? 'RUB',
      ...buildIntakeSchedulePayload({
        startDate: values.start_date,
        endDate: this.isWeeksMode() ? this.calculatedEndDate() : values.end_date,
        durationWeeks: this.isWeeksMode() ? toPositiveNumber(values.duration_weeks) : undefined,
        applicationDeadline: values.application_deadline,
      }),
    };

    this.intakesApi.create(payload).subscribe({
      next: async (created) => {
        await this.router.navigateByUrl(`/intakes/${created.id}`);
        this.submitting.set(false);
      },
      error: () => {
        this.error.set('Не удалось открыть набор.');
        this.submitting.set(false);
      },
    });
  }

  private buildDefaultStartDate(): string {
    const date = new Date();
    date.setHours(0, 0, 0, 0);
    date.setDate(date.getDate() + 7);
    return IntakesNewPageComponent.toDateInputValue(date);
  }

  private static toDateInputValue(date: Date): string {
    const year = date.getFullYear();
    const month = `${date.getMonth() + 1}`.padStart(2, '0');
    const day = `${date.getDate()}`.padStart(2, '0');
    return `${year}-${month}-${day}`;
  }
}
