import { CommonModule } from '@angular/common';
import { Component, inject } from '@angular/core';
import { RouterLink } from '@angular/router';
import { ReportsAnalyticsFacade } from '@features/reports-analytics';
import type { BudgetLimit } from '@entities/budget-limit';

@Component({
  selector: 'app-page-reports-traceability',
  standalone: true,
  imports: [CommonModule, RouterLink],
  templateUrl: './traceability.page.html',
  styleUrl: './traceability.page.scss',
})
export class ReportsTraceabilityPageComponent {
  private readonly facade = inject(ReportsAnalyticsFacade);
  protected readonly routePath = '/reports/traceability';
  protected readonly entitySample: BudgetLimit[] = [];

  protected loadPage(): void {
    this.facade.load();
  }
}
