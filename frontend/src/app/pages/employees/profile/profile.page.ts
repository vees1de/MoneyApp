import { CommonModule } from '@angular/common';
import { Component, OnInit, inject, signal, computed } from '@angular/core';
import { ActivatedRoute, RouterLink } from '@angular/router';
import { MatButtonModule } from '@angular/material/button';
import { MatCardModule } from '@angular/material/card';
import { MatIconModule } from '@angular/material/icon';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatTabsModule } from '@angular/material/tabs';

import { resolveApiUrl } from '@core/api/url.util';
import { UsersApiService } from '@core/api/users-api.service';
import type {
  EmployeePublicProfile,
  EmployeeEnrollmentItem,
} from '@core/api/contracts';

@Component({
  selector: 'app-page-employee-profile',
  standalone: true,
  imports: [
    CommonModule,
    RouterLink,
    MatButtonModule,
    MatCardModule,
    MatIconModule,
    MatProgressSpinnerModule,
    MatTabsModule,
  ],
  templateUrl: './profile.page.html',
  styleUrl: './profile.page.scss',
})
export class EmployeeProfilePageComponent implements OnInit {
  private readonly route = inject(ActivatedRoute);
  private readonly api = inject(UsersApiService);

  protected readonly loading = signal(true);
  protected readonly error = signal<string | null>(null);
  protected readonly profile = signal<EmployeePublicProfile | null>(null);
  protected readonly enrollments = signal<EmployeeEnrollmentItem[]>([]);

  protected readonly fullName = computed(() => {
    const p = this.profile();
    if (!p) return '';
    const parts = [p.last_name, p.first_name, p.middle_name].filter(Boolean);
    return parts.join(' ') || p.display_name || p.email;
  });

  protected readonly avatarInitials = computed(() => {
    const p = this.profile();
    if (!p) return '';
    const first = p.first_name?.[0] ?? '';
    const last = p.last_name?.[0] ?? '';
    return (last + first).toUpperCase() || 'U';
  });

  protected readonly inProgressEnrollments = computed(() =>
    this.enrollments().filter((e) => e.status === 'in_progress' || e.status === 'enrolled'),
  );

  protected readonly completedEnrollments = computed(() =>
    this.enrollments().filter((e) => e.status === 'completed'),
  );

  protected readonly stats = computed(() => {
    const all = this.enrollments();
    return {
      total: all.length,
      inProgress: all.filter((e) => e.status === 'in_progress').length,
      completed: all.filter((e) => e.status === 'completed').length,
      mandatory: all.filter((e) => e.is_mandatory).length,
    };
  });

  ngOnInit(): void {
    const userId = this.route.snapshot.paramMap.get('userId');
    if (!userId) {
      this.error.set('ID сотрудника не указан');
      this.loading.set(false);
      return;
    }
    this.loadProfile(userId);
  }

  protected formatDate(value?: string | null): string {
    if (!value) return '—';
    const date = new Date(value);
    if (isNaN(date.getTime())) return value;
    return date.toLocaleDateString('ru-RU', { day: '2-digit', month: '2-digit', year: 'numeric' });
  }

  protected statusLabel(status: string): string {
    switch (status) {
      case 'enrolled': return 'Записан';
      case 'in_progress': return 'В процессе';
      case 'completed': return 'Завершён';
      case 'canceled': return 'Отменён';
      default: return status;
    }
  }

  protected statusClass(status: string): string {
    switch (status) {
      case 'completed': return 'status-badge--success';
      case 'in_progress': return 'status-badge--info';
      case 'enrolled': return 'status-badge--warning';
      case 'canceled': return 'status-badge--muted';
      default: return 'status-badge--muted';
    }
  }

  protected avatarUrl(value?: string | null): string | null {
    return resolveApiUrl(value);
  }

  private loadProfile(userId: string): void {
    this.loading.set(true);
    this.error.set(null);

    this.api.getEmployeeProfile(userId).subscribe({
      next: (res) => {
        this.profile.set(res.profile);
        this.enrollments.set(res.enrollments);
        this.loading.set(false);
      },
      error: () => {
        this.error.set('Не удалось загрузить профиль сотрудника');
        this.loading.set(false);
      },
    });
  }
}
