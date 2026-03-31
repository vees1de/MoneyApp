import { HttpErrorResponse } from '@angular/common/http';
import { Injectable, computed, inject, signal } from '@angular/core';
import { firstValueFrom } from 'rxjs';

import { ApprovalWorkflowsApiService } from '@core/api/approval-workflows-api.service';
import { AnalyticsApiService } from '@core/api/analytics-api.service';
import { AssignmentsApiService } from '@core/api/assignments-api.service';
import { BudgetLimitsApiService } from '@core/api/budget-limits-api.service';
import type {
  AdminRole,
  ApprovalWorkflow,
  BudgetLimitRecord,
  CourseAssignment,
  ExternalRequest,
  HrDashboardStats,
  PendingApprovalItem,
} from '@core/api/contracts';
import { CoursesApiService } from '@core/api/courses-api.service';
import { ExternalRequestsApiService } from '@core/api/external-requests-api.service';
import { RolesApiService } from '@core/api/roles-api.service';
import { UsersApiService } from '@core/api/users-api.service';
import { AuthStateService } from '@core/auth/auth-state.service';
import type { IdentityUserView } from '@core/auth/auth.types';
import { PERMISSIONS } from '@core/auth/permissions';
import type { Course } from '@entities/course';

const DEFAULT_LIST_LIMIT = 200;

@Injectable({ providedIn: 'root' })
export class AdminFacade {
  private readonly authState = inject(AuthStateService);
  private readonly usersApi = inject(UsersApiService);
  private readonly rolesApi = inject(RolesApiService);
  private readonly coursesApi = inject(CoursesApiService);
  private readonly externalRequestsApi = inject(ExternalRequestsApiService);
  private readonly assignmentsApi = inject(AssignmentsApiService);
  private readonly approvalWorkflowsApi = inject(ApprovalWorkflowsApiService);
  private readonly budgetLimitsApi = inject(BudgetLimitsApiService);
  private readonly analyticsApi = inject(AnalyticsApiService);

  readonly loading = signal(false);
  readonly busyKey = signal<string | null>(null);
  readonly error = signal<string | null>(null);
  readonly notice = signal<string | null>(null);

  readonly users = signal<IdentityUserView[]>([]);
  readonly roles = signal<AdminRole[]>([]);
  readonly courses = signal<Course[]>([]);
  readonly externalRequests = signal<ExternalRequest[]>([]);
  readonly pendingApprovals = signal<PendingApprovalItem[]>([]);
  readonly assignments = signal<CourseAssignment[]>([]);
  readonly workflows = signal<ApprovalWorkflow[]>([]);
  readonly budgetLimits = signal<BudgetLimitRecord[]>([]);
  readonly analytics = signal<HrDashboardStats | null>(null);

  readonly currentUser = this.authState.currentUser;
  readonly canReadUsers = computed(() => this.authState.hasPermission(PERMISSIONS.usersRead));
  readonly canManageRoles = computed(() => this.authState.hasPermission(PERMISSIONS.rolesManage));
  readonly canManageCourses = computed(() =>
    this.authState.hasPermission(PERMISSIONS.coursesWrite),
  );
  readonly canAssignCourses = computed(() =>
    this.authState.hasPermission(PERMISSIONS.coursesAssign),
  );
  readonly canApproveRequests = computed(() =>
    this.authState.hasAnyPermission([
      PERMISSIONS.externalRequestsApproveHr,
      PERMISSIONS.externalRequestsApproveManager,
    ]),
  );
  readonly canSeeAnalytics = computed(() =>
    this.authState.hasPermission(PERMISSIONS.analyticsReadHr),
  );
  readonly canManageSettings = computed(() =>
    this.authState.hasPermission(PERMISSIONS.settingsManage),
  );

  readonly totalEmployees = computed(() => this.users().length);
  readonly adminCount = computed(
    () => this.users().filter((user) => user.roles.includes('admin')).length,
  );
  readonly hrCount = computed(
    () => this.users().filter((user) => user.roles.includes('hr')).length,
  );
  readonly publishedCoursesCount = computed(
    () =>
      this.courses().filter((course) => course.status === 'published' && !course.archived_at)
        .length,
  );
  readonly draftCoursesCount = computed(
    () => this.courses().filter((course) => course.status === 'draft').length,
  );
  readonly hiddenCoursesCount = computed(
    () =>
      this.courses().filter((course) => course.status === 'archived' || !!course.archived_at)
        .length,
  );
  readonly workflowCount = computed(() => this.workflows().length);
  readonly activeBudgetLimitCount = computed(
    () => this.budgetLimits().filter((item) => item.is_active).length,
  );
  readonly roleCatalog = computed(() => new Map(this.roles().map((role) => [role.code, role])));
  readonly knownRoleCodes = computed(() => {
    const catalogCodes = this.roles().map((role) => role.code);
    const userCodes = this.users().flatMap((user) => user.roles);
    return Array.from(new Set([...catalogCodes, ...userCodes])).sort();
  });

  load(): void {
    void this.reload();
  }

  async refresh(): Promise<void> {
    await this.reload();
  }

  clearMessage(): void {
    this.notice.set(null);
  }

  clearError(): void {
    this.error.set(null);
  }

  isBusy(key: string): boolean {
    return this.loading() || this.busyKey() === key;
  }

  async toggleUserRole(user: IdentityUserView, roleCode: string): Promise<void> {
    if (!this.canManageRoles()) {
      this.error.set('Назначение ролей доступно только пользователям с правом roles.manage.');
      return;
    }

    const role = this.roleCatalog().get(roleCode);
    if (!role) {
      this.error.set('Справочник ролей ещё не загрузился.');
      return;
    }

    const hasRole = user.roles.includes(roleCode);
    const key = `role:${user.id}:${roleCode}`;
    const message = hasRole
      ? `Роль ${role.name} снята у пользователя.`
      : `Роль ${role.name} назначена пользователю.`;

    await this.runAction(key, message, async () => {
      if (hasRole) {
        await firstValueFrom(this.usersApi.removeRole(user.id, role.id));
      } else {
        await firstValueFrom(this.usersApi.assignRole(user.id, { role_code: roleCode }));
      }

      await this.loadUsers();
    });
  }

  async saveCourse(courseId: string | null, payload: Record<string, unknown>): Promise<void> {
    if (!this.canManageCourses()) {
      this.error.set('У текущей роли нет прав на управление каталогом.');
      return;
    }

    const key = courseId ? `course:update:${courseId}` : 'course:create';
    const message = courseId ? 'Курс обновлён.' : 'Новый курс добавлен как черновик.';

    await this.runAction(key, message, async () => {
      if (courseId) {
        await firstValueFrom(this.coursesApi.update(courseId, payload));
      } else {
        await firstValueFrom(this.coursesApi.create(payload));
      }

      await Promise.all([this.loadCourses(), this.loadAnalytics()]);
    });
  }

  async publishCourse(id: string): Promise<void> {
    await this.runAction(
      `course:publish:${id}`,
      'Курс опубликован и виден в каталоге.',
      async () => {
        await firstValueFrom(this.coursesApi.publish(id));
        await Promise.all([this.loadCourses(), this.loadAnalytics()]);
      },
    );
  }

  async archiveCourse(id: string): Promise<void> {
    await this.runAction(`course:archive:${id}`, 'Курс скрыт из каталога.', async () => {
      await firstValueFrom(this.coursesApi.archive(id));
      await Promise.all([this.loadCourses(), this.loadAnalytics()]);
    });
  }

  async createAssignment(payload: Record<string, unknown>): Promise<void> {
    if (!this.canAssignCourses()) {
      this.error.set('Назначение обучения недоступно для текущей роли.');
      return;
    }

    await this.runAction('assignment:create', 'Обучение назначено сотруднику.', async () => {
      await firstValueFrom(this.assignmentsApi.create(payload));
      await this.loadAssignments();
    });
  }

  async approveRequest(id: string, comment?: string): Promise<void> {
    await this.handleExternalRequestDecision('approve', id, comment ?? '');
  }

  async rejectRequest(id: string, comment?: string): Promise<void> {
    await this.handleExternalRequestDecision('reject', id, comment ?? '');
  }

  async requestRevision(id: string, comment?: string): Promise<void> {
    await this.handleExternalRequestDecision('request-revision', id, comment ?? '');
  }

  async createBudgetLimit(payload: Record<string, unknown>): Promise<void> {
    await this.runAction('budget:create', 'Лимит бюджета сохранён.', async () => {
      await firstValueFrom(this.budgetLimitsApi.create(payload));
      await this.loadBudgetLimits();
    });
  }

  async createWorkflow(payload: Record<string, unknown>): Promise<void> {
    await this.runAction('workflow:create', 'Маршрут согласования сохранён.', async () => {
      await firstValueFrom(this.approvalWorkflowsApi.create(payload));
      await this.loadWorkflows();
    });
  }

  private async reload(): Promise<void> {
    this.loading.set(true);
    this.error.set(null);

    const results = await Promise.all([
      this.loadUsers(),
      this.loadRoles(),
      this.loadCourses(),
      this.loadExternalRequests(),
      this.loadPendingApprovals(),
      this.loadAssignments(),
      this.loadWorkflows(),
      this.loadBudgetLimits(),
      this.loadAnalytics(),
    ]);

    if (results.some((failed) => failed)) {
      this.error.set(
        'Часть данных для HR/Admin панели не загрузилась. Попробуйте обновить страницу.',
      );
    }

    this.loading.set(false);
  }

  private async loadUsers(): Promise<boolean> {
    if (!this.canReadUsers()) {
      this.users.set([]);
      return false;
    }

    return this.loadSection(
      () => firstValueFrom(this.usersApi.listAdminUsers()),
      (items) => {
        const sorted = [...items].sort((left, right) =>
          this.displayName(left).localeCompare(this.displayName(right), 'ru'),
        );
        this.users.set(sorted);
      },
      [],
    );
  }

  private async loadRoles(): Promise<boolean> {
    if (!this.canManageRoles()) {
      this.roles.set([]);
      return false;
    }

    return this.loadSection(
      () => firstValueFrom(this.rolesApi.list()),
      (items) => this.roles.set(items),
      [],
    );
  }

  private async loadCourses(): Promise<boolean> {
    return this.loadSection(
      () =>
        firstValueFrom(
          this.coursesApi.list({
            limit: DEFAULT_LIST_LIMIT,
            offset: 0,
            sort: 'updated',
          }),
        ),
      (items) => this.courses.set(items),
      [],
    );
  }

  private async loadExternalRequests(): Promise<boolean> {
    return this.loadSection(
      () =>
        firstValueFrom(
          this.externalRequestsApi.list({
            scope: 'all',
            limit: DEFAULT_LIST_LIMIT,
            offset: 0,
          }),
        ),
      (items) => this.externalRequests.set(items),
      [],
    );
  }

  private async loadPendingApprovals(): Promise<boolean> {
    if (!this.canApproveRequests()) {
      this.pendingApprovals.set([]);
      return false;
    }

    return this.loadSection(
      () =>
        firstValueFrom(
          this.externalRequestsApi.listPendingApprovals({
            limit: 50,
            offset: 0,
          }),
        ),
      (items) => this.pendingApprovals.set(items),
      [],
    );
  }

  private async loadAssignments(): Promise<boolean> {
    if (!this.canAssignCourses()) {
      this.assignments.set([]);
      return false;
    }

    return this.loadSection(
      () => firstValueFrom(this.assignmentsApi.list()),
      (items) => this.assignments.set(items),
      [],
    );
  }

  private async loadWorkflows(): Promise<boolean> {
    if (!this.canManageSettings()) {
      this.workflows.set([]);
      return false;
    }

    return this.loadSection(
      () => firstValueFrom(this.approvalWorkflowsApi.list()),
      (items) => this.workflows.set(items),
      [],
    );
  }

  private async loadBudgetLimits(): Promise<boolean> {
    if (!this.canManageSettings()) {
      this.budgetLimits.set([]);
      return false;
    }

    return this.loadSection(
      () => firstValueFrom(this.budgetLimitsApi.list()),
      (items) => this.budgetLimits.set(items),
      [],
    );
  }

  private async loadAnalytics(): Promise<boolean> {
    if (!this.canSeeAnalytics()) {
      this.analytics.set(null);
      return false;
    }

    return this.loadSection(
      () => firstValueFrom(this.analyticsApi.getHrDashboard() as any),
      (payload) => this.analytics.set(payload as HrDashboardStats),
      null,
    );
  }

  private async handleExternalRequestDecision(
    action: 'approve' | 'reject' | 'request-revision',
    id: string,
    comment: string,
  ): Promise<void> {
    if (!this.canApproveRequests()) {
      this.error.set('У текущей роли нет прав на согласование заявок.');
      return;
    }

    const labelMap: Record<typeof action, string> = {
      approve: 'Заявка согласована.',
      reject: 'Заявка отклонена.',
      'request-revision': 'Заявка отправлена на доработку.',
    };

    await this.runAction(`request:${action}:${id}`, labelMap[action], async () => {
      const payload = comment.trim() ? { comment: comment.trim() } : {};
      if (action === 'approve') {
        await firstValueFrom(this.externalRequestsApi.approve(id, payload));
      } else if (action === 'reject') {
        await firstValueFrom(this.externalRequestsApi.reject(id, payload));
      } else {
        await firstValueFrom(this.externalRequestsApi.requestRevision(id, payload));
      }

      await Promise.all([
        this.loadExternalRequests(),
        this.loadPendingApprovals(),
        this.loadAnalytics(),
      ]);
    });
  }

  private async runAction(
    key: string,
    successMessage: string,
    task: () => Promise<void>,
  ): Promise<void> {
    this.busyKey.set(key);
    this.error.set(null);
    this.notice.set(null);

    try {
      await task();
      this.notice.set(successMessage);
    } catch (error) {
      this.error.set(this.describeError(error));
    } finally {
      this.busyKey.set(null);
    }
  }

  private async loadSection<T>(
    loader: () => Promise<T>,
    commit: (value: T) => void,
    fallback: T,
  ): Promise<boolean> {
    try {
      commit(await loader());
      return false;
    } catch {
      commit(fallback);
      return true;
    }
  }

  private describeError(error: unknown): string {
    if (error instanceof HttpErrorResponse) {
      const payload = error.error as { message?: string } | null;
      if (payload?.message) {
        return payload.message;
      }
      if (error.status === 403) {
        return 'Backend отклонил действие из-за недостатка прав.';
      }
    }

    return 'Операция не выполнилась. Проверьте backend и повторите попытку.';
  }

  private displayName(user: IdentityUserView): string {
    const profile = user.employee_profile;
    if (!profile) {
      return user.email;
    }

    return [profile.last_name, profile.first_name, profile.middle_name]
      .filter((item) => !!item)
      .join(' ')
      .trim();
  }
}
