import { CommonModule } from '@angular/common';
import { Component, OnInit, inject, signal } from '@angular/core';
import { MatGridListModule } from '@angular/material/grid-list';

import { IntegrationsApiService } from '@core/api/integrations-api.service';
import { WidgetShellComponent } from '@app/widgets/widget-shell/widget-shell.component';

@Component({
  selector: 'app-jira-summary-widget',
  standalone: true,
  imports: [CommonModule, MatGridListModule, WidgetShellComponent],
  templateUrl: './jira-summary-widget.component.html',
  styleUrl: './jira-summary-widget.component.scss',
})
export class JiraSummaryWidgetComponent implements OnInit {
  private readonly api = inject(IntegrationsApiService);

  protected readonly loading = signal(true);
  protected readonly error = signal<string | null>(null);
  protected readonly totals = signal({
    tasks: 0,
    active: 0,
    completed: 0,
    overdue: 0,
  });

  ngOnInit(): void {
    this.api.getJiraBoardSummary().subscribe({
      next: (summary) => {
        this.totals.set({
          tasks: summary.summary?.tasks_total ?? 0,
          active: summary.summary?.active_total ?? 0,
          completed: summary.summary?.completed_total ?? 0,
          overdue: summary.summary?.overdue_total ?? 0,
        });
        this.loading.set(false);
      },
      error: () => {
        this.error.set('failed');
        this.loading.set(false);
      },
    });
  }
}
