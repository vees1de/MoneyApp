import { CommonModule } from '@angular/common';
import { Component, inject } from '@angular/core';
import { RouterLink } from '@angular/router';
import { ReportsAnalyticsFacade } from '@features/reports-analytics';
import type { BudgetLimit } from '@entities/budget-limit';

@Component({
  selector: 'app-page-reports-budget',
  standalone: true,
  imports: [CommonModule, RouterLink],
  templateUrl: './budget.page.html',
  styleUrl: './budget.page.scss',
})
export class ReportsBudgetPageComponent {
  private readonly facade = inject(ReportsAnalyticsFacade);
  protected readonly routePath = '/reports/budget';
  protected readonly entitySample: BudgetLimit[] = [];

  protected loadPage(): void {
    this.facade.load();
  }
}
