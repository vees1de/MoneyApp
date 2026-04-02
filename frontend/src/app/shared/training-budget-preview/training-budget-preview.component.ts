import { CommonModule } from '@angular/common';
import { Component, computed, input } from '@angular/core';
import { MatIconModule } from '@angular/material/icon';

import type {
  BudgetUsageStatus,
  DepartmentTrainingBudget,
  EmployeeTrainingQuota,
  TrainingBudgetCheckStatus,
  TrainingBudgetValidation,
  TrainingExpense,
  TrainingExpenseStatus,
} from '@entities/budget-limit';

@Component({
  selector: 'app-training-budget-preview',
  standalone: true,
  imports: [CommonModule, MatIconModule],
  templateUrl: './training-budget-preview.component.html',
  styleUrl: './training-budget-preview.component.scss',
})
export class TrainingBudgetPreviewComponent {
  readonly title = input('Budget overview');
  readonly subtitle = input<string | null>(null);
  readonly department = input<DepartmentTrainingBudget | null>(null);
  readonly employee = input<EmployeeTrainingQuota | null>(null);
  readonly validation = input<TrainingBudgetValidation | null>(null);
  readonly latestExpense = input<TrainingExpense | null>(null);
  readonly history = input<TrainingExpense[]>([]);
  readonly compact = input(false);
  readonly showHistory = input(true);
  readonly showLatest = input(true);
  readonly showMockBadge = input(true);

  protected readonly isWarning = computed(() => !!this.validation()?.requiresAdditionalApproval);

  protected readonly limitedHistory = computed(() => this.history().slice(0, this.compact() ? 3 : 5));

  protected width(percent: number): number {
    return Math.max(0, Math.min(100, percent));
  }

  protected usageStatusLabel(status: BudgetUsageStatus): string {
    switch (status) {
      case 'healthy':
        return 'Норма';
      case 'attention':
        return 'Под контролем';
      case 'risk':
        return 'Риск';
      case 'over_limit':
        return 'Превышение';
      default:
        return status;
    }
  }

  protected expenseStatusLabel(status: TrainingExpenseStatus): string {
    switch (status) {
      case 'reserved':
        return 'В резерве';
      case 'approved':
        return 'Согласовано';
      case 'completed':
        return 'Завершено';
      case 'rejected':
        return 'Отклонено';
      case 'revision_returned':
        return 'Возвращено';
      case 'cancelled':
        return 'Отменено';
      default:
        return status;
    }
  }

  protected validationStatusLabel(status: TrainingBudgetCheckStatus): string {
    switch (status) {
      case 'within_budget':
        return 'В пределах бюджета';
      case 'employee_quota_exceeded':
        return 'Превышена квота сотрудника';
      case 'department_budget_exceeded':
        return 'Превышен бюджет отдела';
      case 'department_and_employee_exceeded':
        return 'Превышены квота и бюджет';
      default:
        return status;
    }
  }
}
