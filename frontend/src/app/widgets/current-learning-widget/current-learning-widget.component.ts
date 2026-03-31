import { CommonModule } from '@angular/common';
import { Component, OnInit, inject, signal } from '@angular/core';
import { RouterLink } from '@angular/router';
import { MatListModule } from '@angular/material/list';
import { MatProgressBarModule } from '@angular/material/progress-bar';

import { LearningPlanApiService } from '@core/api/learning-plan-api.service';
import type { LearningPlanItem } from '@core/api/contracts';
import { WidgetShellComponent } from '@app/widgets/widget-shell/widget-shell.component';

@Component({
  selector: 'app-current-learning-widget',
  standalone: true,
  imports: [CommonModule, RouterLink, MatListModule, MatProgressBarModule, WidgetShellComponent],
  templateUrl: './current-learning-widget.component.html',
  styleUrl: './current-learning-widget.component.scss',
})
export class CurrentLearningWidgetComponent implements OnInit {
  private readonly api = inject(LearningPlanApiService);

  protected readonly loading = signal(true);
  protected readonly error = signal<string | null>(null);
  protected readonly items = signal<LearningPlanItem[]>([]);

  ngOnInit(): void {
    this.api.getMyPlan().subscribe({
      next: (payload) => {
        this.items.set((payload?.in_progress ?? []).slice(0, 3));
        this.loading.set(false);
      },
      error: () => {
        this.error.set('failed');
        this.loading.set(false);
      },
    });
  }

  protected progressValue(item: LearningPlanItem): number {
    const value = Number(item.completion_percent);
    if (Number.isNaN(value)) {
      return 0;
    }
    return Math.max(0, Math.min(100, value));
  }
}
