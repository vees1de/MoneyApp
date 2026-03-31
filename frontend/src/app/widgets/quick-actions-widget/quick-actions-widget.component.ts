import { CommonModule } from '@angular/common';
import { Component, OnInit, inject, signal } from '@angular/core';
import { RouterLink } from '@angular/router';
import { MatButtonModule } from '@angular/material/button';
import { forkJoin } from 'rxjs';

import { ExternalRequestsApiService } from '@core/api/external-requests-api.service';
import { LearningPlanApiService } from '@core/api/learning-plan-api.service';
import { WidgetShellComponent } from '@app/widgets/widget-shell/widget-shell.component';

@Component({
  selector: 'app-quick-actions-widget',
  standalone: true,
  imports: [CommonModule, RouterLink, MatButtonModule, WidgetShellComponent],
  templateUrl: './quick-actions-widget.component.html',
  styleUrl: './quick-actions-widget.component.scss',
})
export class QuickActionsWidgetComponent implements OnInit {
  private readonly learningPlanApi = inject(LearningPlanApiService);
  private readonly externalRequestsApi = inject(ExternalRequestsApiService);

  protected readonly loading = signal(true);
  protected readonly error = signal<string | null>(null);
  protected readonly activeLearning = signal(0);
  protected readonly myRequests = signal(0);

  ngOnInit(): void {
    forkJoin({
      plan: this.learningPlanApi.getMyPlan(),
      requests: this.externalRequestsApi.listMy(),
    }).subscribe({
      next: ({ plan, requests }) => {
        this.activeLearning.set(plan?.summary?.in_progress ?? 0);
        this.myRequests.set((requests ?? []).length);
        this.loading.set(false);
      },
      error: () => {
        this.error.set('failed');
        this.loading.set(false);
      },
    });
  }
}
