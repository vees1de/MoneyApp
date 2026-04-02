import { CommonModule, DatePipe } from '@angular/common';
import { Component, computed, inject, input, signal } from '@angular/core';

import { TrainingBudgetMockService } from '@core/domain/training-budget-mock.service';
import { WidgetShellComponent } from '@app/widgets/widget-shell/widget-shell.component';
import type { TrainingExpense, TrainingExpenseStatus } from '@entities/budget-limit';

@Component({
  selector: 'app-spending-history-widget',
  standalone: true,
  imports: [CommonModule, DatePipe, WidgetShellComponent],
  templateUrl: './spending-history-widget.component.html',
  styleUrl: './spending-history-widget.component.scss',
})
export class SpendingHistoryWidgetComponent {
  private readonly budgetMock = inject(TrainingBudgetMockService);

  readonly departmentId = input<string | null>(null);
  readonly limit = input(8);

  protected readonly loading = signal(false);
  protected readonly error = signal<string | null>(null);
  protected readonly expenses = this.budgetMock.history;

  protected readonly displayedExpenses = computed(() => {
    const all = this.expenses();
    const departmentId = this.departmentId();
    const filtered = departmentId ? all.filter((expense) => expense.departmentId === departmentId) : all;
    return filtered.slice(0, this.limit());
  });

  protected readonly totalApproved = computed(() =>
    this.expenses()
      .filter((expense) => expense.status === 'approved' || expense.status === 'completed')
      .reduce((sum, expense) => sum + expense.amount, 0),
  );

  protected statusLabel(status: TrainingExpenseStatus): string {
    switch (status) {
      case 'reserved':
        return 'Зарезервировано';
      case 'approved':
        return 'Одобрено';
      case 'completed':
        return 'Завершено';
      case 'rejected':
        return 'Отклонено';
      case 'revision_returned':
        return 'На доработке';
      case 'cancelled':
        return 'Отменено';
      default:
        return status;
    }
  }

  protected statusClass(status: TrainingExpenseStatus): string {
    switch (status) {
      case 'reserved':
        return 'expense-status--reserved';
      case 'approved':
        return 'expense-status--approved';
      case 'completed':
        return 'expense-status--completed';
      case 'rejected':
        return 'expense-status--rejected';
      case 'revision_returned':
        return 'expense-status--revision';
      case 'cancelled':
        return 'expense-status--cancelled';
      default:
        return 'expense-status--reserved';
    }
  }

  protected formatCurrency(value: number): string {
    return new Intl.NumberFormat('ru-RU', {
      style: 'currency',
      currency: 'RUB',
      maximumFractionDigits: 0,
    }).format(value);
  }
}
