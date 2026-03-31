import { CommonModule } from '@angular/common';
import { Component, OnInit, inject, signal } from '@angular/core';
import { RouterLink } from '@angular/router';
import { MatCardModule } from '@angular/material/card';
import { catchError, map } from 'rxjs';

import { ExternalRequestsApiService } from '@core/api/external-requests-api.service';
import type { ExternalRequest, PendingApprovalItem } from '@core/api/contracts';
import { AuthStateService } from '@core/auth/auth-state.service';
import { WidgetShellComponent } from '@app/widgets/widget-shell/widget-shell.component';

interface CourseRequestCard {
  id: string;
  title: string;
  subtitle: string;
}

@Component({
  selector: 'app-course-requests-widget',
  standalone: true,
  imports: [CommonModule, RouterLink, MatCardModule, WidgetShellComponent],
  templateUrl: './course-requests-widget.component.html',
  styleUrl: './course-requests-widget.component.scss',
})
export class CourseRequestsWidgetComponent implements OnInit {
  private readonly api = inject(ExternalRequestsApiService);
  private readonly authState = inject(AuthStateService);

  protected readonly loading = signal(true);
  protected readonly error = signal<string | null>(null);
  protected readonly cards = signal<CourseRequestCard[]>([]);

  ngOnInit(): void {
    if (this.authState.hasRole('employee')) {
      this.loadEmployeeRequests();
      return;
    }

    this.api.listPendingApprovals().subscribe({
      next: (requests) => {
        this.cards.set(this.mapPendingApprovals(requests));
        this.loading.set(false);
      },
      error: () => {
        this.error.set('failed');
        this.loading.set(false);
      },
    });
  }

  private loadEmployeeRequests(): void {
    this.api
      .listByScopeAndStatuses('my', ['manager_approval', 'hr_approval'])
      .pipe(
        catchError(() =>
          this.api
            .listMy()
            .pipe(
              map((items) =>
                (items ?? []).filter(
                  (item) => item.status === 'manager_approval' || item.status === 'hr_approval',
                ),
              ),
            ),
        ),
      )
      .subscribe({
        next: (items) => {
          this.cards.set(this.mapEmployeeRequests(items));
          this.loading.set(false);
        },
        error: () => {
          this.error.set('failed');
          this.loading.set(false);
        },
      });
  }

  private mapPendingApprovals(items: PendingApprovalItem[]): CourseRequestCard[] {
    return (items ?? []).slice(0, 3).map((item) => ({
      id: item.request.id,
      title: item.request.title,
      subtitle: item.current_step.role_code,
    }));
  }

  private mapEmployeeRequests(items: ExternalRequest[]): CourseRequestCard[] {
    return (items ?? []).slice(0, 3).map((item) => ({
      id: item.id,
      title: item.title,
      subtitle: item.status,
    }));
  }
}
