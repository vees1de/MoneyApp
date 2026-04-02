import { CommonModule } from '@angular/common';
import { Component, computed, inject, signal } from '@angular/core';
import { MatProgressBarModule } from '@angular/material/progress-bar';

import { TrainingBudgetMockService } from '@core/domain/training-budget-mock.service';
import { WidgetShellComponent } from '@app/widgets/widget-shell/widget-shell.component';
import type { BudgetUsageStatus, DepartmentTrainingBudget } from '@entities/budget-limit';

@Component({
  selector: 'app-department-budget-widget',
  standalone: true,
  imports: [CommonModule, MatProgressBarModule, WidgetShellComponent],
  templateUrl: './department-budget-widget.component.html',
  styleUrl: './department-budget-widget.component.scss',
})
export class DepartmentBudgetWidgetComponent {
  private readonly budgetMock = inject(TrainingBudgetMockService);

  protected readonly loading = signal(false);
  protected readonly error = signal<string | null>(null);
  protected readonly budgets = this.budgetMock.departments;

  protected readonly totalBudget = computed(() =>
    this.budgets().reduce((sum, budget) => sum + budget.totalBudget, 0),
  );
  protected readonly totalUsed = computed(() =>
    this.budgets().reduce((sum, budget) => sum + budget.usedAmount, 0),
  );
  protected readonly totalRemaining = computed(() =>
    this.budgets().reduce((sum, budget) => sum + budget.remainingAmount, 0),
  );
  protected readonly overallPercent = computed(() => {
    const total = this.totalBudget();
    return total > 0 ? Math.round((this.totalUsed() / total) * 100) : 0;
  });

  protected barColor(status: BudgetUsageStatus): 'primary' | 'accent' | 'warn' {
    if (status === 'over_limit' || status === 'risk') return 'warn';
    if (status === 'attention') return 'accent';
    return 'primary';
  }

  protected overallBarColor(): 'primary' | 'accent' | 'warn' {
    const percent = this.overallPercent();
    if (percent >= 85) return 'warn';
    if (percent >= 60) return 'accent';
    return 'primary';
  }

  protected formatCurrency(value: number): string {
    return new Intl.NumberFormat('ru-RU', {
      style: 'currency',
      currency: 'RUB',
      maximumFractionDigits: 0,
    }).format(value);
  }

  protected usageLabel(budget: DepartmentTrainingBudget): string {
    return `${this.formatCurrency(budget.usedAmount)} из ${this.formatCurrency(budget.totalBudget)}`;
  }
}
