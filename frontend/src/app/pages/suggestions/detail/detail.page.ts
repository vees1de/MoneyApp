import { CommonModule } from '@angular/common';
import { Component, OnInit, computed, inject, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule } from '@angular/forms';
import { ActivatedRoute, Router, RouterLink } from '@angular/router';
import { MatButtonModule } from '@angular/material/button';
import { MatButtonToggleModule } from '@angular/material/button-toggle';
import { MatCardModule } from '@angular/material/card';
import { MatDialog } from '@angular/material/dialog';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatIconModule } from '@angular/material/icon';
import { MatInputModule } from '@angular/material/input';
import { catchError, forkJoin, of } from 'rxjs';

import { CourseSuggestionsApiService } from '@core/api/course-suggestions-api.service';
import { UsersApiService } from '@core/api/users-api.service';
import type { CourseSuggestion } from '@core/api/contracts';
import { AuthStateService } from '@core/auth/auth-state.service';
import { PERMISSIONS } from '@core/auth/permissions';
import {
  buildIntakeSchedulePayload,
  normalizeText,
  resolveEndDateFromWeeks,
  toPositiveNumber,
} from '@core/domain/course-intake-form.util';
import {
  canOpenIntakeFromSuggestion,
  canReviewSuggestion,
  courseSuggestionStatusLabel,
} from '@core/domain/course-intakes.workflow';
import { identityUserDisplayName } from '@core/domain/identity.util';
import type { IdentityUserView } from '@core/auth/auth.types';
import type { Course } from '@entities/course';

import { CoursePickerDialogComponent } from '../../intakes/course-picker-dialog/course-picker-dialog.component';
import { toSignal } from '@angular/core/rxjs-interop';

@Component({
  selector: 'app-page-suggestion-detail',
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
  templateUrl: './detail.page.html',
  styleUrl: './detail.page.scss',
})
export class SuggestionDetailPageComponent implements OnInit {
  private readonly suggestionsApi = inject(CourseSuggestionsApiService);
  private readonly usersApi = inject(UsersApiService);
  private readonly authState = inject(AuthStateService);
  private readonly route = inject(ActivatedRoute);
  private readonly router = inject(Router);
  private readonly dialog = inject(MatDialog);
  private readonly fb = inject(FormBuilder);

  protected readonly loading = signal(true);
  protected readonly acting = signal(false);
  protected readonly error = signal<string | null>(null);
  protected readonly directoryUnavailable = signal(false);
  protected readonly suggestion = signal<CourseSuggestion | null>(null);
  protected readonly selectedCourse = signal<Course | null>(null);
  protected readonly users = signal<IdentityUserView[]>([]);
  protected readonly usersById = signal<Record<string, IdentityUserView>>({});

  protected readonly reviewForm = this.fb.group({
    comment: [''],
  });

  protected readonly intakeForm = this.fb.group({
    course_id: [''],
    title: [''],
    description: [''],
    max_participants: [''],
    price: [''],
    price_currency: ['RUB'],
    start_date: [''],
    schedule_mode: ['weeks'],
    duration_weeks: [''],
    end_date: [''],
    application_deadline: [''],
  });

  protected readonly canReview = computed(() => {
    const role = this.authState.currentUser()?.roles[0] ?? 'employee';
    return canReviewSuggestion(role) || this.authState.hasPermission(PERMISSIONS.intakesManage);
  });
  schedule_mode_signal = toSignal(this.intakeForm.controls.schedule_mode.valueChanges);
  protected readonly isWeeksMode = computed(() => this.schedule_mode_signal() === 'weeks');
  protected readonly calculatedEndDate = computed(() =>
    resolveEndDateFromWeeks(
      this.intakeForm.controls.start_date.value,
      toPositiveNumber(this.intakeForm.controls.duration_weeks.value) ?? null,
    ),
  );

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

  protected openCatalogDialog(): void {
    const dialogRef = this.dialog.open(CoursePickerDialogComponent, {
      width: '960px',
      maxWidth: '96vw',
      data: {
        selectedCourseId: this.selectedCourse()?.id ?? this.intakeForm.controls.course_id.value,
      },
    });

    dialogRef.afterClosed().subscribe((course) => {
      if (!course) {
        return;
      }

      this.selectedCourse.set(course);
      this.intakeForm.patchValue({
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
    const suggestion = this.suggestion();
    this.intakeForm.patchValue({
      course_id: '',
      title: suggestion?.title ?? '',
      description: suggestion?.description ?? '',
      price: suggestion?.price ?? '',
      price_currency: suggestion?.price_currency ?? 'RUB',
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
    const selectedCourse = this.selectedCourse();
    if (this.isWeeksMode() && !toPositiveNumber(values.duration_weeks)) {
      this.error.set('Укажите количество недель.');
      return;
    }
    if (!this.isWeeksMode() && !normalizeText(values.end_date)) {
      this.error.set('Укажите дату окончания.');
      return;
    }

    this.acting.set(true);
    this.error.set(null);

    this.suggestionsApi
      .openIntake(suggestion.id, {
        course_id: normalizeText(values.course_id),
        title: normalizeText(values.title) ?? selectedCourse?.title ?? suggestion.title,
        description: normalizeText(values.description),
        max_participants: toPositiveNumber(values.max_participants),
        price: values.price ?? '',
        price_currency: normalizeText(values.price_currency) ?? suggestion.price_currency ?? 'RUB',
        ...buildIntakeSchedulePayload({
          startDate: values.start_date,
          endDate: this.isWeeksMode() ? this.calculatedEndDate() : values.end_date,
          durationWeeks: this.isWeeksMode() ? toPositiveNumber(values.duration_weeks) : undefined,
          applicationDeadline: values.application_deadline,
        }),
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
          price: suggestion.price ?? '',
          price_currency: suggestion.price_currency ?? 'RUB',
        });

        if (!this.canReview()) {
          this.loading.set(false);
          return;
        }

        forkJoin({
          users: this.usersApi.listAdminUsers().pipe(
            catchError(() => {
              this.directoryUnavailable.set(true);
              return of([]);
            }),
          ),
        }).subscribe({
          next: ({ users }) => {
            const usersById = (users ?? []).reduce<Record<string, IdentityUserView>>(
              (accumulator, user) => {
                accumulator[user.id] = user;
                return accumulator;
              },
              {},
            );

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
