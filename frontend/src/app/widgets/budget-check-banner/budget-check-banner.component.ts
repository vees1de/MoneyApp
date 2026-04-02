import { CommonModule } from '@angular/common';
import { Component, OnInit, inject, signal, input, output, computed } from '@angular/core';
import { MatIconModule } from '@angular/material/icon';
import { MatProgressBarModule } from '@angular/material/progress-bar';

import { BudgetMockApiService } from '@core/api/budget-mock-api.service';
import type { TrainingBudgetValidation } from '@entities/budget-limit';

@Component({
  selector: 'app-budget-check-banner',
  standalone: true,
  imports: [CommonModule, MatIconModule, MatProgressBarModule],
  templateUrl: './budget-check-banner.component.html',
  styleUrl: './budget-check-banner.component.scss',
})
export class BudgetCheckBannerComponent implements OnInit {
  private readonly api = inject(BudgetMockApiService);

  readonly courseCost = input.required<number>();
  readonly userId = input<string>('user-1');
  readonly departmentId = input<string>('dept-1');

  readonly budgetBlocked = output<boolean>();

  protected readonly loading = signal(true);
  protected readonly validation = signal<TrainingBudgetValidation | null>(null);

  protected readonly isBlocked = computed(() => {
    const v = this.validation();
    return v ? v.requiresAdditionalApproval : false;
  });

  protected readonly quotaPercent = computed(() => {
    const v = this.validation();
    if (!v) return 0;
    const total = v.employeeRemainingAmount + v.requestAmount;
    return total > 0 ? Math.round(((total - v.employeeRemainingAmount) / total) * 100) : 0;
  });

  ngOnInit(): void {
    this.api.validateBudget(this.userId(), this.departmentId(), this.courseCost()).subscribe({
      next: (result) => {
        this.validation.set(result);
        this.loading.set(false);
        this.budgetBlocked.emit(result.requiresAdditionalApproval);
      },
      error: () => {
        this.loading.set(false);
      },
    });
  }

  protected formatCurrency(value: number): string {
    return new Intl.NumberFormat('ru-RU', { style: 'currency', currency: 'RUB', maximumFractionDigits: 0 }).format(value);
  }
}
