import { Injectable, inject, signal, computed } from '@angular/core';

import { BudgetMockApiService } from '@core/api/budget-mock-api.service';
import type {
  DepartmentTrainingBudget,
  EmployeeTrainingQuota,
  TrainingExpense,
  TrainingBudgetValidation,
  DepartmentBudgetReportRow,
  CourseSpendReportRow,
} from '@entities/budget-limit';

@Injectable({ providedIn: 'root' })
export class BudgetFacade {
  private readonly api = inject(BudgetMockApiService);

  readonly departments = signal<DepartmentTrainingBudget[]>([]);
  readonly employees = signal<EmployeeTrainingQuota[]>([]);
  readonly expenses = signal<TrainingExpense[]>([]);
  readonly departmentReports = signal<DepartmentBudgetReportRow[]>([]);
  readonly topCourses = signal<CourseSpendReportRow[]>([]);
  readonly validation = signal<TrainingBudgetValidation | null>(null);

  readonly loading = signal(false);
  readonly error = signal<string | null>(null);
  readonly validating = signal(false);

  readonly selectedDepartmentId = signal<string | null>(null);

  readonly totalBudget = computed(() =>
    this.departments().reduce((sum, d) => sum + d.totalBudget, 0),
  );
  readonly totalSpent = computed(() =>
    this.departments().reduce((sum, d) => sum + d.spentAmount, 0),
  );
  readonly totalUsed = computed(() =>
    this.departments().reduce((sum, d) => sum + d.usedAmount, 0),
  );
  readonly totalRemaining = computed(() =>
    this.departments().reduce((sum, d) => sum + d.remainingAmount, 0),
  );
  readonly overallUtilizationPercent = computed(() => {
    const total = this.totalBudget();
    return total > 0 ? Math.round((this.totalUsed() / total) * 100) : 0;
  });

  readonly selectedDepartment = computed(() => {
    const id = this.selectedDepartmentId();
    if (!id) return null;
    return this.departments().find((d) => d.departmentId === id) ?? null;
  });

  readonly filteredEmployees = computed(() => {
    const id = this.selectedDepartmentId();
    if (!id) return this.employees();
    return this.employees().filter((e) => e.departmentId === id);
  });

  readonly filteredExpenses = computed(() => {
    const id = this.selectedDepartmentId();
    if (!id) return this.expenses();
    return this.expenses().filter((e) => e.departmentId === id);
  });

  loadAll(): void {
    this.loading.set(true);
    this.error.set(null);

    let completed = 0;
    const total = 5;
    const checkDone = () => {
      completed++;
      if (completed >= total) this.loading.set(false);
    };

    this.api.getDepartmentBudgets().subscribe({
      next: (data) => { this.departments.set(data); checkDone(); },
      error: () => { this.error.set('Ошибка загрузки бюджетов'); checkDone(); },
    });

    this.api.getEmployeeQuotas().subscribe({
      next: (data) => { this.employees.set(data); checkDone(); },
      error: () => { this.error.set('Ошибка загрузки квот'); checkDone(); },
    });

    this.api.getExpenseHistory().subscribe({
      next: (data) => { this.expenses.set(data); checkDone(); },
      error: () => { this.error.set('Ошибка загрузки истории'); checkDone(); },
    });

    this.api.getDepartmentReports().subscribe({
      next: (data) => { this.departmentReports.set(data); checkDone(); },
      error: () => { this.error.set('Ошибка загрузки отчётов'); checkDone(); },
    });

    this.api.getTopExpensiveCourses().subscribe({
      next: (data) => { this.topCourses.set(data); checkDone(); },
      error: () => { this.error.set('Ошибка загрузки курсов'); checkDone(); },
    });
  }

  selectDepartment(departmentId: string | null): void {
    this.selectedDepartmentId.set(departmentId);
  }

  validateBudgetForRequest(employeeId: string, departmentId: string, amount: number): void {
    this.validating.set(true);
    this.validation.set(null);

    this.api.validateBudget(employeeId, departmentId, amount).subscribe({
      next: (result) => {
        this.validation.set(result);
        this.validating.set(false);
      },
      error: () => {
        this.validating.set(false);
      },
    });
  }

  clearValidation(): void {
    this.validation.set(null);
  }
}
