import { CommonModule } from '@angular/common';
import { Component, OnInit, computed, inject, signal } from '@angular/core';
import { RouterLink } from '@angular/router';
import { MatButtonModule } from '@angular/material/button';
import { MatCardModule } from '@angular/material/card';
import { MatIconModule } from '@angular/material/icon';
import { catchError, forkJoin, of } from 'rxjs';

import { WidgetShellComponent } from '@app/widgets/widget-shell/widget-shell.component';
import { CourseSuggestionsApiService } from '@core/api/course-suggestions-api.service';
import type { CourseSuggestion } from '@core/api/contracts';
import { UsersApiService } from '@core/api/users-api.service';
import type { IdentityUserView } from '@core/auth/auth.types';
import { resolveApiUrl } from '@core/api/url.util';
import { courseSuggestionStatusLabel } from '@core/domain/course-intakes.workflow';
import { identityUserDisplayName } from '@core/domain/identity.util';

type SuggestionFilter = 'all' | 'pending' | 'approved' | 'rejected' | 'intake_opened';

@Component({
  selector: 'app-hr-suggestions-widget',
  standalone: true,
  imports: [CommonModule, RouterLink, MatButtonModule, MatCardModule, MatIconModule, WidgetShellComponent],
  templateUrl: './hr-suggestions-widget.component.html',
  styleUrl: './hr-suggestions-widget.component.scss',
})
export class HrSuggestionsWidgetComponent implements OnInit {
  private readonly suggestionsApi = inject(CourseSuggestionsApiService);
  private readonly usersApi = inject(UsersApiService);

  protected readonly loading = signal(true);
  protected readonly error = signal<string | null>(null);
  protected readonly directoryUnavailable = signal(false);
  protected readonly selectedStatus = signal<SuggestionFilter>('pending');
  protected readonly suggestions = signal<CourseSuggestion[]>([]);
  protected readonly usersById = signal<Record<string, IdentityUserView>>({});

  protected readonly statuses: ReadonlyArray<{ key: SuggestionFilter; label: string }> = [
    { key: 'pending', label: 'На ревью HR' },
    { key: 'all', label: 'Все' },
    { key: 'approved', label: 'Одобрено' },
    { key: 'rejected', label: 'Отклонено' },
    { key: 'intake_opened', label: 'Набор открыт' },
  ];

  protected readonly filteredSuggestions = computed(() => {
    const status = this.selectedStatus();
    return this.suggestions().filter((item) => status === 'all' || item.status === status);
  });

  protected readonly visibleSuggestions = computed(() =>
    this.filteredSuggestions().slice(0, 6),
  );

  ngOnInit(): void {
    this.load();
  }

  protected setStatus(status: SuggestionFilter): void {
    this.selectedStatus.set(status);
  }

  protected statusCount(status: SuggestionFilter): number {
    if (status === 'all') {
      return this.suggestions().length;
    }
    return this.suggestions().filter((item) => item.status === status).length;
  }

  protected statusLabel(status: string): string {
    return courseSuggestionStatusLabel(status);
  }

  protected statusChipClass(status: string): string {
    switch (status) {
      case 'pending':
        return 'status-chip--pending';
      case 'approved':
      case 'intake_opened':
        return 'status-chip--approved';
      case 'rejected':
        return 'status-chip--rejected';
      default:
        return 'status-chip--neutral';
    }
  }

  protected authorLabel(userId: string | null | undefined): string {
    if (!userId) {
      return '—';
    }

    const user = this.usersById()[userId];
    return user ? identityUserDisplayName(user) : `Сотрудник ${userId.slice(0, 8)}`;
  }

  protected authorAvatarUrl(userId: string | null | undefined): string | null {
    if (!userId) {
      return null;
    }
    return resolveApiUrl(this.usersById()[userId]?.avatar_url);
  }

  protected authorAvatarInitials(userId: string | null | undefined): string {
    if (!userId) {
      return 'U';
    }

    const user = this.usersById()[userId];
    if (!user) {
      return 'U';
    }

    const profile = user.employee_profile;
    if (profile) {
      const initials = `${profile.last_name?.[0] ?? ''}${profile.first_name?.[0] ?? ''}`
        .trim()
        .toUpperCase();
      if (initials) {
        return initials;
      }
    }

    const emailInitial = user.email?.trim()?.[0]?.toUpperCase();
    return emailInitial || 'U';
  }

  private load(): void {
    this.loading.set(true);
    this.error.set(null);
    this.directoryUnavailable.set(false);

    forkJoin({
      suggestions: this.suggestionsApi.list().pipe(catchError(() => of([]))),
      users: this.usersApi.listAdminUsers().pipe(
        catchError(() => {
          this.directoryUnavailable.set(true);
          return of([]);
        }),
      ),
    }).subscribe({
      next: ({ suggestions, users }) => {
        this.usersById.set(
          (users ?? []).reduce<Record<string, IdentityUserView>>((accumulator, user) => {
            accumulator[user.id] = user;
            return accumulator;
          }, {}),
        );

        const sorted = [...(suggestions ?? [])].sort((left, right) => {
          const statusDiff =
            suggestionStatusOrder(left.status) - suggestionStatusOrder(right.status);
          if (statusDiff !== 0) {
            return statusDiff;
          }
          return new Date(right.updated_at).getTime() - new Date(left.updated_at).getTime();
        });

        this.suggestions.set(sorted);
        this.loading.set(false);
      },
      error: () => {
        this.error.set('Не удалось загрузить предложения курсов.');
        this.loading.set(false);
      },
    });
  }
}

function suggestionStatusOrder(status: string): number {
  if (status === 'pending') {
    return 0;
  }
  if (status === 'approved' || status === 'intake_opened') {
    return 1;
  }
  if (status === 'rejected') {
    return 2;
  }
  return 3;
}
