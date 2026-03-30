import { CommonModule } from '@angular/common';
import { Component, inject } from '@angular/core';
import { RouterLink } from '@angular/router';
import { ReportsAnalyticsFacade } from '@features/reports-analytics';
import type { BudgetLimit } from '@entities/budget-limit';

@Component({
  selector: 'app-page-reports-overview',
  standalone: true,
  imports: [CommonModule, RouterLink],
  templateUrl: './overview.page.html',
  styleUrl: './overview.page.scss',
})
export class ReportsOverviewPageComponent {
  private readonly facade = inject(ReportsAnalyticsFacade);
  protected readonly routePath = '/reports/overview';
  protected readonly entitySample: BudgetLimit[] = [];

  protected loadPage(): void {
    this.facade.load();
  }
}
