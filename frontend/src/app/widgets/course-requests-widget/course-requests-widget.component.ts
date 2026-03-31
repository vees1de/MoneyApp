import { CommonModule } from '@angular/common';
import { Component, OnInit, inject, signal } from '@angular/core';
import { RouterLink } from '@angular/router';
import { MatCardModule } from '@angular/material/card';

import { ExternalRequestsApiService } from '@core/api/external-requests-api.service';
import type { PendingApprovalItem } from '@core/api/contracts';
import { WidgetShellComponent } from '@app/widgets/widget-shell/widget-shell.component';

@Component({
  selector: 'app-course-requests-widget',
  standalone: true,
  imports: [CommonModule, RouterLink, MatCardModule, WidgetShellComponent],
  templateUrl: './course-requests-widget.component.html',
  styleUrl: './course-requests-widget.component.scss',
})
export class CourseRequestsWidgetComponent implements OnInit {
  private readonly api = inject(ExternalRequestsApiService);

  protected readonly loading = signal(true);
  protected readonly error = signal<string | null>(null);
  protected readonly requests = signal<PendingApprovalItem[]>([]);

  ngOnInit(): void {
    this.api.listPendingApprovals().subscribe({
      next: (requests) => {
        this.requests.set((requests ?? []).slice(0, 3));
        this.loading.set(false);
      },
      error: () => {
        this.error.set('failed');
        this.loading.set(false);
      },
    });
  }
}
