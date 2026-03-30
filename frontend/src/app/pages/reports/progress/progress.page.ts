import { CommonModule } from '@angular/common';
import { Component, inject } from '@angular/core';
import { RouterLink } from '@angular/router';
import { ReportsAnalyticsFacade } from '@features/reports-analytics';
import type { BudgetLimit } from '@entities/budget-limit';

@Component({
  selector: 'app-page-reports-progress',
  standalone: true,
  imports: [CommonModule, RouterLink],
  templateUrl: './progress.page.html',
  styleUrl: './progress.page.scss',
})
export class ReportsProgressPageComponent {
  private readonly facade = inject(ReportsAnalyticsFacade);
  protected readonly routePath = '/reports/progress';
  protected readonly entitySample: BudgetLimit[] = [];

  protected loadPage(): void {
    this.facade.load();
  }
}
