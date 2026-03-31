import { CommonModule } from '@angular/common';
import { Component, OnInit, computed, inject, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule } from '@angular/forms';
import { ActivatedRoute, Router, RouterLink } from '@angular/router';
import { MatButtonModule } from '@angular/material/button';
import { MatCardModule } from '@angular/material/card';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatIconModule } from '@angular/material/icon';
import { MatInputModule } from '@angular/material/input';
import { MatSelectModule } from '@angular/material/select';
import { catchError, forkJoin, of } from 'rxjs';

import { CourseSuggestionsApiService } from '@core/api/course-suggestions-api.service';
import { CoursesApiService } from '@core/api/courses-api.service';
import { UsersApiService } from '@core/api/users-api.service';
import type { CourseSuggestion } from '@core/api/contracts';
import { AuthStateService } from '@core/auth/auth-state.service';
import { PERMISSIONS } from '@core/auth/permissions';
import { toDatetimeLocalValue, toIsoFromDatetimeLocal } from '@core/domain/date-input.util';
import {
  canOpenIntakeFromSuggestion,
  canReviewSuggestion,
  courseSuggestionStatusLabel,
} from '@core/domain/course-intakes.workflow';
import { identityUserDisplayName } from '@core/domain/identity.util';
import type { IdentityUserView } from '@core/auth/auth.types';
import type { Course } from '@entities/course';

@Component({
  selector: 'app-page-suggestion-detail',
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
  templateUrl: './detail.page.html',
  styleUrl: './detail.page.scss',
})
export class SuggestionDetailPageComponent implements OnInit {
  private readonly suggestionsApi = inject(CourseSuggestionsApiService);
  private readonly coursesApi = inject(CoursesApiService);
  private readonly usersApi = inject(UsersApiService);
  private readonly authState = inject(AuthStateService);
  private readonly route = inject(ActivatedRoute);
  private readonly router = inject(Router);
  private readonly fb = inject(FormBuilder);

  protected readonly loading = signal(true);
  protected readonly acting = signal(false);
  protected readonly error = signal<string | null>(null);
  protected readonly directoryUnavailable = signal(false);
  protected readonly suggestion = signal<CourseSuggestion | null>(null);
  protected readonly courses = signal<Course[]>([]);
  protected readonly users = signal<IdentityUserView[]>([]);
  protected readonly usersById = signal<Record<string, IdentityUserView>>({});

  protected readonly canReview = computed(() => {
    const role = this.authState.currentUser()?.roles[0] ?? 'employee';
    return canReviewSuggestion(role) || this.authState.hasPermission(PERMISSIONS.intakesManage);
  });

  protected readonly reviewForm = this.fb.group({
    comment: [''],
  });

  protected readonly intakeForm = this.fb.group({
    course_id: [''],
    title: [''],
    description: [''],
    approver_id: [''],
    max_participants: [''],
    start_date: [''],
    end_date: [''],
    application_deadline: [''],
  });

  ngOnInit(): void {
    this.load();
  }

  protected statusLabel(status: string): string {
    return courseSuggestionStatusLabel(status);
  }

  protected userLabel(userId: string | null | undefined): string {
    if (!userId) {
      return '—';
    }

    const currentUser = this.authState.currentUser();
    if (currentUser?.id === userId) {
      return identityUserDisplayName(currentUser);
    }

    const user = this.usersById()[userId];
    return user ? identityUserDisplayName(user) : `Сотрудник ${userId.slice(0, 8)}`;
  }

  protected canOpenIntake(): boolean {
    const suggestion = this.suggestion();
    return !!suggestion && this.canReview() && canOpenIntakeFromSuggestion(suggestion.status);
  }

  protected syncSelectedCourse(courseId: string | null): void {
    if (!courseId) {
      return;
    }

    const course = this.courses().find((item) => item.id === courseId);
    if (!course) {
      return;
    }

    this.intakeForm.patchValue({
      title: course.title,
      description: course.description ?? course.short_description ?? '',
    });
  }

  protected approve(): void {
    const suggestion = this.suggestion();
    if (!suggestion || this.acting()) {
      return;
    }

    this.acting.set(true);
    this.error.set(null);

    this.suggestionsApi
      .approve(suggestion.id, { comment: normalizeText(this.reviewForm.controls.comment.value) })
      .subscribe({
        next: (updated) => {
          this.suggestion.set(updated);
          this.acting.set(false);
        },
        error: () => {
          this.error.set('Не удалось одобрить предложение.');
          this.acting.set(false);
        },
      });
  }

  protected reject(): void {
    const suggestion = this.suggestion();
    if (!suggestion || this.acting()) {
      return;
    }

    this.acting.set(true);
    this.error.set(null);

    this.suggestionsApi
      .reject(suggestion.id, { comment: normalizeText(this.reviewForm.controls.comment.value) })
      .subscribe({
        next: (updated) => {
          this.suggestion.set(updated);
          this.acting.set(false);
        },
        error: () => {
          this.error.set('Не удалось отклонить предложение.');
          this.acting.set(false);
        },
      });
  }

  protected async openIntake(): Promise<void> {
    const suggestion = this.suggestion();
    if (!suggestion || this.acting()) {
      return;
    }

    const values = this.intakeForm.getRawValue();
    const selectedCourse = this.courses().find((item) => item.id === values.course_id);

    this.acting.set(true);
    this.error.set(null);

    this.suggestionsApi
      .openIntake(suggestion.id, {
        course_id: normalizeText(values.course_id),
        title: normalizeText(values.title) ?? selectedCourse?.title ?? suggestion.title,
        description:
          normalizeText(values.description) ?? suggestion.description ?? selectedCourse?.description ?? undefined,
        approver_id: normalizeText(values.approver_id) ?? suggestion.approver_id ?? undefined,
        max_participants: toPositiveNumber(values.max_participants),
        start_date: normalizeText(values.start_date),
        end_date: normalizeText(values.end_date),
        application_deadline: toIsoFromDatetimeLocal(values.application_deadline),
      })
      .subscribe({
        next: async (response) => {
          this.suggestion.set(response.suggestion);
          await this.router.navigateByUrl(`/intakes/${response.intake.id}`);
          this.acting.set(false);
        },
        error: () => {
          this.error.set('Не удалось открыть набор по предложению.');
          this.acting.set(false);
        },
      });
  }

  private load(): void {
    const suggestionId = this.route.snapshot.paramMap.get('suggestionId');
    if (!suggestionId) {
      this.error.set('Не найден suggestionId.');
      this.loading.set(false);
      return;
    }

    this.loading.set(true);
    this.error.set(null);

    this.suggestionsApi.getById(suggestionId).subscribe({
      next: (suggestion) => {
        this.suggestion.set(suggestion);
        this.intakeForm.patchValue({
          title: suggestion.title,
          description: suggestion.description ?? '',
          approver_id: suggestion.approver_id ?? '',
        });

        if (!this.canReview()) {
          this.loading.set(false);
          return;
        }

        forkJoin({
          courses: this.coursesApi.list({ limit: 100 }).pipe(catchError(() => of([]))),
          users: this.usersApi.listAdminUsers().pipe(
            catchError(() => {
              this.directoryUnavailable.set(true);
              return of([]);
            }),
          ),
        }).subscribe({
          next: ({ courses, users }) => {
            const usersById = (users ?? []).reduce<Record<string, IdentityUserView>>(
              (accumulator, user) => {
                accumulator[user.id] = user;
                return accumulator;
              },
              {},
            );

            this.courses.set(courses ?? []);
            this.users.set(users ?? []);
            this.usersById.set(usersById);
            this.loading.set(false);
          },
          error: () => {
            this.error.set('Не удалось подготовить справочники для HR.');
            this.loading.set(false);
          },
        });
      },
      error: () => {
        this.error.set('Не удалось загрузить предложение.');
        this.loading.set(false);
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
