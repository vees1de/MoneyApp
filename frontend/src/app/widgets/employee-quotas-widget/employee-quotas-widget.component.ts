import { CommonModule } from '@angular/common';
import { Component, OnInit, inject, signal, input, computed } from '@angular/core';
import { MatProgressBarModule } from '@angular/material/progress-bar';
import { RouterLink } from '@angular/router';

import { BudgetMockApiService } from '@core/api/budget-mock-api.service';
import { WidgetShellComponent } from '@app/widgets/widget-shell/widget-shell.component';
import type { EmployeeTrainingQuota, BudgetUsageStatus } from '@entities/budget-limit';

@Component({
  selector: 'app-employee-quotas-widget',
  standalone: true,
  imports: [CommonModule, MatProgressBarModule, RouterLink, WidgetShellComponent],
  templateUrl: './employee-quotas-widget.component.html',
  styleUrl: './employee-quotas-widget.component.scss',
})
export class EmployeeQuotasWidgetComponent implements OnInit {
  private readonly api = inject(BudgetMockApiService);

  readonly departmentId = input<string | null>(null);
  readonly limit = input(5);

  protected readonly loading = signal(true);
  protected readonly error = signal<string | null>(null);
  protected readonly quotas = signal<EmployeeTrainingQuota[]>([]);

  protected readonly displayedQuotas = computed(() => {
    const all = this.quotas();
    const depId = this.departmentId();
    const filtered = depId ? all.filter((q) => q.departmentId === depId) : all;
    return filtered.slice(0, this.limit());
  });

  ngOnInit(): void {
    this.api.getEmployeeQuotas(this.departmentId() ?? undefined).subscribe({
      next: (data) => {
        this.quotas.set(data);
        this.loading.set(false);
      },
      error: () => {
        this.error.set('failed');
        this.loading.set(false);
      },
    });
  }

  protected barColor(status: BudgetUsageStatus): 'primary' | 'accent' | 'warn' {
    if (status === 'over_limit' || status === 'risk') return 'warn';
    if (status === 'attention') return 'accent';
    return 'primary';
  }

  protected statusLabel(status: BudgetUsageStatus): string {
    switch (status) {
      case 'over_limit': return 'Исчерпана';
      case 'risk': return 'Критично';
      case 'attention': return 'Внимание';
      case 'healthy': return 'Доступна';
    }
  }

  protected statusClass(status: BudgetUsageStatus): string {
    switch (status) {
      case 'over_limit': return 'status--exhausted';
      case 'risk': return 'status--risk';
      case 'attention': return 'status--low';
      case 'healthy': return 'status--ok';
    }
  }

  protected formatCurrency(value: number): string {
    return new Intl.NumberFormat('ru-RU', { style: 'currency', currency: 'RUB', maximumFractionDigits: 0 }).format(value);
  }
}
