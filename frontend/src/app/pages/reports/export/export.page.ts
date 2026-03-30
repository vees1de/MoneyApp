import { CommonModule } from '@angular/common';
import { Component, inject } from '@angular/core';
import { RouterLink } from '@angular/router';
import { ReportsAnalyticsFacade } from '@features/reports-analytics';
import type { BudgetLimit } from '@entities/budget-limit';

@Component({
  selector: 'app-page-reports-export',
  standalone: true,
  imports: [CommonModule, RouterLink],
  templateUrl: './export.page.html',
  styleUrl: './export.page.scss',
})
export class ReportsExportPageComponent {
  private readonly facade = inject(ReportsAnalyticsFacade);
  protected readonly routePath = '/reports/export';
  protected readonly entitySample: BudgetLimit[] = [];

  protected loadPage(): void {
    this.facade.load();
  }
}
