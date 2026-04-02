import { CommonModule } from '@angular/common';
import { Component, computed, inject } from '@angular/core';
import { RouterLink } from '@angular/router';
import { MatButtonModule } from '@angular/material/button';
import { MatCardModule } from '@angular/material/card';
import { MatIconModule } from '@angular/material/icon';

import { TrainingBudgetMockService } from '@core/domain/training-budget-mock.service';
import type { BudgetUsageStatus, TrainingExpenseStatus } from '@entities/budget-limit';

@Component({
  selector: 'app-page-reports-budget',
  standalone: true,
  imports: [CommonModule, RouterLink, MatButtonModule, MatCardModule, MatIconModule],
  templateUrl: './budget.page.html',
  styleUrl: './budget.page.scss',
})
export class ReportsBudgetPageComponent {
  private readonly budgetMock = inject(TrainingBudgetMockService);

  protected readonly snapshot = this.budgetMock.snapshot;
  protected readonly departmentRows = this.budgetMock.departmentReport;
  protected readonly employeeRows = this.budgetMock.employeeReport;
  protected readonly topCourses = this.budgetMock.topCourses;
  protected readonly history = computed(() => this.budgetMock.history().slice(0, 12));

  protected readonly totals = computed(() => {
    const departments = this.snapshot().departments;
    const totalBudget = departments.reduce((sum, item) => sum + item.totalBudget, 0);
    const spentAmount = departments.reduce((sum, item) => sum + item.spentAmount, 0);
    const reservedAmount = departments.reduce((sum, item) => sum + item.reservedAmount, 0);
    const usedAmount = spentAmount + reservedAmount;
    const remainingAmount = totalBudget - usedAmount;
    const activeRequests = departments.reduce(
      (sum, item) => sum + item.expenses.filter((expense) => ['reserved', 'approved', 'completed'].includes(expense.status)).length,
      0,
    );

    return {
      totalBudget,
      spentAmount,
      reservedAmount,
      usedAmount,
      remainingAmount,
      activeRequests,
    };
  });

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

  protected tone(status: BudgetUsageStatus): string {
    return status;
  }

  protected progress(value: number): number {
    return Math.max(0, Math.min(100, value));
  }
}
