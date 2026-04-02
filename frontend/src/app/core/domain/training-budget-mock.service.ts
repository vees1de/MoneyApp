import { Injectable, computed, signal } from '@angular/core';

import type { ExternalRequest } from '@core/api/contracts';
import type {
  BudgetLimit,
  BudgetUsageStatus,
  CourseSpendReportRow,
  DepartmentBudgetReportRow,
  DepartmentTrainingBudget,
  EmployeeBudgetReportRow,
  EmployeeTrainingQuota,
  TrackedTrainingRequest,
  TrainingBudgetCheckStatus,
  TrainingBudgetPreview,
  TrainingBudgetRequestView,
  TrainingBudgetValidation,
  TrainingExpense,
  TrainingExpenseStatus,
} from '@entities/budget-limit';

interface BudgetActorContext {
  departmentId: string;
  departmentName: string;
  employeeId: string;
  employeeName: string;
  employeeEmail?: string | null;
}

interface EmployeeSeed {
  id: string;
  name: string;
  email?: string | null;
  quotaAmount: number;
}

interface DepartmentSeed {
  id: string;
  name: string;
  totalBudget: number;
  currency: string;
  employees: EmployeeSeed[];
}

interface BudgetStoreState {
  departments: DepartmentSeed[];
  expenses: TrainingExpense[];
  trackedRequests: TrackedTrainingRequest[];
}

const DEFAULT_CURRENCY = 'RUB';
const DEFAULT_DEPARTMENT_BUDGET = 960_000;
const DEFAULT_EMPLOYEE_QUOTA = 180_000;
const DEFAULT_DEPARTMENT_NAME = 'Мой отдел';
const DEFAULT_EMPLOYEE_NAME = 'Сотрудник';
const DEFAULT_REQUEST_TITLE = 'Заявка на внешний курс';

const ACTIVE_EXPENSE_STATUSES: readonly TrainingExpenseStatus[] = ['reserved', 'approved', 'completed'];
const SPENT_EXPENSE_STATUSES: readonly TrainingExpenseStatus[] = ['approved', 'completed'];
const RESERVED_EXPENSE_STATUSES: readonly TrainingExpenseStatus[] = ['reserved'];

const INITIAL_DEPARTMENTS: DepartmentSeed[] = [
  {
    id: 'dept-product',
    name: 'Product',
    totalBudget: 920_000,
    currency: DEFAULT_CURRENCY,
    employees: [
      { id: 'emp-irina', name: 'Irina Volkova', email: 'irina.volkova@moneyapp.local', quotaAmount: 180_000 },
      { id: 'emp-pavel', name: 'Pavel Smirnov', email: 'pavel.smirnov@moneyapp.local', quotaAmount: 150_000 },
      { id: 'emp-anna', name: 'Anna Petrova', email: 'anna.petrova@moneyapp.local', quotaAmount: 170_000 },
    ],
  },
  {
    id: 'dept-ops',
    name: 'Operations',
    totalBudget: 760_000,
    currency: DEFAULT_CURRENCY,
    employees: [
      { id: 'emp-elena', name: 'Elena Sokolova', email: 'elena.sokolova@moneyapp.local', quotaAmount: 140_000 },
      { id: 'emp-roman', name: 'Roman Egorov', email: 'roman.egorov@moneyapp.local', quotaAmount: 135_000 },
      { id: 'emp-sergey', name: 'Sergey Kozhin', email: 'sergey.kozhin@moneyapp.local', quotaAmount: 125_000 },
    ],
  },
  {
    id: 'dept-hr',
    name: 'HR',
    totalBudget: 540_000,
    currency: DEFAULT_CURRENCY,
    employees: [
      { id: 'emp-olga', name: 'Olga Kim', email: 'olga.kim@moneyapp.local', quotaAmount: 160_000 },
      { id: 'emp-nina', name: 'Nina Kovaleva', email: 'nina.kovaleva@moneyapp.local', quotaAmount: 120_000 },
    ],
  },
];

const INITIAL_EXPENSES: TrainingExpense[] = [
  {
    id: 'expense-1',
    requestId: 'seed-req-1',
    departmentId: 'dept-product',
    departmentName: 'Product',
    employeeId: 'emp-irina',
    employeeName: 'Irina Volkova',
    employeeEmail: 'irina.volkova@moneyapp.local',
    courseTitle: 'Product Leadership Sprint',
    providerName: 'Skillbox',
    amount: 62_000,
    currency: DEFAULT_CURRENCY,
    status: 'approved',
    budgetCheckStatus: 'within_budget',
    requiresAdditionalApproval: false,
    createdAt: '2026-03-03T10:00:00.000Z',
    updatedAt: '2026-03-06T10:00:00.000Z',
  },
  {
    id: 'expense-2',
    requestId: 'seed-req-2',
    departmentId: 'dept-product',
    departmentName: 'Product',
    employeeId: 'emp-pavel',
    employeeName: 'Pavel Smirnov',
    employeeEmail: 'pavel.smirnov@moneyapp.local',
    courseTitle: 'Advanced SQL for PMs',
    providerName: 'Practicum',
    amount: 48_000,
    currency: DEFAULT_CURRENCY,
    status: 'reserved',
    budgetCheckStatus: 'within_budget',
    requiresAdditionalApproval: false,
    createdAt: '2026-03-10T12:00:00.000Z',
    updatedAt: '2026-03-10T12:00:00.000Z',
  },
  {
    id: 'expense-3',
    requestId: 'seed-req-3',
    departmentId: 'dept-ops',
    departmentName: 'Operations',
    employeeId: 'emp-elena',
    employeeName: 'Elena Sokolova',
    employeeEmail: 'elena.sokolova@moneyapp.local',
    courseTitle: 'Lean Operations Lab',
    providerName: 'Netology',
    amount: 71_000,
    currency: DEFAULT_CURRENCY,
    status: 'completed',
    budgetCheckStatus: 'within_budget',
    requiresAdditionalApproval: false,
    createdAt: '2026-02-18T09:00:00.000Z',
    updatedAt: '2026-03-01T09:00:00.000Z',
  },
  {
    id: 'expense-4',
    requestId: 'seed-req-4',
    departmentId: 'dept-ops',
    departmentName: 'Operations',
    employeeId: 'emp-roman',
    employeeName: 'Roman Egorov',
    employeeEmail: 'roman.egorov@moneyapp.local',
    courseTitle: 'Negotiation Mastery',
    providerName: 'Eduson',
    amount: 56_000,
    currency: DEFAULT_CURRENCY,
    status: 'cancelled',
    budgetCheckStatus: 'within_budget',
    requiresAdditionalApproval: false,
    approvalComment: 'Course was cancelled by the employee, budget returned.',
    createdAt: '2026-02-07T08:30:00.000Z',
    updatedAt: '2026-02-10T14:30:00.000Z',
  },
  {
    id: 'expense-5',
    requestId: 'seed-req-5',
    departmentId: 'dept-hr',
    departmentName: 'HR',
    employeeId: 'emp-olga',
    employeeName: 'Olga Kim',
    employeeEmail: 'olga.kim@moneyapp.local',
    courseTitle: 'People Analytics Intensive',
    providerName: 'Yandex Practicum',
    amount: 84_000,
    currency: DEFAULT_CURRENCY,
    status: 'approved',
    budgetCheckStatus: 'within_budget',
    requiresAdditionalApproval: false,
    createdAt: '2026-03-08T07:00:00.000Z',
    updatedAt: '2026-03-13T07:00:00.000Z',
  },
  {
    id: 'expense-6',
    requestId: 'seed-req-6',
    departmentId: 'dept-hr',
    departmentName: 'HR',
    employeeId: 'emp-nina',
    employeeName: 'Nina Kovaleva',
    employeeEmail: 'nina.kovaleva@moneyapp.local',
    courseTitle: 'HRBP Finance Essentials',
    providerName: 'Skill Cup',
    amount: 96_000,
    currency: DEFAULT_CURRENCY,
    status: 'revision_returned',
    budgetCheckStatus: 'employee_quota_exceeded',
    requiresAdditionalApproval: true,
    approvalComment: 'Returned after budget review, amount released back to the quota.',
    createdAt: '2026-03-16T11:00:00.000Z',
    updatedAt: '2026-03-18T11:00:00.000Z',
  },
];

@Injectable({ providedIn: 'root' })
export class TrainingBudgetMockService {
  private readonly state = signal<BudgetStoreState>({
    departments: INITIAL_DEPARTMENTS,
    expenses: INITIAL_EXPENSES,
    trackedRequests: [],
  });

  readonly departments = computed<DepartmentTrainingBudget[]>(() =>
    this.state()
      .departments.map((department) => this.buildDepartmentBudget(department))
      .sort((left, right) => left.departmentName.localeCompare(right.departmentName)),
  );

  readonly history = computed<TrainingExpense[]>(() =>
    [...this.state().expenses].sort((left, right) => right.updatedAt.localeCompare(left.updatedAt)),
  );

  readonly departmentReport = computed<DepartmentBudgetReportRow[]>(() =>
    this.departments().map((department) => ({
      departmentId: department.departmentId,
      departmentName: department.departmentName,
      totalBudget: department.totalBudget,
      spentAmount: department.spentAmount,
      reservedAmount: department.reservedAmount,
      usedAmount: department.usedAmount,
      remainingAmount: department.remainingAmount,
      usagePercent: department.usagePercent,
      status: department.status,
      headcount: department.employees.length,
      activeRequests: department.expenses.filter((expense) => isActiveExpense(expense.status)).length,
      currency: department.currency,
    })),
  );

  readonly employeeReport = computed<EmployeeBudgetReportRow[]>(() =>
    this.departments()
      .flatMap((department) =>
        department.employees.map((employee) => ({
          employeeId: employee.employeeId,
          employeeName: employee.employeeName,
          departmentId: department.departmentId,
          departmentName: department.departmentName,
          quotaAmount: employee.quotaAmount,
          spentAmount: employee.spentAmount,
          reservedAmount: employee.reservedAmount,
          usedAmount: employee.usedAmount,
          remainingAmount: employee.remainingAmount,
          usagePercent: employee.usagePercent,
          status: employee.status,
          activeRequests: department.expenses.filter(
            (expense) => expense.employeeId === employee.employeeId && isActiveExpense(expense.status),
          ).length,
        })),
      )
      .sort((left, right) => right.usedAmount - left.usedAmount),
  );

  readonly topCourses = computed<CourseSpendReportRow[]>(() => {
    const report = new Map<string, CourseSpendReportRow & { employees: Set<string> }>();

    for (const expense of this.state().expenses) {
      if (!isActiveExpense(expense.status)) {
        continue;
      }

      const key = expense.courseTitle.trim().toLowerCase();
      const existing = report.get(key);
      if (existing) {
        existing.totalAmount += expense.amount;
        existing.requestCount += 1;
        existing.employees.add(expense.employeeId);
        continue;
      }

      report.set(key, {
        courseTitle: expense.courseTitle,
        totalAmount: expense.amount,
        requestCount: 1,
        uniqueEmployees: 1,
        currency: expense.currency,
        employees: new Set([expense.employeeId]),
      });
    }

    return [...report.values()]
      .map((item) => ({
        courseTitle: item.courseTitle,
        totalAmount: item.totalAmount,
        requestCount: item.requestCount,
        uniqueEmployees: item.employees.size,
        currency: item.currency,
      }))
      .sort((left, right) => right.totalAmount - left.totalAmount)
      .slice(0, 6);
  });

  readonly snapshot = computed<BudgetLimit>(() => ({
    departments: this.departments(),
    departmentReport: this.departmentReport(),
    employeeReport: this.employeeReport(),
    topCourses: this.topCourses(),
    history: this.history(),
  }));

  ensureActorContext(input: {
    departmentId?: string | null;
    departmentName?: string | null;
    employeeId?: string | null;
    employeeName?: string | null;
    employeeEmail?: string | null;
  }): BudgetActorContext {
    const normalizedDepartmentId = sanitizeId(input.departmentId, 'dept');
    const normalizedDepartmentName = sanitizeText(input.departmentName, DEFAULT_DEPARTMENT_NAME);
    const normalizedEmployeeId = sanitizeId(input.employeeId, 'emp');
    const normalizedEmployeeName = sanitizeText(input.employeeName, DEFAULT_EMPLOYEE_NAME);
    const normalizedEmployeeEmail = input.employeeEmail?.trim() || null;

    let department = this.findDepartment(normalizedDepartmentId, normalizedDepartmentName);
    if (!department) {
      const newDept: DepartmentSeed = {
        id: normalizedDepartmentId,
        name: normalizedDepartmentName,
        totalBudget: DEFAULT_DEPARTMENT_BUDGET,
        currency: DEFAULT_CURRENCY,
        employees: [],
      };
      department = newDept;
      this.state.update((state) => ({
        ...state,
        departments: [...state.departments, newDept],
      }));
    }

    const employeeExists = department.employees.some((employee) =>
      matchesEmployee(employee, normalizedEmployeeId, normalizedEmployeeName, normalizedEmployeeEmail),
    );

    if (!employeeExists) {
      const nextEmployee: EmployeeSeed = {
        id: normalizedEmployeeId,
        name: normalizedEmployeeName,
        email: normalizedEmployeeEmail,
        quotaAmount: DEFAULT_EMPLOYEE_QUOTA,
      };

      this.state.update((state) => ({
        ...state,
        departments: state.departments.map((item) =>
          item.id === department.id
            ? { ...item, employees: [...item.employees, nextEmployee] }
            : item,
        ),
      }));
    }

    return {
      departmentId: department.id,
      departmentName: department.name,
      employeeId: normalizedEmployeeId,
      employeeName: normalizedEmployeeName,
      employeeEmail: normalizedEmployeeEmail,
    };
  }

  previewForActor(
    actor: BudgetActorContext,
    amount: number,
    currency = DEFAULT_CURRENCY,
  ): TrainingBudgetPreview {
    return {
      department: this.getDepartmentBudget(actor.departmentId),
      employee: this.getEmployeeQuota(actor.employeeId),
      validation: this.buildValidation(actor, amount, currency),
    };
  }

  storeDraftRequest(input: {
    requestId: string;
    title: string;
    providerName?: string | null;
    amount: number;
    currency?: string | null;
    departmentId?: string | null;
    departmentName?: string | null;
    employeeId?: string | null;
    employeeName?: string | null;
    employeeEmail?: string | null;
    requiresAdditionalApproval?: boolean;
    additionalApprovalComment?: string | null;
    lastKnownStatus?: string;
  }): TrackedTrainingRequest {
    const existing = this.state().trackedRequests.find((item) => item.requestId === input.requestId);
    const actor = this.ensureActorContext(input);
    const tracked: TrackedTrainingRequest = {
      requestId: input.requestId,
      title: sanitizeText(input.title || existing?.title, DEFAULT_REQUEST_TITLE),
      providerName: input.providerName?.trim() || existing?.providerName || null,
      amount: sanitizeAmount(input.amount),
      currency: sanitizeText(input.currency || existing?.currency, DEFAULT_CURRENCY),
      departmentId: actor.departmentId,
      departmentName: actor.departmentName,
      employeeId: actor.employeeId,
      employeeName: actor.employeeName,
      employeeEmail: actor.employeeEmail,
      requiresAdditionalApproval:
        input.requiresAdditionalApproval ?? existing?.requiresAdditionalApproval ?? false,
      additionalApprovalComment:
        input.additionalApprovalComment?.trim() || existing?.additionalApprovalComment || null,
      lastKnownStatus: input.lastKnownStatus ?? existing?.lastKnownStatus ?? 'draft',
      updatedAt: new Date().toISOString(),
    };

    this.upsertTrackedRequest(tracked);
    return tracked;
  }

  syncRequest(
    request: ExternalRequest,
    options?: {
      departmentId?: string | null;
      departmentName?: string | null;
      employeeId?: string | null;
      employeeName?: string | null;
      employeeEmail?: string | null;
      requiresAdditionalApproval?: boolean;
      additionalApprovalComment?: string | null;
    },
  ): void {
    const tracked = this.storeDraftRequest({
      requestId: request.id,
      title: request.title,
      providerName: request.provider_name,
      amount: sanitizeAmount(request.cost_amount),
      currency: request.currency,
      departmentId: options?.departmentId ?? request.department_id,
      departmentName: options?.departmentName ?? request.department_name,
      employeeId: options?.employeeId ?? request.employee_user_id,
      employeeName: options?.employeeName ?? request.employee_full_name ?? request.employee_email,
      employeeEmail: options?.employeeEmail ?? request.employee_email,
      requiresAdditionalApproval: options?.requiresAdditionalApproval,
      additionalApprovalComment: options?.additionalApprovalComment,
      lastKnownStatus: request.status,
    });

    switch (request.status) {
      case 'manager_approval':
      case 'hr_approval':
        this.upsertExpenseForTrackedRequest(tracked, 'reserved');
        break;
      case 'approved':
        this.upsertExpenseForTrackedRequest(tracked, 'approved');
        break;
      case 'completed':
        this.upsertExpenseForTrackedRequest(tracked, 'completed');
        break;
      case 'rejected':
        this.returnFunds(request.id, 'rejected', 'Request was rejected, budget released.');
        break;
      case 'revision_requested':
        this.returnFunds(
          request.id,
          'revision_returned',
          'Request was returned for revision, budget released.',
        );
        break;
      default:
        break;
    }
  }

  getRequestView(requestId: string): TrainingBudgetRequestView | null {
    const tracked = this.state().trackedRequests.find((item) => item.requestId === requestId);
    if (!tracked) {
      return null;
    }

    const actor: BudgetActorContext = {
      departmentId: tracked.departmentId,
      departmentName: tracked.departmentName,
      employeeId: tracked.employeeId,
      employeeName: tracked.employeeName,
      employeeEmail: tracked.employeeEmail,
    };

    const department = this.getDepartmentBudget(tracked.departmentId);
    const employee = this.getEmployeeQuota(tracked.employeeId);
    const latestExpense = this.getLatestExpense(requestId);

    return {
      trackedRequest: tracked,
      department,
      employee,
      validation: this.buildValidation(actor, tracked.amount, tracked.currency, requestId),
      latestExpense,
      history: department?.expenses.slice(0, 6) ?? [],
    };
  }

  markCourseCancelled(requestId: string, comment?: string | null): void {
    this.returnFunds(requestId, 'cancelled', comment ?? 'Training was cancelled, budget released.');
  }

  hasActiveAllocation(requestId: string): boolean {
    return this.state().expenses.some(
      (expense) => expense.requestId === requestId && isActiveExpense(expense.status),
    );
  }

  getDepartmentBudget(departmentId: string): DepartmentTrainingBudget | null {
    return this.departments().find((department) => department.departmentId === departmentId) ?? null;
  }

  getEmployeeQuota(employeeId: string): EmployeeTrainingQuota | null {
    for (const department of this.departments()) {
      const employee = department.employees.find((item) => item.employeeId === employeeId);
      if (employee) {
        return employee;
      }
    }

    return null;
  }

  private upsertTrackedRequest(tracked: TrackedTrainingRequest): void {
    this.state.update((state) => {
      const index = state.trackedRequests.findIndex((item) => item.requestId === tracked.requestId);
      if (index === -1) {
        return {
          ...state,
          trackedRequests: [...state.trackedRequests, tracked],
        };
      }

      const nextTrackedRequests = [...state.trackedRequests];
      nextTrackedRequests[index] = {
        ...nextTrackedRequests[index],
        ...tracked,
        updatedAt: tracked.updatedAt,
      };

      return {
        ...state,
        trackedRequests: nextTrackedRequests,
      };
    });
  }

  private upsertExpenseForTrackedRequest(
    tracked: TrackedTrainingRequest,
    status: Extract<TrainingExpenseStatus, 'reserved' | 'approved' | 'completed'>,
  ): void {
    const now = new Date().toISOString();
    const latestExpense = this.getLatestExpense(tracked.requestId);

    if (latestExpense && isActiveExpense(latestExpense.status)) {
      this.state.update((state) => ({
        ...state,
        expenses: state.expenses.map((expense) =>
          expense.id === latestExpense.id
            ? {
                ...expense,
                departmentId: tracked.departmentId,
                departmentName: tracked.departmentName,
                employeeId: tracked.employeeId,
                employeeName: tracked.employeeName,
                employeeEmail: tracked.employeeEmail,
                courseTitle: tracked.title,
                providerName: tracked.providerName,
                amount: tracked.amount,
                currency: tracked.currency,
                status,
                requiresAdditionalApproval: tracked.requiresAdditionalApproval,
                approvalComment: tracked.additionalApprovalComment,
                updatedAt: now,
              }
            : expense,
        ),
      }));
      return;
    }

    const validation = this.buildValidation(
      {
        departmentId: tracked.departmentId,
        departmentName: tracked.departmentName,
        employeeId: tracked.employeeId,
        employeeName: tracked.employeeName,
        employeeEmail: tracked.employeeEmail,
      },
      tracked.amount,
      tracked.currency,
    );

    const nextExpense: TrainingExpense = {
      id: buildId('expense'),
      requestId: tracked.requestId,
      departmentId: tracked.departmentId,
      departmentName: tracked.departmentName,
      employeeId: tracked.employeeId,
      employeeName: tracked.employeeName,
      employeeEmail: tracked.employeeEmail,
      courseTitle: tracked.title,
      providerName: tracked.providerName,
      amount: tracked.amount,
      currency: tracked.currency,
      status,
      budgetCheckStatus: validation.status,
      requiresAdditionalApproval:
        tracked.requiresAdditionalApproval || validation.requiresAdditionalApproval,
      approvalComment: tracked.additionalApprovalComment,
      createdAt: now,
      updatedAt: now,
    };

    this.state.update((state) => ({
      ...state,
      expenses: [...state.expenses, nextExpense],
    }));
  }

  private returnFunds(
    requestId: string,
    status: Extract<TrainingExpenseStatus, 'rejected' | 'revision_returned' | 'cancelled'>,
    approvalComment: string,
  ): void {
    const latestExpense = this.getLatestExpense(requestId);
    if (!latestExpense || !isActiveExpense(latestExpense.status)) {
      return;
    }

    const now = new Date().toISOString();
    this.state.update((state) => ({
      ...state,
      expenses: state.expenses.map((expense) =>
        expense.id === latestExpense.id
          ? {
              ...expense,
              status,
              approvalComment,
              updatedAt: now,
            }
          : expense,
      ),
    }));
  }

  private buildDepartmentBudget(seed: DepartmentSeed): DepartmentTrainingBudget {
    const departmentExpenses = this.history()
      .filter((expense) => expense.departmentId === seed.id)
      .sort((left, right) => right.updatedAt.localeCompare(left.updatedAt));

    const spentAmount = sumAmounts(
      departmentExpenses.filter((expense) => SPENT_EXPENSE_STATUSES.includes(expense.status)),
    );
    const reservedAmount = sumAmounts(
      departmentExpenses.filter((expense) => RESERVED_EXPENSE_STATUSES.includes(expense.status)),
    );
    const usedAmount = spentAmount + reservedAmount;

    return {
      departmentId: seed.id,
      departmentName: seed.name,
      totalBudget: seed.totalBudget,
      spentAmount,
      reservedAmount,
      usedAmount,
      remainingAmount: seed.totalBudget - usedAmount,
      usagePercent: calculatePercent(usedAmount, seed.totalBudget),
      currency: seed.currency,
      status: getUsageStatus(usedAmount, seed.totalBudget),
      employees: seed.employees
        .map((employee) => this.buildEmployeeQuota(seed, employee, departmentExpenses))
        .sort((left, right) => right.usedAmount - left.usedAmount),
      expenses: departmentExpenses,
    };
  }

  private buildEmployeeQuota(
    department: DepartmentSeed,
    employee: EmployeeSeed,
    departmentExpenses: TrainingExpense[],
  ): EmployeeTrainingQuota {
    const employeeExpenses = departmentExpenses.filter((expense) => expense.employeeId === employee.id);
    const spentAmount = sumAmounts(
      employeeExpenses.filter((expense) => SPENT_EXPENSE_STATUSES.includes(expense.status)),
    );
    const reservedAmount = sumAmounts(
      employeeExpenses.filter((expense) => RESERVED_EXPENSE_STATUSES.includes(expense.status)),
    );
    const usedAmount = spentAmount + reservedAmount;

    return {
      employeeId: employee.id,
      employeeName: employee.name,
      employeeEmail: employee.email,
      departmentId: department.id,
      departmentName: department.name,
      quotaAmount: employee.quotaAmount,
      spentAmount,
      reservedAmount,
      usedAmount,
      remainingAmount: employee.quotaAmount - usedAmount,
      usagePercent: calculatePercent(usedAmount, employee.quotaAmount),
      status: getUsageStatus(usedAmount, employee.quotaAmount),
    };
  }

  private buildValidation(
    actor: BudgetActorContext,
    amount: number,
    currency: string,
    excludeRequestId?: string,
  ): TrainingBudgetValidation {
    const sanitizedAmount = sanitizeAmount(amount);
    const departmentSeed = this.state().departments.find((department) => department.id === actor.departmentId);
    const employeeSeed = departmentSeed?.employees.find((employee) => employee.id === actor.employeeId);
    const relevantExpenses = this.state().expenses.filter(
      (expense) => expense.requestId !== excludeRequestId,
    );

    const departmentUsedAmount = sumAmounts(
      relevantExpenses.filter(
        (expense) => expense.departmentId === actor.departmentId && isActiveExpense(expense.status),
      ),
    );
    const employeeUsedAmount = sumAmounts(
      relevantExpenses.filter(
        (expense) => expense.employeeId === actor.employeeId && isActiveExpense(expense.status),
      ),
    );

    const departmentRemainingAmount = (departmentSeed?.totalBudget ?? DEFAULT_DEPARTMENT_BUDGET) - departmentUsedAmount;
    const employeeRemainingAmount = (employeeSeed?.quotaAmount ?? DEFAULT_EMPLOYEE_QUOTA) - employeeUsedAmount;

    const employeeEnough = employeeRemainingAmount >= sanitizedAmount;
    const departmentEnough = departmentRemainingAmount >= sanitizedAmount;

    let status: TrainingBudgetCheckStatus = 'within_budget';
    if (!employeeEnough && !departmentEnough) {
      status = 'department_and_employee_exceeded';
    } else if (!employeeEnough) {
      status = 'employee_quota_exceeded';
    } else if (!departmentEnough) {
      status = 'department_budget_exceeded';
    }

    const warnings: string[] = [];
    if (!employeeEnough) {
      warnings.push('Персональной квоты сотрудника недостаточно для этой заявки.');
    }
    if (!departmentEnough) {
      warnings.push('В бюджете отдела недостаточно средств на это обучение.');
    }
    if (status !== 'within_budget') {
      warnings.push('Без дополнительного согласования заявку нельзя отправить дальше.');
    }
    if (currency !== DEFAULT_CURRENCY) {
      warnings.push('Мок работает в RUB. Для реального API позже понадобится конвертация валют.');
    }

    return {
      departmentId: actor.departmentId,
      employeeId: actor.employeeId,
      requestAmount: sanitizedAmount,
      currency,
      employeeRemainingAmount,
      departmentRemainingAmount,
      employeeEnough,
      departmentEnough,
      requiresAdditionalApproval: status !== 'within_budget',
      status,
      warnings,
    };
  }

  private getLatestExpense(requestId: string): TrainingExpense | null {
    const expenses = this.state()
      .expenses.filter((expense) => expense.requestId === requestId)
      .sort((left, right) => right.updatedAt.localeCompare(left.updatedAt));

    return expenses[0] ?? null;
  }

  private findDepartment(departmentId: string, departmentName: string): DepartmentSeed | undefined {
    const normalizedId = normalizeValue(departmentId);
    const normalizedName = normalizeValue(departmentName);

    return this.state().departments.find(
      (department) =>
        normalizeValue(department.id) === normalizedId || normalizeValue(department.name) === normalizedName,
    );
  }
}

function sumAmounts(expenses: TrainingExpense[]): number {
  return expenses.reduce((total, expense) => total + expense.amount, 0);
}

function calculatePercent(value: number, total: number): number {
  if (total <= 0) {
    return 0;
  }
  return Math.round((value / total) * 100);
}

function getUsageStatus(value: number, total: number): BudgetUsageStatus {
  const usagePercent = calculatePercent(value, total);
  if (usagePercent > 100) {
    return 'over_limit';
  }
  if (usagePercent >= 85) {
    return 'risk';
  }
  if (usagePercent >= 60) {
    return 'attention';
  }
  return 'healthy';
}

function isActiveExpense(status: TrainingExpenseStatus): boolean {
  return ACTIVE_EXPENSE_STATUSES.includes(status);
}

function matchesEmployee(
  employee: EmployeeSeed,
  employeeId: string,
  employeeName: string,
  employeeEmail?: string | null,
): boolean {
  return (
    normalizeValue(employee.id) === normalizeValue(employeeId) ||
    normalizeValue(employee.name) === normalizeValue(employeeName) ||
    (!!employeeEmail && normalizeValue(employee.email) === normalizeValue(employeeEmail))
  );
}

function sanitizeAmount(value: number | string | null | undefined): number {
  const normalized = Number.parseFloat(String(value ?? '0').replace(/\s+/g, '').replace(',', '.'));
  if (!Number.isFinite(normalized) || normalized < 0) {
    return 0;
  }
  return Math.round(normalized);
}

function sanitizeText(value: string | null | undefined, fallback: string): string {
  const normalized = value?.trim();
  return normalized && normalized.length > 0 ? normalized : fallback;
}

function sanitizeId(value: string | null | undefined, prefix: string): string {
  const normalized = value?.trim();
  return normalized && normalized.length > 0 ? normalized : buildId(prefix);
}

function buildId(prefix: string): string {
  return `${prefix}-${Math.random().toString(36).slice(2, 10)}`;
}

function normalizeValue(value: string | null | undefined): string {
  return (value ?? '').trim().toLowerCase();
}
