import { CommonModule } from '@angular/common';
import { Component, OnInit, computed, inject, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { ActivatedRoute, RouterLink } from '@angular/router';
import { MatButtonModule } from '@angular/material/button';
import { MatCardModule } from '@angular/material/card';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatIconModule } from '@angular/material/icon';
import { MatInputModule } from '@angular/material/input';
import { MatSelectModule } from '@angular/material/select';
import { catchError, forkJoin, of } from 'rxjs';

import { CourseApplicationsApiService } from '@core/api/course-applications-api.service';
import { CourseIntakesApiService } from '@core/api/course-intakes-api.service';
import { CoursesApiService } from '@core/api/courses-api.service';
import { UsersApiService } from '@core/api/users-api.service';
import type { CourseApplication, CourseIntake } from '@core/api/contracts';
import { AuthStateService } from '@core/auth/auth-state.service';
import type { IdentityUserView, RoleCode } from '@core/auth/auth.types';
import { PERMISSIONS } from '@core/auth/permissions';
import { toDatetimeLocalValue, toIsoFromDatetimeLocal } from '@core/domain/date-input.util';
import {
  canApplyToIntake,
  canEnrollApplication,
  canHrReviewApplication,
  canWithdrawApplication,
  courseApplicationStatusLabel,
  courseIntakeStatusLabel,
  isIntakeManageRole,
} from '@core/domain/course-intakes.workflow';
import { identityUserDisplayName } from '@core/domain/identity.util';
import type { Course } from '@entities/course';

@Component({
  selector: 'app-page-intake-detail',
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
export class IntakeDetailPageComponent implements OnInit {
  private readonly intakesApi = inject(CourseIntakesApiService);
  private readonly applicationsApi = inject(CourseApplicationsApiService);
  private readonly coursesApi = inject(CoursesApiService);
  private readonly usersApi = inject(UsersApiService);
  private readonly authState = inject(AuthStateService);
  private readonly route = inject(ActivatedRoute);
  private readonly fb = inject(FormBuilder);

  protected readonly loading = signal(true);
  protected readonly acting = signal(false);
  protected readonly error = signal<string | null>(null);
  protected readonly directoryUnavailable = signal(false);
  protected readonly intake = signal<CourseIntake | null>(null);
  protected readonly course = signal<Course | null>(null);
  protected readonly applications = signal<CourseApplication[]>([]);
  protected readonly myApplication = signal<CourseApplication | null>(null);
  protected readonly approvers = signal<IdentityUserView[]>([]);
  protected readonly applicationComments = signal<Record<string, string>>({});

  protected readonly role = computed<RoleCode>(
    () => this.authState.currentUser()?.roles[0] ?? 'employee',
  );
  protected readonly canManage = computed(
    () =>
      isIntakeManageRole(this.role()) || this.authState.hasPermission(PERMISSIONS.intakesManage),
  );
  protected readonly hasDirectoryAccess = computed(
    () =>
      this.authState.hasPermission(PERMISSIONS.usersRead) || this.authState.hasAnyRole(['hr', 'admin']),
  );
  protected readonly canApply = computed(() => {
    const item = this.intake();
    return !!item && canApplyToIntake(item.status, !!this.myApplication());
  });
  protected readonly canWithdrawMyApplication = computed(() => {
    const application = this.myApplication();
    return !!application && canWithdrawApplication(application.status);
  });

  protected readonly statusOptions = [
    { value: 'open', label: 'Открыт' },
    { value: 'closed', label: 'Закрыт' },
    { value: 'canceled', label: 'Отменён' },
    { value: 'completed', label: 'Завершён' },
  ];

  protected readonly editForm = this.fb.group({
    title: ['', [Validators.required]],
    description: [''],
    approver_id: [''],
    max_participants: [''],
    start_date: [''],
    end_date: [''],
    application_deadline: [''],
    status: ['open', [Validators.required]],
  });

  protected readonly applyForm = this.fb.group({
    motivation: [''],
  });

  ngOnInit(): void {
    this.load();
  }

  protected intakeStatusLabel(status: string): string {
    return courseIntakeStatusLabel(status);
  }

  protected applicationStatusLabel(status: string): string {
    return courseApplicationStatusLabel(status);
  }

  protected approverLabel(user: IdentityUserView): string {
    return identityUserDisplayName(user);
  }

  protected userLabel(userId: string | null | undefined): string {
    if (!userId) {
      return '—';
    }

    const currentUser = this.authState.currentUser();
    if (currentUser?.id === userId) {
      return identityUserDisplayName(currentUser);
    }

    const match = this.approvers().find((item) => item.id === userId);
    return match ? identityUserDisplayName(match) : `Сотрудник ${userId.slice(0, 8)}`;
  }

  protected applicationComment(id: string): string {
    return this.applicationComments()[id] ?? '';
  }

  protected setApplicationComment(id: string, value: string): void {
    this.applicationComments.update((state) => ({
      ...state,
      [id]: value,
    }));
  }

  protected canHrApprove(application: CourseApplication): boolean {
    return canHrReviewApplication(application.status);
  }

  protected canHrEnroll(application: CourseApplication): boolean {
    return canEnrollApplication(application.status);
  }

  protected saveIntake(): void {
    const item = this.intake();
    if (!item || this.editForm.invalid || this.acting()) {
      this.editForm.markAllAsTouched();
      return;
    }

    const values = this.editForm.getRawValue();
    const title = normalizeText(values.title);
    if (!title) {
      this.error.set('Название набора обязательно.');
      return;
    }

    this.acting.set(true);
    this.error.set(null);

    this.intakesApi
      .update(item.id, {
        title,
        description: normalizeText(values.description),
        approver_id: normalizeText(values.approver_id),
        max_participants: toPositiveNumber(values.max_participants),
        start_date: normalizeText(values.start_date),
        end_date: normalizeText(values.end_date),
        application_deadline: toIsoFromDatetimeLocal(values.application_deadline),
        status: normalizeText(values.status),
      })
      .subscribe({
        next: (updated) => {
          this.applyIntakeState(updated);
          this.acting.set(false);
        },
        error: () => {
          this.error.set('Не удалось обновить набор.');
          this.acting.set(false);
        },
      });
  }

  protected closeIntake(): void {
    const item = this.intake();
    if (!item || this.acting()) {
      return;
    }

    this.acting.set(true);
    this.error.set(null);

    this.intakesApi.close(item.id).subscribe({
      next: (updated) => {
        this.applyIntakeState(updated);
        this.acting.set(false);
      },
      error: () => {
        this.error.set('Не удалось закрыть набор.');
        this.acting.set(false);
      },
    });
  }

  protected applyToIntake(): void {
    const item = this.intake();
    if (!item || this.acting()) {
      return;
    }

    this.acting.set(true);
    this.error.set(null);

    this.applicationsApi
      .apply({
        intake_id: item.id,
        motivation: normalizeText(this.applyForm.controls.motivation.value),
      })
      .subscribe({
        next: (application) => {
          this.myApplication.set(application);
          this.acting.set(false);
        },
        error: () => {
          this.error.set('Не удалось подать заявку в набор.');
          this.acting.set(false);
        },
      });
  }

  protected withdrawMyApplication(): void {
    const application = this.myApplication();
    if (!application || this.acting()) {
      return;
    }

    this.acting.set(true);
    this.error.set(null);

    this.applicationsApi.withdraw(application.id).subscribe({
      next: (updated) => {
        this.replaceApplication(updated);
        this.acting.set(false);
      },
      error: () => {
        this.error.set('Не удалось отозвать заявку.');
        this.acting.set(false);
      },
    });
  }

  protected approveByHr(application: CourseApplication): void {
    this.updateHrDecision(application.id, 'approve');
  }

  protected rejectByHr(application: CourseApplication): void {
    this.updateHrDecision(application.id, 'reject');
  }

  protected enroll(application: CourseApplication): void {
    if (this.acting()) {
      return;
    }

    this.acting.set(true);
    this.error.set(null);

    this.applicationsApi.enroll(application.id).subscribe({
      next: (updated) => {
        this.replaceApplication(updated);
        this.acting.set(false);
      },
      error: () => {
        this.error.set('Не удалось зачислить сотрудника.');
        this.acting.set(false);
      },
    });
  }

  private load(): void {
    const intakeId = this.route.snapshot.paramMap.get('intakeId');
    if (!intakeId) {
      this.error.set('Не найден intakeId.');
      this.loading.set(false);
      return;
    }

    this.loading.set(true);
    this.error.set(null);

    this.intakesApi.getById(intakeId).subscribe({
      next: (intake) => {
        this.applyIntakeState(intake);

        forkJoin({
          course: intake.course_id
            ? this.coursesApi.getById(intake.course_id).pipe(catchError(() => of(null)))
            : of(null),
          applications: this.canManage()
            ? this.intakesApi.listApplications(intake.id).pipe(catchError(() => of([])))
            : of([]),
          myApplications: this.applicationsApi.listMy().pipe(catchError(() => of([]))),
          approvers: this.hasDirectoryAccess()
            ? this.usersApi.listAdminUsers().pipe(
                catchError(() => {
                  this.directoryUnavailable.set(true);
                  return of([]);
                }),
              )
            : of([]),
        }).subscribe({
          next: ({ course, applications, myApplications, approvers }) => {
            this.course.set(course);
            this.applications.set(applications ?? []);
            this.myApplication.set(
              (myApplications ?? []).find((item) => item.intake_id === intake.id) ?? null,
            );
            this.approvers.set(approvers ?? []);
            this.loading.set(false);
          },
          error: () => {
            this.error.set('Не удалось загрузить детали набора.');
            this.loading.set(false);
          },
        });
      },
      error: () => {
        this.error.set('Не удалось загрузить набор.');
        this.loading.set(false);
      },
    });
  }

  private applyIntakeState(intake: CourseIntake): void {
    this.intake.set(intake);
    this.editForm.patchValue({
      title: intake.title,
      description: intake.description ?? '',
      approver_id: intake.approver_id ?? '',
      max_participants: intake.max_participants?.toString() ?? '',
      start_date: intake.start_date ?? '',
      end_date: intake.end_date ?? '',
      application_deadline: toDatetimeLocalValue(intake.application_deadline),
      status: intake.status,
    });
  }

  private updateHrDecision(applicationId: string, action: 'approve' | 'reject'): void {
    if (this.acting()) {
      return;
    }

    this.acting.set(true);
    this.error.set(null);

    const request$ =
      action === 'approve'
        ? this.applicationsApi.approveHr(applicationId, {
            comment: normalizeText(this.applicationComment(applicationId)),
          })
        : this.applicationsApi.rejectHr(applicationId, {
            comment: normalizeText(this.applicationComment(applicationId)),
          });

    request$.subscribe({
      next: (updated) => {
        this.replaceApplication(updated);
        this.acting.set(false);
      },
      error: () => {
        this.error.set(
          action === 'approve'
            ? 'Не удалось согласовать заявку HR.'
            : 'Не удалось отклонить заявку HR.',
        );
        this.acting.set(false);
      },
    });
  }

  private replaceApplication(updated: CourseApplication): void {
    this.applications.update((items) =>
      items.map((item) => (item.id === updated.id ? updated : item)),
    );

    const current = this.myApplication();
    if (current?.id === updated.id) {
      this.myApplication.set(updated);
    }
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
