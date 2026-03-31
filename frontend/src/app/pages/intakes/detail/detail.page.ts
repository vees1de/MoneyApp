import { CommonModule } from '@angular/common';
import { Component, OnInit, computed, inject, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { ActivatedRoute, RouterLink } from '@angular/router';
import { MatButtonModule } from '@angular/material/button';
import { MatButtonToggleModule } from '@angular/material/button-toggle';
import { MatCardModule } from '@angular/material/card';
import { MatCheckboxModule } from '@angular/material/checkbox';
import { MatDialog } from '@angular/material/dialog';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatIconModule } from '@angular/material/icon';
import { MatInputModule } from '@angular/material/input';
import { MatSelectModule } from '@angular/material/select';
import { catchError, forkJoin, of } from 'rxjs';

import { CertificatesApiService } from '@core/api/certificates-api.service';
import { CourseApplicationsApiService } from '@core/api/course-applications-api.service';
import { CourseIntakesApiService } from '@core/api/course-intakes-api.service';
import { CoursesApiService } from '@core/api/courses-api.service';
import { EnrollmentsApiService } from '@core/api/enrollments-api.service';
import { UsersApiService } from '@core/api/users-api.service';
import type { CourseApplication, CourseIntake } from '@core/api/contracts';
import { AuthStateService } from '@core/auth/auth-state.service';
import type { IdentityUserView, RoleCode } from '@core/auth/auth.types';
import { PERMISSIONS } from '@core/auth/permissions';
import { COURSE_INTAKE_MANAGER_APPROVAL_ENABLED } from '@core/config/feature-flags';
import { toDatetimeLocalValue } from '@core/domain/date-input.util';
import {
  buildIntakeSchedulePayload,
  normalizeText,
  resolveEndDateFromWeeks,
  toPositiveNumber,
} from '@core/domain/course-intake-form.util';
import {
  canApplyToIntake,
  canEnrollApplication,
  canHrReviewApplication,
  canWithdrawApplication,
  courseApplicationPaymentStatusLabel,
  courseApplicationStatusLabel,
  courseIntakeStatusLabel,
  isIntakeManageRole,
} from '@core/domain/course-intakes.workflow';
import { identityUserDisplayName } from '@core/domain/identity.util';
import type { Certificate } from '@entities/certificate';
import type { Course } from '@entities/course';
import type { Enrollment } from '@entities/enrollment';

import {
  IntakeSettingsDialogComponent,
  type IntakeSettingsDialogResult,
} from './intake-settings-dialog.component';

@Component({
  selector: 'app-page-intake-detail',
  standalone: true,
  imports: [
    CommonModule,
    ReactiveFormsModule,
    RouterLink,
    MatButtonModule,
    MatButtonToggleModule,
    MatCardModule,
    MatCheckboxModule,
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
  private readonly certificatesApi = inject(CertificatesApiService);
  private readonly coursesApi = inject(CoursesApiService);
  private readonly enrollmentsApi = inject(EnrollmentsApiService);
  private readonly usersApi = inject(UsersApiService);
  private readonly authState = inject(AuthStateService);
  private readonly route = inject(ActivatedRoute);
  private readonly fb = inject(FormBuilder);
  private readonly dialog = inject(MatDialog);

  protected readonly loading = signal(true);
  protected readonly acting = signal(false);
  protected readonly error = signal<string | null>(null);
  protected readonly directoryUnavailable = signal(false);
  protected readonly intake = signal<CourseIntake | null>(null);
  protected readonly course = signal<Course | null>(null);
  protected readonly applications = signal<CourseApplication[]>([]);
  protected readonly myApplication = signal<CourseApplication | null>(null);
  protected readonly myEnrollment = signal<Enrollment | null>(null);
  protected readonly users = signal<IdentityUserView[]>([]);
  protected readonly applicationComments = signal<Record<string, string>>({});
  protected readonly certificateComments = signal<Record<string, string>>({});

  protected readonly editForm = this.fb.group({
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
    status: ['open', [Validators.required]],
  });

  protected readonly applyForm = this.fb.group({
    motivation: [''],
    use_manager_approval: [COURSE_INTAKE_MANAGER_APPROVAL_ENABLED],
  });
  protected readonly managerApprovalEnabled = COURSE_INTAKE_MANAGER_APPROVAL_ENABLED;

  protected readonly role = computed<RoleCode>(
    () => this.authState.currentUser()?.roles[0] ?? 'employee',
  );
  protected readonly canManage = computed(
    () =>
      isIntakeManageRole(this.role()) || this.authState.hasPermission(PERMISSIONS.intakesManage),
  );
  protected readonly hasDirectoryAccess = computed(
    () =>
      this.authState.hasPermission(PERMISSIONS.usersRead) ||
      this.authState.hasAnyRole(['hr', 'admin']),
  );
  protected readonly canApply = computed(() => {
    const item = this.intake();
    return !!item && canApplyToIntake(item.status, !!this.myApplication());
  });
  protected readonly canVerifyCertificates = computed(() =>
    this.authState.hasPermission(PERMISSIONS.certificatesVerify),
  );
  protected readonly canWithdrawMyApplication = computed(() => {
    const application = this.myApplication();
    return !!application && canWithdrawApplication(application.status);
  });
  protected readonly isWeeksMode = computed(
    () => this.editForm.controls.schedule_mode.value === 'weeks',
  );
  protected readonly calculatedEndDate = computed(() =>
    resolveEndDateFromWeeks(
      this.editForm.controls.start_date.value,
      toPositiveNumber(this.editForm.controls.duration_weeks.value) ?? null,
    ),
  );

  protected readonly statusOptions = [
    { value: 'open', label: 'Открыт' },
    { value: 'closed', label: 'Набор закрыт' },
    { value: 'canceled', label: 'Отменён' },
    { value: 'completed', label: 'Завершён' },
  ];

  ngOnInit(): void {
    this.load();
  }

  protected intakeStatusLabel(status: string): string {
    return courseIntakeStatusLabel(status);
  }

  protected applicationStatusLabel(status: string): string {
    return courseApplicationStatusLabel(status);
  }

  protected paymentStatusLabel(status: string): string {
    return courseApplicationPaymentStatusLabel(status);
  }

  protected enrollmentStatusLabel(status?: string | null): string {
    const labels: Record<string, string> = {
      enrolled: 'Назначено',
      in_progress: 'В процессе',
      completed: 'Завершено',
      canceled: 'Отменено',
    };

    if (!status) {
      return '—';
    }

    return labels[status] ?? status;
  }

  protected certificateStatusLabel(status?: string | null): string {
    const labels: Record<string, string> = {
      uploaded: 'На проверке HR',
      verified: 'Подтверждён HR',
      rejected: 'Отклонён HR',
    };

    if (!status) {
      return 'Не загружен';
    }

    return labels[status] ?? status;
  }

  protected userLabel(userId: string | null | undefined): string {
    if (!userId) {
      return '—';
    }

    const currentUser = this.authState.currentUser();
    if (currentUser?.id === userId) {
      return identityUserDisplayName(currentUser);
    }

    const match = this.users().find((item) => item.id === userId);
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

  protected certificateComment(id: string): string {
    return this.certificateComments()[id] ?? '';
  }

  protected setCertificateComment(id: string, value: string): void {
    this.certificateComments.update((state) => ({
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

  protected canReviewCertificate(application: CourseApplication): boolean {
    return (
      !!application.certificate_id &&
      this.canVerifyCertificates() &&
      application.certificate_status !== 'verified'
    );
  }

  protected hasEnrolledApplications(): boolean {
    return this.applications().some((application) => application.status === 'enrolled');
  }

  protected hasStartableEnrollments(): boolean {
    return this.applications().some((application) => this.isEnrollmentStartable(application));
  }

  protected openManageDialog(): void {
    const item = this.intake();
    if (!item || this.acting()) {
      return;
    }

    const dialogRef = this.dialog.open<
      IntakeSettingsDialogComponent,
      { intake: CourseIntake },
      IntakeSettingsDialogResult | null
    >(IntakeSettingsDialogComponent, {
      width: '860px',
      maxWidth: '96vw',
      data: { intake: item },
    });

    dialogRef.afterClosed().subscribe((result) => {
      if (!result) {
        return;
      }

      if (result.action === 'close') {
        this.closeIntake();
        return;
      }

      this.editForm.patchValue(result.values);
      this.saveIntake();
    });
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

    this.intakesApi
      .update(item.id, {
        title,
        description: normalizeText(values.description),
        max_participants: toPositiveNumber(values.max_participants),
        price: values.price ?? '',
        price_currency: normalizeText(values.price_currency) ?? 'RUB',
        status: normalizeText(values.status),
        ...buildIntakeSchedulePayload({
          startDate: values.start_date,
          endDate: this.isWeeksMode() ? this.calculatedEndDate() : values.end_date,
          durationWeeks: this.isWeeksMode() ? toPositiveNumber(values.duration_weeks) : undefined,
          applicationDeadline: values.application_deadline,
        }),
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
        use_manager_approval: this.managerApprovalEnabled
          ? !!this.applyForm.controls.use_manager_approval.value
          : false,
      })
      .subscribe({
        next: (application) => {
          this.myApplication.set(application);
          this.applications.update((items) => [application, ...items]);
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
        this.error.set('Не удалось отметить сотрудника как взятого.');
        this.acting.set(false);
      },
    });
  }

  protected startEnrolledApplications(): void {
    if (this.acting()) {
      return;
    }

    const startableApplications = this.applications().filter((application) =>
      this.isEnrollmentStartable(application),
    );
    if (startableApplications.length === 0) {
      return;
    }

    this.acting.set(true);
    this.error.set(null);

    forkJoin(
      startableApplications.map((application) =>
        this.enrollmentsApi
          .start(application.enrollment_id as string)
          .pipe(catchError(() => of(null))),
      ),
    ).subscribe({
      next: (updatedEnrollments) => {
        let successCount = 0;

        this.applications.update((items) =>
          items.map((item) => {
            const index = startableApplications.findIndex(
              (application) => application.id === item.id,
            );
            if (index === -1) {
              return item;
            }

            const updatedEnrollment = updatedEnrollments[index];
            if (!updatedEnrollment) {
              return item;
            }

            successCount += 1;
            return {
              ...item,
              enrollment_status: updatedEnrollment.status,
            };
          }),
        );

        const current = this.myApplication();
        if (current) {
          const index = startableApplications.findIndex(
            (application) => application.id === current.id,
          );
          const updatedEnrollment = index >= 0 ? updatedEnrollments[index] : null;
          if (updatedEnrollment) {
            this.myEnrollment.set(updatedEnrollment);
            this.myApplication.set({
              ...current,
              enrollment_status: updatedEnrollment.status,
            });
          }
        }

        if (successCount === 0) {
          this.error.set('Не удалось стартовать ни один курс.');
        } else if (successCount < startableApplications.length) {
          this.error.set('Часть курсов не удалось стартовать.');
        }

        this.acting.set(false);
      },
    });
  }

  protected verifyCertificate(application: CourseApplication): void {
    this.updateCertificateDecision(application, 'verify');
  }

  protected rejectCertificate(application: CourseApplication): void {
    this.updateCertificateDecision(application, 'reject');
  }

  protected updateAllPayments(status: 'paid' | 'unpaid'): void {
    const item = this.intake();
    if (!item || this.acting()) {
      return;
    }

    this.acting.set(true);
    this.error.set(null);

    this.intakesApi.updatePaymentStatus(item.id, status).subscribe({
      next: (updatedApplications) => {
        const map = new Map(
          updatedApplications.map((application) => [application.id, application]),
        );
        this.applications.update((items) =>
          items.map((item) => (map.has(item.id) ? { ...item, ...(map.get(item.id) ?? {}) } : item)),
        );

        const current = this.myApplication();
        if (current && map.has(current.id)) {
          this.myApplication.set({ ...current, ...(map.get(current.id) ?? {}) });
        }

        this.acting.set(false);
      },
      error: () => {
        this.error.set('Не удалось обновить статус оплаты.');
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
          users: this.hasDirectoryAccess()
            ? this.usersApi.listAdminUsers().pipe(
                catchError(() => {
                  this.directoryUnavailable.set(true);
                  return of([]);
                }),
              )
            : of([]),
        }).subscribe({
          next: ({ course, applications, myApplications, users }) => {
            const currentMyApplication =
              (myApplications ?? []).find((application) => application.intake_id === intake.id) ??
              null;
            this.course.set(course);
            this.applications.set(applications ?? []);
            this.myApplication.set(currentMyApplication);
            this.users.set(users ?? []);

            if (!currentMyApplication?.enrollment_id) {
              this.myEnrollment.set(null);
              this.loading.set(false);
              return;
            }

            this.enrollmentsApi
              .getById(currentMyApplication.enrollment_id)
              .pipe(catchError(() => of(null)))
              .subscribe({
                next: (enrollment) => {
                  this.myEnrollment.set(enrollment);
                  if (enrollment && currentMyApplication) {
                    this.myApplication.set({
                      ...currentMyApplication,
                      enrollment_status: enrollment.status,
                    });
                  }
                  this.loading.set(false);
                },
                error: () => {
                  this.loading.set(false);
                },
              });
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
      max_participants: intake.max_participants?.toString() ?? '',
      price: intake.price ?? '',
      price_currency: intake.price_currency ?? 'RUB',
      start_date: intake.start_date ?? '',
      schedule_mode: intake.duration_weeks ? 'weeks' : 'end',
      duration_weeks: intake.duration_weeks?.toString() ?? '',
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
      items.map((item) => (item.id === updated.id ? { ...item, ...updated } : item)),
    );

    const current = this.myApplication();
    if (current?.id === updated.id) {
      this.myApplication.set({ ...current, ...updated });
      if (updated.enrollment_id) {
        this.enrollmentsApi
          .getById(updated.enrollment_id)
          .pipe(catchError(() => of(null)))
          .subscribe((enrollment) => {
            this.myEnrollment.set(enrollment);
            if (enrollment) {
              this.myApplication.update((application) =>
                application
                  ? {
                      ...application,
                      enrollment_status: enrollment.status,
                    }
                  : application,
              );
            }
          });
      }
    }
  }

  private updateCertificateDecision(
    application: CourseApplication,
    action: 'verify' | 'reject',
  ): void {
    if (!application.certificate_id || this.acting()) {
      return;
    }

    this.acting.set(true);
    this.error.set(null);

    const request$ =
      action === 'verify'
        ? this.certificatesApi.verify(application.certificate_id, {
            comment: normalizeText(this.certificateComment(application.id)),
          })
        : this.certificatesApi.reject(application.certificate_id, {
            comment: normalizeText(this.certificateComment(application.id)),
          });

    request$.subscribe({
      next: (certificate: Certificate) => {
        this.applications.update((items) =>
          items.map((item) =>
            item.id === application.id
              ? {
                  ...item,
                  certificate_id: certificate.id,
                  certificate_status: certificate.status,
                  certificate_uploaded_at: certificate.uploaded_at,
                }
              : item,
          ),
        );

        const current = this.myApplication();
        if (current?.id === application.id) {
          this.myApplication.set({
            ...current,
            certificate_id: certificate.id,
            certificate_status: certificate.status,
            certificate_uploaded_at: certificate.uploaded_at,
          });
        }

        this.acting.set(false);
      },
      error: () => {
        this.error.set(
          action === 'verify'
            ? 'Не удалось подтвердить сертификат.'
            : 'Не удалось отклонить сертификат.',
        );
        this.acting.set(false);
      },
    });
  }

  private isEnrollmentStartable(application: CourseApplication): boolean {
    return (
      !!application.enrollment_id &&
      (application.enrollment_status === 'enrolled' || application.enrollment_status == null)
    );
  }
}
