import { CommonModule } from '@angular/common';
import { Component, OnInit, computed, inject, signal } from '@angular/core';
import { RouterLink } from '@angular/router';
import { MatButtonModule } from '@angular/material/button';
import { MatCardModule } from '@angular/material/card';
import { catchError, map } from 'rxjs';

import { ExternalRequestsApiService } from '@core/api/external-requests-api.service';
import type { ExternalRequest, PendingApprovalItem } from '@core/api/contracts';
import { AuthStateService } from '@core/auth/auth-state.service';
import { WidgetShellComponent } from '@app/widgets/widget-shell/widget-shell.component';

interface CourseRequestCard {
  id: string;
  title: string;
  employee: string;
  status: string;
  subtitle: string;
  amount: string;
}

@Component({
  selector: 'app-course-requests-widget',
  standalone: true,
  imports: [CommonModule, RouterLink, MatCardModule, MatButtonModule, WidgetShellComponent],
  templateUrl: './course-requests-widget.component.html',
  styleUrl: './course-requests-widget.component.scss',
})
export class CourseRequestsWidgetComponent implements OnInit {
  private readonly api = inject(ExternalRequestsApiService);
  private readonly authState = inject(AuthStateService);

  protected readonly loading = signal(true);
  protected readonly error = signal<string | null>(null);
  protected readonly cards = signal<CourseRequestCard[]>([]);
  protected readonly selectedStatus = signal<string>('all');
  protected readonly statuses = [
    { key: 'all', label: 'Все' },
    { key: 'manager_approval', label: 'На согласовании у руководителя' },
    { key: 'hr_approval', label: 'На согласовании у HR' },
    { key: 'approved', label: 'Одобрено' },
    { key: 'rejected', label: 'Отклонено' },
  ];
  protected readonly filteredCards = computed(() => {
    const status = this.selectedStatus();
    const source = this.cards();
    return status === 'all' ? source : source.filter((item) => item.status === status);
  });

  ngOnInit(): void {
    this.loadByRoleScope();
  }

  protected setStatus(status: string): void {
    this.selectedStatus.set(status);
  }

  protected statusCount(status: string): number {
    if (status === 'all') {
      return this.cards().length;
    }
    return this.cards().filter((item) => item.status === status).length;
  }

  private loadByRoleScope(): void {
    const scope = this.authState.hasRole('employee') ? 'my' : 'all';

    this.api
      .listFiltered({ scope })
      .pipe(
        catchError(() => {
          if (scope === 'my') {
            return this.api.listByScopeAndStatuses('my', ['manager_approval', 'hr_approval']);
          }
          return this.api.listPendingApprovals().pipe(map((items) => this.mapPending(items)));
        }),
      )
      .subscribe({
        next: (items) => {
          this.cards.set(this.mapRequests(items));
          this.loading.set(false);
        },
        error: () => {
          this.error.set('failed');
          this.loading.set(false);
        },
      });
  }

  private mapPending(items: PendingApprovalItem[]): ExternalRequest[] {
    return (items ?? []).map((item) => item.request);
  }

  private mapRequests(items: ExternalRequest[]): CourseRequestCard[] {
    return (items ?? []).slice(0, 10).map((item) => ({
      id: item.id,
      title: item.title,
      employee: item.employee_full_name || item.employee_email || 'Сотрудник не указан',
      status: item.status,
      subtitle: item.provider_name || item.current_approval_role_code || item.status,
      amount: `${item.cost_amount || '0'} ${item.currency || 'RUB'}`,
    }));
  }
}
