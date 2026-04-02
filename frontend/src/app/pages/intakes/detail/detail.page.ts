import { CommonModule } from '@angular/common';
import { HttpErrorResponse } from '@angular/common/http';
import { Component, OnInit, computed, inject, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { ActivatedRoute, Router, RouterLink } from '@angular/router';
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
import { EnrollmentsApiService } from '@core/api/enrollments-api.service';
import { ReportsApiService } from '@core/api/reports-api.service';
import { UsersApiService } from '@core/api/users-api.service';
import type { CourseApplication, CourseIntake } from '@core/api/contracts';
import { resolveApiUrl } from '@core/api/url.util';
import { API_BASE_URL } from '@core/config/api.config';
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
  canHrReviewApplication,
  canWithdrawApplication,
  courseApplicationPaymentStatusLabel,
  courseApplicationStatusLabel,
  courseIntakeStatusLabel,
  isIntakeManageRole,
} from '@core/domain/course-intakes.workflow';
import { identityUserDisplayName } from '@core/domain/identity.util';
import type { Enrollment } from '@entities/enrollment';

import { BudgetCheckBannerComponent } from '@app/widgets/budget-check-banner/budget-check-banner.component';

import {
  IntakeSettingsDialogComponent,
  type IntakeSettingsDialogResult,
} from './intake-settings-dialog.component';
import {
  CertificateReviewDialogComponent,
  type CertificateReviewDialogResult,
} from './certificate-review-dialog.component';

type HrApplicationFilter = 'all' | 'pending' | 'accepted' | 'rejected';

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
    BudgetCheckBannerComponent,
  ],
  templateUrl: './detail.page.html',
  styleUrl: './detail.page.scss',
})
export class IntakeDetailPageComponent implements OnInit {
  private readonly intakesApi = inject(CourseIntakesApiService);
  private readonly applicationsApi = inject(CourseApplicationsApiService);
  private readonly enrollmentsApi = inject(EnrollmentsApiService);
  private readonly usersApi = inject(UsersApiService);
  private readonly reportsApi = inject(ReportsApiService);
  private readonly certificatesApi = inject(CertificatesApiService);
  private readonly authState = inject(AuthStateService);
  private readonly route = inject(ActivatedRoute);
  private readonly router = inject(Router);
  private readonly fb = inject(FormBuilder);
  private readonly dialog = inject(MatDialog);

  protected readonly loading = signal(true);
  protected readonly acting = signal(false);
  protected readonly exporting = signal(false);
  protected readonly hrFilter = signal<HrApplicationFilter>('all');
  protected readonly error = signal<string | null>(null);
  protected readonly directoryUnavailable = signal(false);
  protected readonly intake = signal<CourseIntake | null>(null);
  protected readonly applications = signal<CourseApplication[]>([]);
  protected readonly myApplication = signal<CourseApplication | null>(null);
  protected readonly myEnrollment = signal<Enrollment | null>(null);
  protected readonly users = signal<IdentityUserView[]>([]);
  protected readonly applicationComments = signal<Record<string, string>>({});
  protected readonly budgetBlocked = signal(false);

  protected readonly intakePriceNumeric = computed(() => {
    const item = this.intake();
    if (!item) return 0;
    const normalized = item.price?.trim().replace(/\s+/g, '').replace(',', '.');
    if (!normalized) return 0;
    const parsed = Number(normalized);
    return Number.isFinite(parsed) ? parsed : 0;
  });

  protected onBudgetBlocked(blocked: boolean): void {
    this.budgetBlocked.set(blocked);
  }

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
  protected readonly canWithdrawMyApplication = computed(() => {
    const application = this.myApplication();
    return !!application && canWithdrawApplication(application.status);
  });
  protected readonly canOpenLearningCta = computed(() => {
    const application = this.myApplication();
    if (!application?.enrollment_id || application.status !== 'enrolled') {
      return false;
    }

    const enrollmentStatus =
      this.myEnrollment()?.status ?? application.enrollment_status ?? null;
    return enrollmentStatus === 'in_progress';
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
  protected readonly isCourseStarted = computed(() =>
    this.applications().some(
      (a) =>
        a.enrollment_status === 'in_progress' ||
        a.enrollment_status === 'completed',
    ),
  );

  protected readonly poolApplicants = computed(() =>
    this.applications().filter(
      (a) =>
        a.status === 'pending' ||
        a.status === 'pending_manager' ||
        a.status === 'approved_by_manager' ||
        a.status === 'approved' ||
        a.status === 'rejected_by_hr' ||
        a.status === 'rejected_by_manager' ||
        a.status === 'withdrawn',
    ),
  );

  protected readonly poolLearning = computed(() =>
    this.applications().filter((a) => a.status === 'enrolled'),
  );

  protected readonly hrFilterOptions: ReadonlyArray<{
    value: HrApplicationFilter;
    label: string;
  }> = [
    { value: 'all', label: 'Все' },
    { value: 'pending', label: 'Без решения' },
    { value: 'accepted', label: 'Приняты' },
    { value: 'rejected', label: 'Не приняты' },
  ];
  protected readonly hrFilterCounts = computed(() => {
    const counts = {
      pending: 0,
      accepted: 0,
      rejected: 0,
    };

    for (const application of this.applications()) {
      counts[this.applicationDecisionGroup(application)] += 1;
    }

    return counts;
  });
  protected readonly intakeMaxPriceLabel = computed(() => {
    const item = this.intake();
    if (!item) {
      return '—';
    }

    const maxParticipants = item.max_participants;
    const pricePerParticipant = this.parsePrice(item.price);
    if (
      maxParticipants === null ||
      maxParticipants === undefined ||
      maxParticipants <= 0 ||
      pricePerParticipant === null ||
      pricePerParticipant < 0
    ) {
      return '—';
    }

    const total = maxParticipants * pricePerParticipant;
    const currency = item.price_currency?.trim() || 'RUB';
    return this.formatCurrencyAmount(total, currency);
  });
  protected readonly intakeCurrentPriceLabel = computed(() => {
    const item = this.intake();
    if (!item) {
      return '—';
    }

    const pricePerParticipant = this.parsePrice(item.price);
    if (pricePerParticipant === null || pricePerParticipant < 0) {
      return '—';
    }

    const acceptedParticipants = this.hrFilterCounts().accepted;
    const total = acceptedParticipants * pricePerParticipant;
    const currency = item.price_currency?.trim() || 'RUB';
    return this.formatCurrencyAmount(total, currency);
  });
  protected readonly filteredApplications = computed(() => {
    const filter = this.hrFilter();

    return [...this.applications()]
      .filter(
        (application) =>
          filter === 'all' || this.applicationDecisionGroup(application) === filter,
      )
      .sort((left, right) => {
        const groupDiff =
          this.applicationGroupOrder(left) - this.applicationGroupOrder(right);
        if (groupDiff !== 0) {
          return groupDiff;
        }

        const nameDiff = this.userLabel(left.applicant_id).localeCompare(
          this.userLabel(right.applicant_id),
          'ru',
        );
        if (nameDiff !== 0) {
          return nameDiff;
        }

        return new Date(right.updated_at).getTime() - new Date(left.updated_at).getTime();
      });
  });

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

  protected canHrApprove(application: CourseApplication): boolean {
    return canHrReviewApplication(application.status);
  }

  protected hasEnrolledApplications(): boolean {
    return this.applications().some((application) => application.status === 'enrolled');
  }

  protected setHrFilter(filter: HrApplicationFilter): void {
    this.hrFilter.set(filter);
  }

  protected exportParticipantsExcel(): void {
    const item = this.intake();
    if (!item || this.exporting()) {
      return;
    }

    this.exporting.set(true);
    this.error.set(null);

    this.reportsApi.exportIntakeApplicationsExcel(item.id).subscribe({
      next: (blob) => {
        this.downloadIntakeExport(blob, item.title);
        this.exporting.set(false);
      },
      error: () => {
        this.error.set('Не удалось выгрузить Excel по выбранному набору.');
        this.exporting.set(false);
      },
    });
  }

  protected hrFilterCount(filter: HrApplicationFilter): number {
    if (filter === 'all') {
      return this.applications().length;
    }

    return this.hrFilterCounts()[filter];
  }

  protected applicationDecisionLabel(application: CourseApplication): string | null {
    const group = this.applicationDecisionGroup(application);
    if (group === 'accepted') {
      return 'Принят';
    }
    if (group === 'rejected') {
      return 'Не принят';
    }

    return null;
  }

  protected applicationDecisionClass(application: CourseApplication): string {
    const group = this.applicationDecisionGroup(application);
    if (group === 'accepted') {
      return 'employee-card__decision--accepted';
    }
    if (group === 'rejected') {
      return 'employee-card__decision--rejected';
    }

    return 'employee-card__decision--pending';
  }

  protected isRejectedByHr(application: CourseApplication): boolean {
    return application.status === 'rejected_by_hr';
  }

  protected applicationWorkflowHint(application: CourseApplication): string {
    const status = application.status;

    if (status === 'pending_manager') {
      return 'Ждёт решения руководителя';
    }
    if (status === 'approved_by_manager' || status === 'pending') {
      return 'Ждёт решения HR';
    }
    if (status === 'approved') {
      return 'Принят HR';
    }
    if (status === 'enrolled') {
      return 'Принят и добавлен в набор';
    }
    if (status === 'rejected_by_manager') {
      return 'Отклонён руководителем';
    }
    if (status === 'rejected_by_hr') {
      return 'Не принят HR';
    }
    if (status === 'withdrawn') {
      return 'Сотрудник отозвал заявку';
    }

    return this.applicationStatusLabel(status);
  }

  protected userPosition(userId: string | null | undefined): string | null {
    return normalizeText(this.lookupUser(userId)?.employee_profile?.position_title) ?? null;
  }

  protected userEmail(userId: string | null | undefined): string | null {
    return this.lookupUser(userId)?.email ?? null;
  }

  protected userAvatarUrl(userId: string | null | undefined): string | null {
    return resolveApiUrl(this.lookupUser(userId)?.avatar_url);
  }

  protected userAvatarInitials(userId: string | null | undefined): string {
    const user = this.lookupUser(userId);
    if (!user) {
      return 'U';
    }

    const profile = user.employee_profile;
    if (profile) {
      const initials = `${profile.last_name?.[0] ?? ''}${profile.first_name?.[0] ?? ''}`
        .trim()
        .toUpperCase();
      if (initials) {
        return initials;
      }
    }

    const emailInitial = user.email?.trim()?.[0]?.toUpperCase();
    return emailInitial || 'U';
  }

  protected hasStartableEnrollments(): boolean {
    return this.applications().some((application) => this.isEnrollmentStartable(application));
  }

  protected deleteIntake(): void {
    const item = this.intake();
    if (!item || this.acting()) {
      return;
    }

    const confirmed = window.confirm(
      `Удалить набор "${item.title}"? Это действие нельзя отменить.`,
    );
    if (!confirmed) {
      return;
    }

    this.acting.set(true);
    this.error.set(null);

    this.intakesApi.delete(item.id).subscribe({
      next: () => {
        this.acting.set(false);
        void this.router.navigateByUrl('/intakes');
      },
      error: (error) => {
        this.error.set(this.extractDeleteErrorMessage(error));
        this.acting.set(false);
      },
    });
  }

  protected startButtonHint(): string {
    if (this.applications().some((application) => application.status === 'enrolled')) {
      return 'Кнопка активна, пока в наборе есть принятые сотрудники, у которых курс ещё не стартован.';
    }

    return 'Кнопка станет активной после действия «Принять».';
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
    if (this.acting()) {
      return;
    }

    this.acting.set(true);
    this.error.set(null);

    this.applicationsApi
      .approveHr(application.id, {
        comment: normalizeText(this.applicationComment(application.id)),
      })
      .subscribe({
        next: (approved) => {
          this.replaceApplication(approved);

          this.applicationsApi.enroll(approved.id).subscribe({
            next: (enrolled) => {
              this.replaceApplication(enrolled);
              this.acting.set(false);
            },
            error: () => {
              this.error.set(
                'Заявка согласована HR, но сотрудника не удалось добавить в набор.',
              );
              this.acting.set(false);
            },
          });
        },
        error: () => {
          this.error.set('Не удалось принять сотрудника в набор.');
          this.acting.set(false);
        },
      });
  }

  protected rejectByHr(application: CourseApplication): void {
    this.updateHrDecision(application.id, 'reject');
  }

  protected startEnrolledApplications(): void {
    const item = this.intake();
    if (!item || this.acting()) {
      return;
    }

    if (!this.hasStartableEnrollments()) {
      return;
    }

    this.acting.set(true);
    this.error.set(null);

    this.intakesApi.startCourse(item.id).subscribe({
      next: (updatedApplications) => {
        const map = new Map(
          updatedApplications.map((application) => [application.id, application]),
        );
        this.applications.update((items) =>
          items.map((current) =>
            map.has(current.id) ? { ...current, ...(map.get(current.id) ?? {}) } : current,
          ),
        );

        const current = this.myApplication();
        if (current && map.has(current.id)) {
          const updated = map.get(current.id);
          this.myApplication.set({ ...current, ...(updated ?? {}) });
          if (updated?.enrollment_id) {
            this.enrollmentsApi
              .getById(updated.enrollment_id)
              .pipe(catchError(() => of(null)))
              .subscribe((enrollment) => this.myEnrollment.set(enrollment));
          }
        }

        this.acting.set(false);
      },
      error: () => {
        this.error.set('Не удалось стартовать курс для набора.');
        this.acting.set(false);
      },
    });
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

  protected learningStatusLabel(application: CourseApplication): string {
    if (application.certificate_status === 'verified') {
      return 'Завершён';
    }
    if (application.certificate_status === 'uploaded') {
      return 'Сертификат предоставлен';
    }
    if (application.certificate_status === 'rejected') {
      return 'Сертификат отклонён';
    }
    if (application.enrollment_status === 'in_progress') {
      return 'В процессе';
    }
    if (application.enrollment_status === 'completed') {
      return 'Завершено';
    }
    return 'Ожидает старта';
  }

  protected learningStatusClass(application: CourseApplication): string {
    if (application.certificate_status === 'verified') {
      return 'pool-status--success';
    }
    if (application.certificate_status === 'uploaded') {
      return 'pool-status--info';
    }
    if (application.certificate_status === 'rejected') {
      return 'pool-status--error';
    }
    if (application.enrollment_status === 'in_progress') {
      return 'pool-status--warning';
    }
    return 'pool-status--neutral';
  }

  protected openCertificateReview(application: CourseApplication): void {
    if (!application.certificate_id) {
      return;
    }

    const fileUrl = this.buildCertificateUrl(application);

    const dialogRef = this.dialog.open<
      CertificateReviewDialogComponent,
      {
        application: CourseApplication;
        userName: string;
        fileUrl: string | null;
        fileName: string | null;
      },
      CertificateReviewDialogResult | null
    >(CertificateReviewDialogComponent, {
      width: '100vw',
      maxWidth: '100vw',
      height: '100vh',
      panelClass: 'certificate-review-fullscreen',
      data: {
        application,
        userName: this.userLabel(application.applicant_id),
        fileUrl,
        fileName: application.certificate_file_original_name ?? null,
      },
    });

    dialogRef.afterClosed().subscribe((result) => {
      if (!result || !application.certificate_id) {
        return;
      }

      this.acting.set(true);
      this.error.set(null);

      const request$ =
        result.action === 'approve'
          ? this.certificatesApi.verify(application.certificate_id, { comment: result.comment })
          : this.certificatesApi.reject(application.certificate_id, { comment: result.comment });

      request$.subscribe({
        next: () => {
          const newCertStatus = result.action === 'approve' ? 'verified' : 'rejected';
          this.applications.update((items) =>
            items.map((item) =>
              item.id === application.id
                ? { ...item, certificate_status: newCertStatus }
                : item,
            ),
          );
          this.acting.set(false);
        },
        error: () => {
          this.error.set(
            result.action === 'approve'
              ? 'Не удалось подтвердить сертификат.'
              : 'Не удалось отклонить сертификат.',
          );
          this.acting.set(false);
        },
      });
    });
  }

  private buildCertificateUrl(application: CourseApplication): string | null {
    const storageKey = application.certificate_file_storage_key?.trim();
    if (!storageKey) {
      return null;
    }
    return `${API_BASE_URL}/uploads/${encodeURI(storageKey)}`;
  }

  private downloadIntakeExport(blob: Blob, intakeTitle: string): void {
    const now = new Date();
    const datePart = now.toISOString().slice(0, 10).replace(/-/g, '');
    const safeTitle = intakeTitle
      .trim()
      .toLowerCase()
      .replace(/[^a-z0-9а-яё]+/gi, '-')
      .replace(/^-+|-+$/g, '')
      .slice(0, 60);
    const filename = `${safeTitle || 'intake'}-participants-${datePart}.xlsx`;

    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = filename;
    a.click();
    URL.revokeObjectURL(url);
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
          applications: this.canManage()
            ? this.intakesApi.listApplications(intake.id).pipe(catchError(() => of([])))
            : of([]),
          myApplications: this.canManage()
            ? of([])
            : this.applicationsApi.listMy().pipe(catchError(() => of([]))),
          users: this.hasDirectoryAccess()
            ? this.usersApi.listAdminUsers().pipe(
                catchError(() => {
                  this.directoryUnavailable.set(true);
                  return of([]);
                }),
              )
            : of([]),
        }).subscribe({
          next: ({ applications, myApplications, users }) => {
            const currentMyApplication =
              (myApplications ?? []).find((application) => application.intake_id === intake.id) ??
              null;
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

  private applicationDecisionGroup(
    application: CourseApplication,
  ): Exclude<HrApplicationFilter, 'all'> {
    if (application.status === 'approved' || application.status === 'enrolled') {
      return 'accepted';
    }

    if (
      application.status === 'rejected_by_hr' ||
      application.status === 'rejected_by_manager' ||
      application.status === 'withdrawn'
    ) {
      return 'rejected';
    }

    return 'pending';
  }

  private applicationGroupOrder(application: CourseApplication): number {
    const group = this.applicationDecisionGroup(application);
    if (group === 'pending') {
      return 0;
    }
    if (group === 'accepted') {
      return 1;
    }

    return 2;
  }

  private lookupUser(userId: string | null | undefined): IdentityUserView | null {
    if (!userId) {
      return null;
    }

    const currentUser = this.authState.currentUser();
    if (currentUser?.id === userId) {
      return currentUser;
    }

    return this.users().find((item) => item.id === userId) ?? null;
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

  private extractDeleteErrorMessage(error: unknown): string {
    if (error instanceof HttpErrorResponse) {
      const apiError = error.error?.error;
      const apiCode = typeof apiError?.code === 'string' ? apiError.code : null;
      if (apiCode === 'intake_has_applications') {
        return 'Нельзя удалить набор, пока по нему есть заявки.';
      }
      if (apiCode === 'intake_linked_to_suggestion') {
        return 'Набор связан с предложением и не может быть удалён.';
      }

      const apiMessage = apiError?.message;
      if (typeof apiMessage === 'string' && apiMessage.trim()) {
        return apiMessage.trim();
      }
    }

    return 'Не удалось удалить набор.';
  }

  private isEnrollmentStartable(application: CourseApplication): boolean {
    if (application.status !== 'enrolled') {
      return false;
    }

    if (!application.enrollment_id) {
      return true;
    }

    return application.enrollment_status === 'enrolled' || application.enrollment_status == null;
  }

  private parsePrice(value: string | null | undefined): number | null {
    const normalized = value?.trim().replace(/\s+/g, '').replace(',', '.');
    if (!normalized) {
      return null;
    }

    const parsed = Number(normalized);
    return Number.isFinite(parsed) ? parsed : null;
  }

  private formatCurrencyAmount(amount: number, currency: string): string {
    try {
      return new Intl.NumberFormat('ru-RU', {
        style: 'currency',
        currency,
        maximumFractionDigits: 2,
      }).format(amount);
    } catch {
      const formatted = new Intl.NumberFormat('ru-RU', {
        maximumFractionDigits: 2,
      }).format(amount);
      return `${formatted} ${currency}`;
    }
  }
}
