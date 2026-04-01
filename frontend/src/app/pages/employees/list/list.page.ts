import { CommonModule } from '@angular/common';
import { Component, OnDestroy, OnInit, inject } from '@angular/core';
import { MatButtonModule } from '@angular/material/button';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatIconModule } from '@angular/material/icon';
import { MatInputModule } from '@angular/material/input';
import { MatTabsModule } from '@angular/material/tabs';
import { MatSortModule, type Sort } from '@angular/material/sort';

import { UsersApiService } from '@core/api/users-api.service';
import type {
  EmployeeLearningStats,
  EmployeesLearningStatsQuery,
  EmployeesScope,
} from '@core/api/contracts';
import { AuthStateService } from '@core/auth/auth-state.service';
import { UpcomingEventsWidgetComponent } from '@app/widgets/upcoming-events-widget/upcoming-events-widget.component';

@Component({
  selector: 'app-page-employees-list',
  standalone: true,
  imports: [
    CommonModule,
    MatButtonModule,
    MatFormFieldModule,
    MatIconModule,
    MatInputModule,
    MatSortModule,
    MatTabsModule,
    UpcomingEventsWidgetComponent,
  ],
  templateUrl: './list.page.html',
  styleUrl: './list.page.scss',
})
export class EmployeesListPageComponent implements OnInit, OnDestroy {
  private readonly usersApi = inject(UsersApiService);
  private readonly authState = inject(AuthStateService);
  private readonly pageLimit = 200;
  private searchDebounceTimer: ReturnType<typeof setTimeout> | null = null;
  private requestSerial = 0;

  protected loading = true;
  protected error: string | null = null;
  protected searchTerm = '';
  protected allRows: EmployeeLearningStats[] = [];
  protected sortedData: EmployeeLearningStats[] = [];

  protected readonly tabCounts = {
    all: 0,
    external: 0,
    university: 0,
    mandatory: 0,
  };

  ngOnInit(): void {
    this.loadEmployees();
  }

  ngOnDestroy(): void {
    if (this.searchDebounceTimer) {
      clearTimeout(this.searchDebounceTimer);
    }
  }

  protected onSearchChange(event: Event): void {
    const target = event.target as HTMLInputElement | null;
    this.searchTerm = target?.value?.trim() ?? '';

    if (this.searchDebounceTimer) {
      clearTimeout(this.searchDebounceTimer);
    }
    this.searchDebounceTimer = setTimeout(() => this.loadEmployees(), 300);
  }

  protected sortData(sort: Sort): void {
    const data = [...this.allRows];
    if (!sort.active || sort.direction === '') {
      this.sortedData = data;
      return;
    }

    this.sortedData = data.sort((a, b) => {
      const isAsc = sort.direction === 'asc';
      switch (sort.active) {
        case 'name':
          return this.compare(a.full_name, b.full_name, isAsc);
        case 'department':
          return this.compare(a.department_name ?? '', b.department_name ?? '', isAsc);
        case 'inProgress':
          return this.compare(a.in_progress_count, b.in_progress_count, isAsc);
        case 'completed':
          return this.compare(a.completed_count, b.completed_count, isAsc);
        case 'overdue':
          return this.compare(a.overdue_count, b.overdue_count, isAsc);
        default:
          return 0;
      }
    });
  }

  protected trackByUserId(_: number, row: EmployeeLearningStats): string {
    return row.user_id;
  }

  protected get totalEmployees(): number {
    return this.allRows.length;
  }

  protected get completedEmployees(): number {
    return this.allRows.filter((item) => item.completed_count > 0).length;
  }

  private loadEmployees(): void {
    const requestId = ++this.requestSerial;
    this.loading = true;
    this.error = null;

    const query: EmployeesLearningStatsQuery = {
      scope: this.defaultScope(),
      search: this.searchTerm || undefined,
      limit: this.pageLimit,
      offset: 0,
    };

    this.usersApi.getEmployeesLearningStats(query).subscribe({
      next: (response) => {
        if (requestId !== this.requestSerial) {
          return;
        }
        this.allRows = response.items ?? [];
        this.tabCounts.all = response.total ?? this.allRows.length;
        this.sortedData = [...this.allRows];
        this.loading = false;
      },
      error: () => {
        if (requestId !== this.requestSerial) {
          return;
        }
        this.error = 'Не удалось загрузить список сотрудников';
        this.loading = false;
      },
    });
  }

  private defaultScope(): EmployeesScope {
    return this.authState.hasRole('hr') ? 'all' : 'team';
  }

  private compare(a: number | string, b: number | string, isAsc: boolean): number {
    if (typeof a === 'number' && typeof b === 'number') {
      return (a - b) * (isAsc ? 1 : -1);
    }
    return String(a).localeCompare(String(b), 'ru') * (isAsc ? 1 : -1);
  }
}
