import { Injectable } from '@angular/core';
import { Observable, of, delay } from 'rxjs';

import {
  MOCK_DEPARTMENT_BUDGETS,
  MOCK_ALL_EMPLOYEES,
  MOCK_EXPENSES,
  MOCK_DEPARTMENT_REPORTS,
  MOCK_TOP_COURSES,
  MOCK_BUDGET_LIMIT,
} from '@entities/budget-limit';
import type {
  DepartmentTrainingBudget,
  EmployeeTrainingQuota,
  TrainingExpense,
  TrainingBudgetValidation,
  DepartmentBudgetReportRow,
  CourseSpendReportRow,
  BudgetLimit,
} from '@entities/budget-limit';

const MOCK_DELAY = 400;

@Injectable({ providedIn: 'root' })
export class BudgetMockApiService {
  getBudgetOverview(): Observable<BudgetLimit> {
    return of({ ...MOCK_BUDGET_LIMIT }).pipe(delay(MOCK_DELAY));
  }

  getDepartmentBudgets(): Observable<DepartmentTrainingBudget[]> {
    return of([...MOCK_DEPARTMENT_BUDGETS]).pipe(delay(MOCK_DELAY));
  }

  getDepartmentBudgetById(departmentId: string): Observable<DepartmentTrainingBudget | null> {
    const found = MOCK_DEPARTMENT_BUDGETS.find((d) => d.departmentId === departmentId) ?? null;
    return of(found ? { ...found } : null).pipe(delay(MOCK_DELAY));
  }

  getEmployeeQuotas(departmentId?: string): Observable<EmployeeTrainingQuota[]> {
    let items = [...MOCK_ALL_EMPLOYEES];
    if (departmentId) {
      items = items.filter((q) => q.departmentId === departmentId);
    }
    return of(items).pipe(delay(MOCK_DELAY));
  }

  getEmployeeQuotaByUserId(userId: string): Observable<EmployeeTrainingQuota | null> {
    const found = MOCK_ALL_EMPLOYEES.find((q) => q.employeeId === userId) ?? null;
    return of(found ? { ...found } : null).pipe(delay(MOCK_DELAY));
  }

  getExpenseHistory(departmentId?: string): Observable<TrainingExpense[]> {
    let items = [...MOCK_EXPENSES];
    if (departmentId) {
      items = items.filter((e) => e.departmentId === departmentId);
    }
    return of(items).pipe(delay(MOCK_DELAY));
  }

  validateBudget(employeeId: string, departmentId: string, amount: number): Observable<TrainingBudgetValidation> {
    const employee = MOCK_ALL_EMPLOYEES.find((q) => q.employeeId === employeeId);
    const department = MOCK_DEPARTMENT_BUDGETS.find((d) => d.departmentId === departmentId);

    const empRemaining = employee?.remainingAmount ?? 0;
    const deptRemaining = department?.remainingAmount ?? 0;
    const empEnough = empRemaining >= amount;
    const deptEnough = deptRemaining >= amount;

    const warnings: string[] = [];
    let status: TrainingBudgetValidation['status'] = 'within_budget';

    if (!empEnough && !deptEnough) {
      status = 'department_and_employee_exceeded';
      warnings.push(`Персональная квота сотрудника превышена на ${formatAmount(amount - empRemaining)}`);
      warnings.push(`Бюджет отдела превышен на ${formatAmount(amount - deptRemaining)}`);
    } else if (!empEnough) {
      status = 'employee_quota_exceeded';
      warnings.push(`Персональная квота сотрудника превышена на ${formatAmount(amount - empRemaining)}`);
    } else if (!deptEnough) {
      status = 'department_budget_exceeded';
      warnings.push(`Бюджет отдела превышен на ${formatAmount(amount - deptRemaining)}`);
    }

    const result: TrainingBudgetValidation = {
      departmentId: departmentId,
      employeeId: employeeId,
      requestAmount: amount,
      currency: 'RUB',
      employeeRemainingAmount: empRemaining,
      departmentRemainingAmount: deptRemaining,
      employeeEnough: empEnough,
      departmentEnough: deptEnough,
      requiresAdditionalApproval: !empEnough || !deptEnough,
      status,
      warnings,
    };

    return of(result).pipe(delay(MOCK_DELAY));
  }

  getDepartmentReports(): Observable<DepartmentBudgetReportRow[]> {
    return of([...MOCK_DEPARTMENT_REPORTS]).pipe(delay(MOCK_DELAY));
  }

  getTopExpensiveCourses(): Observable<CourseSpendReportRow[]> {
    return of([...MOCK_TOP_COURSES]).pipe(delay(MOCK_DELAY));
  }
}

function formatAmount(value: number): string {
  return new Intl.NumberFormat('ru-RU', { style: 'currency', currency: 'RUB', maximumFractionDigits: 0 }).format(value);
}
