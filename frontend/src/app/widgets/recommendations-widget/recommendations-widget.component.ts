import { CommonModule } from '@angular/common';
import { Component, OnInit, inject, signal } from '@angular/core';
import { RouterLink } from '@angular/router';
import { MatCardModule } from '@angular/material/card';

import { RecommendationsApiService } from '@core/api/recommendations-api.service';
import type { RecommendedCourseItem } from '@core/api/contracts';
import { WidgetShellComponent } from '@app/widgets/widget-shell/widget-shell.component';

@Component({
  selector: 'app-recommendations-widget',
  standalone: true,
  imports: [CommonModule, RouterLink, MatCardModule, WidgetShellComponent],
  templateUrl: './recommendations-widget.component.html',
  styleUrl: './recommendations-widget.component.scss',
})
export class RecommendationsWidgetComponent implements OnInit {
  private readonly api = inject(RecommendationsApiService);

  protected readonly loading = signal(true);
  protected readonly error = signal<string | null>(null);
  protected readonly courses = signal<RecommendedCourseItem[]>([]);

  ngOnInit(): void {
    this.api.list({ limit: 3, offset: 0 }).subscribe({
      next: (courses) => {
        this.courses.set(courses ?? []);
        this.loading.set(false);
      },
      error: () => {
        this.error.set('failed');
        this.loading.set(false);
      },
    });
  }
}
