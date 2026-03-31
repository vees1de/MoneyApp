import { CommonModule } from '@angular/common';
import { Component, OnInit, computed, inject, signal } from '@angular/core';
import {
  FormsModule,
  ReactiveFormsModule,
  UntypedFormArray,
  UntypedFormBuilder,
  UntypedFormGroup,
  Validators,
} from '@angular/forms';
import { MatButtonModule } from '@angular/material/button';
import { MatCardModule } from '@angular/material/card';
import { MatCheckboxModule } from '@angular/material/checkbox';
import { MatChipsModule } from '@angular/material/chips';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatIconModule } from '@angular/material/icon';
import { MatInputModule } from '@angular/material/input';
import { MatSelectModule } from '@angular/material/select';
import { MatTabsModule } from '@angular/material/tabs';
import { RouterLink } from '@angular/router';

import type {
  AdminRole,
  ApprovalWorkflow,
  BudgetLimitRecord,
  ExternalRequest,
  HrDashboardStats,
} from '@core/api/contracts';
import type { IdentityUserView } from '@core/auth/auth.types';
import type { Course } from '@entities/course';
import { AdminFacade } from '@features/admin';

@Component({
  selector: 'app-page-admin-testadmin',
  standalone: true,
  imports: [
    CommonModule,
    FormsModule,
    ReactiveFormsModule,
    RouterLink,
    MatButtonModule,
    MatCardModule,
    MatCheckboxModule,
    MatChipsModule,
    MatFormFieldModule,
    MatIconModule,
    MatInputModule,
    MatSelectModule,
    MatTabsModule,
  ],
  templateUrl: './testadmin.page.html',
  styleUrl: './testadmin.page.scss',
})
export class AdminTestAdminPageComponent implements OnInit {
  private readonly fb = inject(UntypedFormBuilder);

  protected readonly facade = inject(AdminFacade);
  protected readonly editingCourseId = signal<string | null>(null);
  protected readonly userSearch = signal('');
  protected readonly courseSearch = signal('');
  protected readonly requestSearch = signal('');
  protected readonly requestStatus = signal('all');
  protected readonly requestComments = signal<Record<string, string>>({});

  protected readonly filteredUsers = computed(() => {
    const query = this.userSearch().trim().toLowerCase();
    if (!query) {
      return this.facade.users();
    }

    return this.facade.users().filter((user) => {
      const haystack = [
        this.userName(user),
        user.email,
        user.employee_profile?.position_title ?? '',
        user.roles.join(' '),
      ]
        .join(' ')
        .toLowerCase();

      return haystack.includes(query);
    });
  });

  protected readonly filteredCourses = computed(() => {
    const query = this.courseSearch().trim().toLowerCase();
    if (!query) {
      return this.facade.courses();
    }

    return this.facade
      .courses()
      .filter((course) =>
        [
          course.title,
          course.short_description ?? '',
          course.description ?? '',
          course.language ?? '',
        ]
          .join(' ')
          .toLowerCase()
          .includes(query),
      );
  });

  protected readonly filteredRequests = computed(() => {
    const query = this.requestSearch().trim().toLowerCase();
    const status = this.requestStatus();

    return this.facade.externalRequests().filter((request) => {
      const statusMatches = status === 'all' ? true : request.status === status;
      if (!statusMatches) {
        return false;
      }

      if (!query) {
        return true;
      }

      const haystack = [
        request.title,
        request.employee_full_name ?? '',
        request.employee_email ?? '',
        request.provider_name ?? '',
        request.status,
        request.current_approval_role_code ?? '',
      ]
        .join(' ')
        .toLowerCase();

      return haystack.includes(query);
    });
  });

  protected readonly courseForm = this.fb.group({
    title: ['', Validators.required],
    type: ['internal', Validators.required],
    source_type: ['catalog', Validators.required],
    short_description: [''],
    description: [''],
    language: [''],
    duration_hours: [''],
    external_url: [''],
    price: [''],
    price_currency: ['RUB'],
    next_start_date: [''],
    is_mandatory_default: [false],
  });

  protected readonly assignmentForm = this.fb.group({
    course_id: ['', Validators.required],
    target_id: ['', Validators.required],
    priority: ['recommended', Validators.required],
    reason: [''],
    start_at: [''],
    deadline_at: [''],
  });

  protected readonly budgetLimitForm = this.fb.group({
    scope_type: ['company', Validators.required],
    scope_id: [''],
    period_year: [new Date().getFullYear(), Validators.required],
    period_month: [''],
    limit_amount: ['', Validators.required],
    currency: ['RUB', Validators.required],
    is_active: [true],
  });

  protected readonly workflowForm = this.fb.group({
    name: ['', Validators.required],
    is_active: [true],
    steps: this.fb.array([
      this.createWorkflowStep('manager', 'line_manager'),
      this.createWorkflowStep('hr', 'specific_role'),
    ]),
  });

  ngOnInit(): void {
    this.facade.load();
  }

  protected get workflowSteps(): UntypedFormArray {
    return this.workflowForm.get('steps') as UntypedFormArray;
  }

  protected workflowStepGroups(): UntypedFormGroup[] {
    return this.workflowSteps.controls as UntypedFormGroup[];
  }

  protected metrics(): Array<{ label: string; value: number; tone: string }> {
    const analytics = this.facade.analytics();
    return [
      {
        label: 'Сотрудники',
        value: analytics?.users ?? this.facade.totalEmployees(),
        tone: 'info',
      },
      {
        label: 'Курсы в каталоге',
        value: analytics?.courses ?? this.facade.courses().length,
        tone: 'success',
      },
      {
        label: 'Активные назначения',
        value: this.facade.assignments().length,
        tone: 'warning',
      },
      {
        label: 'Внешние заявки',
        value: analytics?.external_requests ?? this.facade.externalRequests().length,
        tone: 'info',
      },
      {
        label: 'Ожидают согласования',
        value: analytics?.pending_approvals ?? this.facade.pendingApprovals().length,
        tone: 'warning',
      },
      {
        label: 'Лимиты бюджета',
        value: this.facade.activeBudgetLimitCount(),
        tone: 'success',
      },
    ];
  }

  protected analyticsRows(): Array<{ label: string; value: number }> {
    const analytics = this.facade.analytics();
    if (!analytics) {
      return [];
    }

    return [
      { label: 'Пользователи', value: analytics.users },
      { label: 'Курсы', value: analytics.courses },
      { label: 'Enrollments', value: analytics.enrollments },
      { label: 'Внешние заявки', value: analytics.external_requests },
      { label: 'Pending approvals', value: analytics.pending_approvals },
    ];
  }

  protected roleCodes(): string[] {
    return this.facade.knownRoleCodes();
  }

  protected userName(user: IdentityUserView): string {
    const profile = user.employee_profile;
    if (!profile) {
      return user.email;
    }

    return [profile.last_name, profile.first_name, profile.middle_name]
      .filter((item) => !!item)
      .join(' ')
      .trim();
  }

  protected roleName(roleCode: string): string {
    return this.facade.roles().find((role) => role.code === roleCode)?.name ?? roleCode;
  }

  protected statusClass(status: string | null | undefined): string {
    const value = (status ?? '').toLowerCase();
    if (value.includes('approved') || value.includes('published') || value === 'active') {
      return 'status-badge--success';
    }
    if (value.includes('draft') || value.includes('revision') || value.includes('pending')) {
      return 'status-badge--warning';
    }
    if (value.includes('reject') || value.includes('archived') || value.includes('blocked')) {
      return 'status-badge--danger';
    }
    return 'status-badge--info';
  }

  protected courseState(course: Course): string {
    if (course.archived_at || course.status === 'archived') {
      return 'Скрыт';
    }
    if (course.status === 'published') {
      return 'Опубликован';
    }
    return 'Черновик';
  }

  protected requestStep(request: ExternalRequest): string {
    if (request.current_approval_role_code) {
      return request.current_approval_role_code;
    }
    if (request.current_approval_status) {
      return request.current_approval_status;
    }
    return 'n/a';
  }

  protected requestComment(id: string): string {
    return this.requestComments()[id] ?? '';
  }

  protected setRequestComment(id: string, value: string): void {
    this.requestComments.update((state) => ({
      ...state,
      [id]: value,
    }));
  }

  protected canActOnRequest(request: ExternalRequest): boolean {
    return this.facade.canApproveRequests() && request.current_approval_status === 'pending';
  }

  protected assignmentCourseTitle(courseId: string): string {
    return this.facade.courses().find((course) => course.id === courseId)?.title ?? courseId;
  }

  protected budgetScopeLabel(item: BudgetLimitRecord): string {
    if (item.scope_type === 'company') {
      return 'Company-wide';
    }
    if (item.scope_type === 'employee') {
      return this.userNameById(item.scope_id);
    }
    return item.scope_id ?? 'Department';
  }

  protected workflowSummary(workflow: ApprovalWorkflow): string {
    return (workflow.steps ?? [])
      .map(
        (step) => `${step.step_order}. ${this.roleName(step.role_code)} / ${step.approver_source}`,
      )
      .join(' · ');
  }

  protected addWorkflowStep(): void {
    this.workflowSteps.push(this.createWorkflowStep('hr', 'specific_role'));
  }

  protected removeWorkflowStep(index: number): void {
    if (this.workflowSteps.length <= 1) {
      return;
    }

    this.workflowSteps.removeAt(index);
  }

  protected async refresh(): Promise<void> {
    await this.facade.refresh();
  }

  protected async toggleRole(user: IdentityUserView, roleCode: string): Promise<void> {
    await this.facade.toggleUserRole(user, roleCode);
  }

  protected startEditingCourse(course: Course): void {
    this.editingCourseId.set(course.id);
    this.courseForm.patchValue({
      title: course.title,
      type: course.type,
      source_type: course.source_type,
      short_description: course.short_description ?? '',
      description: course.description ?? '',
      language: course.language ?? '',
      duration_hours: course.duration_hours ?? '',
      external_url: course.external_url ?? '',
      price: course.price ?? '',
      price_currency: course.price_currency ?? 'RUB',
      next_start_date: this.dateInputValue(course.next_start_date),
      is_mandatory_default: course.is_mandatory_default,
    });
  }

  protected resetCourseForm(): void {
    this.editingCourseId.set(null);
    this.courseForm.reset({
      title: '',
      type: 'internal',
      source_type: 'catalog',
      short_description: '',
      description: '',
      language: '',
      duration_hours: '',
      external_url: '',
      price: '',
      price_currency: 'RUB',
      next_start_date: '',
      is_mandatory_default: false,
    });
  }

  protected async submitCourse(): Promise<void> {
    if (this.courseForm.invalid) {
      this.courseForm.markAllAsTouched();
      return;
    }

    const value = this.courseForm.getRawValue();
    await this.facade.saveCourse(this.editingCourseId(), {
      title: this.requiredTrim(value.title),
      type: value.type,
      source_type: value.source_type,
      short_description: this.optionalTrim(value.short_description),
      description: this.optionalTrim(value.description),
      language: this.optionalTrim(value.language),
      duration_hours: this.optionalTrim(value.duration_hours),
      external_url: this.optionalTrim(value.external_url),
      price: this.optionalTrim(value.price),
      price_currency: this.optionalTrim(value.price_currency),
      next_start_date: this.dateToIso(value.next_start_date),
      is_mandatory_default: !!value.is_mandatory_default,
    });

    if (!this.facade.error()) {
      this.resetCourseForm();
    }
  }

  protected async publishCourse(course: Course): Promise<void> {
    await this.facade.publishCourse(course.id);
  }

  protected async archiveCourse(course: Course): Promise<void> {
    await this.facade.archiveCourse(course.id);
  }

  protected async submitAssignment(): Promise<void> {
    if (this.assignmentForm.invalid) {
      this.assignmentForm.markAllAsTouched();
      return;
    }

    const value = this.assignmentForm.getRawValue();
    await this.facade.createAssignment({
      course_id: value.course_id,
      assignment_type: 'individual',
      target_type: 'user',
      target_id: value.target_id,
      priority: value.priority,
      reason: this.optionalTrim(value.reason),
      start_at: this.dateToIso(value.start_at),
      deadline_at: this.dateToIso(value.deadline_at),
    });

    if (!this.facade.error()) {
      this.assignmentForm.reset({
        course_id: '',
        target_id: '',
        priority: 'recommended',
        reason: '',
        start_at: '',
        deadline_at: '',
      });
    }
  }

  protected async approveRequest(request: ExternalRequest): Promise<void> {
    await this.facade.approveRequest(request.id, this.requestComment(request.id));
  }

  protected async rejectRequest(request: ExternalRequest): Promise<void> {
    await this.facade.rejectRequest(request.id, this.requestComment(request.id));
  }

  protected async requestRevision(request: ExternalRequest): Promise<void> {
    await this.facade.requestRevision(request.id, this.requestComment(request.id));
  }

  protected async submitBudgetLimit(): Promise<void> {
    if (this.budgetLimitForm.invalid) {
      this.budgetLimitForm.markAllAsTouched();
      return;
    }

    const value = this.budgetLimitForm.getRawValue();
    await this.facade.createBudgetLimit({
      scope_type: value.scope_type,
      scope_id: value.scope_type === 'company' ? null : this.optionalTrim(value.scope_id),
      period_year: Number(value.period_year),
      period_month: value.period_month ? Number(value.period_month) : null,
      limit_amount: this.requiredTrim(value.limit_amount),
      currency: this.requiredTrim(value.currency),
      is_active: !!value.is_active,
    });

    if (!this.facade.error()) {
      this.budgetLimitForm.reset({
        scope_type: 'company',
        scope_id: '',
        period_year: new Date().getFullYear(),
        period_month: '',
        limit_amount: '',
        currency: 'RUB',
        is_active: true,
      });
    }
  }

  protected async submitWorkflow(): Promise<void> {
    if (this.workflowForm.invalid) {
      this.workflowForm.markAllAsTouched();
      return;
    }

    const value = this.workflowForm.getRawValue();
    const steps = (value.steps ?? []).map(
      (
        step: {
          role_code: string;
          approver_source: string;
          approver_user_id?: string;
          sla_hours?: string;
          is_required?: boolean;
        },
        index: number,
      ) => ({
        step_order: index + 1,
        role_code: step.role_code,
        approver_source: step.approver_source,
        approver_user_id: this.optionalTrim(step.approver_user_id),
        sla_hours: step.sla_hours ? Number(step.sla_hours) : null,
        is_required: step.is_required !== false,
      }),
    );

    await this.facade.createWorkflow({
      entity_type: 'external_course_request',
      name: this.requiredTrim(value.name),
      is_active: !!value.is_active,
      steps,
    });

    if (!this.facade.error()) {
      this.workflowForm.reset({
        name: '',
        is_active: true,
      });
      this.workflowSteps.clear();
      this.workflowSteps.push(this.createWorkflowStep('manager', 'line_manager'));
      this.workflowSteps.push(this.createWorkflowStep('hr', 'specific_role'));
    }
  }

  private createWorkflowStep(roleCode: string, approverSource: string) {
    return this.fb.group({
      role_code: [roleCode, Validators.required],
      approver_source: [approverSource, Validators.required],
      approver_user_id: [''],
      sla_hours: [''],
      is_required: [true],
    });
  }

  private requiredTrim(value: string): string {
    return value.trim();
  }

  private optionalTrim(value: string | null | undefined): string | null {
    const normalized = value?.trim() ?? '';
    return normalized ? normalized : null;
  }

  private dateToIso(value: string | null | undefined): string | null {
    if (!value) {
      return null;
    }

    return new Date(`${value}T00:00:00Z`).toISOString();
  }

  private dateInputValue(value: string | null | undefined): string {
    if (!value) {
      return '';
    }

    return value.slice(0, 10);
  }

  private userNameById(userId: string | null | undefined): string {
    if (!userId) {
      return 'Не указан';
    }

    const user = this.facade.users().find((item) => item.id === userId);
    return user ? this.userName(user) : userId;
  }
}
