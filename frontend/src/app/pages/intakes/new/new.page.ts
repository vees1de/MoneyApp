import { CommonModule } from '@angular/common';
import { Component, OnInit, inject, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { Router, RouterLink } from '@angular/router';
import { MatButtonModule } from '@angular/material/button';
import { MatCardModule } from '@angular/material/card';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatIconModule } from '@angular/material/icon';
import { MatInputModule } from '@angular/material/input';
import { MatSelectModule } from '@angular/material/select';
import { catchError, forkJoin, of } from 'rxjs';

import { CourseIntakesApiService } from '@core/api/course-intakes-api.service';
import { CoursesApiService } from '@core/api/courses-api.service';
import { UsersApiService } from '@core/api/users-api.service';
import { toIsoFromDatetimeLocal } from '@core/domain/date-input.util';
import { identityUserDisplayName } from '@core/domain/identity.util';
import type { IdentityUserView } from '@core/auth/auth.types';
import type { Course } from '@entities/course';

@Component({
  selector: 'app-page-intakes-new',
  standalone: true,
  imports: [
    CommonModule,
    ReactiveFormsModule,
    RouterLink,
    MatButtonModule,
    MatCardModule,
    MatFormFieldModule,
    MatIconModule,
    MatInputModule,
    MatSelectModule,
  ],
  templateUrl: './new.page.html',
  styleUrl: './new.page.scss',
})
export class IntakesNewPageComponent implements OnInit {
  private readonly intakesApi = inject(CourseIntakesApiService);
  private readonly coursesApi = inject(CoursesApiService);
  private readonly usersApi = inject(UsersApiService);
  private readonly fb = inject(FormBuilder);
  private readonly router = inject(Router);

  protected readonly loadingOptions = signal(true);
  protected readonly submitting = signal(false);
  protected readonly error = signal<string | null>(null);
  protected readonly directoryUnavailable = signal(false);
  protected readonly courses = signal<Course[]>([]);
  protected readonly approvers = signal<IdentityUserView[]>([]);

  protected readonly form = this.fb.group({
    source_mode: ['catalog', [Validators.required]],
    course_id: [''],
    title: ['', [Validators.required]],
    description: [''],
    approver_id: [''],
    max_participants: [''],
    start_date: [''],
    end_date: [''],
    application_deadline: [''],
  });

  ngOnInit(): void {
    forkJoin({
      courses: this.coursesApi.list({ limit: 100 }).pipe(catchError(() => of([]))),
      approvers: this.usersApi.listAdminUsers().pipe(
        catchError(() => {
          this.directoryUnavailable.set(true);
          return of([]);
        }),
      ),
    }).subscribe({
      next: ({ courses, approvers }) => {
        this.courses.set(courses ?? []);
        this.approvers.set(approvers ?? []);
        this.loadingOptions.set(false);
      },
      error: () => {
        this.error.set('Не удалось подготовить форму набора.');
        this.loadingOptions.set(false);
      },
    });
  }

  protected isCatalogMode(): boolean {
    return this.form.controls.source_mode.value === 'catalog';
  }

  protected approverLabel(user: IdentityUserView): string {
    return identityUserDisplayName(user);
  }

  protected syncSelectedCourse(courseId: string | null): void {
    if (!courseId || !this.isCatalogMode()) {
      return;
    }

    const course = this.courses().find((item) => item.id === courseId);
    if (!course) {
      return;
    }

    this.form.patchValue({
      title: course.title,
      description: course.description ?? course.short_description ?? '',
    });
  }

  protected async submit(): Promise<void> {
    if (this.form.invalid || this.submitting()) {
      this.form.markAllAsTouched();
      return;
    }

    const values = this.form.getRawValue();
    const selectedCourse = this.courses().find((item) => item.id === values.course_id);
    const title = normalizeText(values.title) ?? selectedCourse?.title;

    if (!title) {
      this.error.set('Укажите курс из каталога или введите название набора.');
      return;
    }

    this.submitting.set(true);
    this.error.set(null);

    const payload: Record<string, unknown> = {
      course_id: this.isCatalogMode() ? normalizeText(values.course_id) : undefined,
      title,
      description:
        normalizeText(values.description) ??
        selectedCourse?.description ??
        selectedCourse?.short_description ??
        undefined,
      approver_id: normalizeText(values.approver_id),
      max_participants: toPositiveNumber(values.max_participants),
      start_date: normalizeText(values.start_date),
      end_date: normalizeText(values.end_date),
      application_deadline: toIsoFromDatetimeLocal(values.application_deadline),
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

function normalizeText(value: string | null | undefined): string | undefined {
  const trimmed = value?.trim();
  return trimmed ? trimmed : undefined;
}

function toPositiveNumber(value: string | null | undefined): number | undefined {
  if (!value) {
    return undefined;
  }

  const parsed = Number(value);
  return Number.isFinite(parsed) && parsed > 0 ? parsed : undefined;
}
