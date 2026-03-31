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

  protected readonly form = this.fb.group({
    source_mode: ['catalog', [Validators.required]],
    course_id: [''],
    title: ['', [Validators.required]],
    description: [''],
    max_participants: [''],
    price: [''],
    price_currency: ['RUB'],
    start_date: ['', [Validators.required]],
    schedule_mode: ['weeks', [Validators.required]],
    duration_weeks: [''],
    end_date: [''],
    application_deadline: [''],
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

  constructor() {
    this.form.controls.source_mode.valueChanges.subscribe((mode) => {
      if (mode === 'custom') {
        this.selectedCourse.set(null);
        this.form.patchValue({ course_id: '' }, { emitEvent: false });
      }
    });
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
}
