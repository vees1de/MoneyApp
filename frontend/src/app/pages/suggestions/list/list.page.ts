import { CommonModule } from '@angular/common';
import { Component, OnInit, computed, inject, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule } from '@angular/forms';
import { RouterLink } from '@angular/router';
import { MatButtonModule } from '@angular/material/button';
import { MatCardModule } from '@angular/material/card';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatIconModule } from '@angular/material/icon';
import { MatInputModule } from '@angular/material/input';
import { MatSelectModule } from '@angular/material/select';
import { catchError, forkJoin, of } from 'rxjs';

import { CourseSuggestionsApiService } from '@core/api/course-suggestions-api.service';
import { UsersApiService } from '@core/api/users-api.service';
import type { CourseSuggestion } from '@core/api/contracts';
import { AuthStateService } from '@core/auth/auth-state.service';
import { PERMISSIONS } from '@core/auth/permissions';
import {
  canReviewSuggestion,
  courseSuggestionStatusLabel,
} from '@core/domain/course-intakes.workflow';
import { identityUserDisplayName } from '@core/domain/identity.util';
import type { IdentityUserView } from '@core/auth/auth.types';

@Component({
  selector: 'app-page-suggestions-list',
  standalone: true,
  imports: [
    CommonModule,
    ReactiveFormsModule,
    RouterLink,
    MatButtonModule,
    MatCardModule,
    MatFormFieldModule,
    MatIconModule,
    MatInputModule,
    MatSelectModule,
  ],
  templateUrl: './list.page.html',
  styleUrl: './list.page.scss',
})
export class SuggestionsListPageComponent implements OnInit {
  private readonly suggestionsApi = inject(CourseSuggestionsApiService);
  private readonly usersApi = inject(UsersApiService);
  private readonly authState = inject(AuthStateService);
  private readonly fb = inject(FormBuilder);

  protected readonly loading = signal(true);
  protected readonly error = signal<string | null>(null);
  protected readonly suggestions = signal<CourseSuggestion[]>([]);
  protected readonly usersById = signal<Record<string, IdentityUserView>>({});

  protected readonly canReview = computed(() => {
    const role = this.authState.currentUser()?.roles[0] ?? 'employee';
    return canReviewSuggestion(role) || this.authState.hasPermission(PERMISSIONS.intakesManage);
  });

  protected readonly form = this.fb.group({
    status: [''],
    query: [''],
  });

  protected readonly statusOptions = [
    { value: '', label: 'Все статусы' },
    { value: 'pending', label: 'На ревью HR' },
    { value: 'approved', label: 'Одобрено' },
    { value: 'rejected', label: 'Отклонено' },
    { value: 'intake_opened', label: 'Набор открыт' },
  ];

  ngOnInit(): void {
    this.load();
  }

  protected applyFilters(): void {
    this.load();
  }

  protected clearFilters(): void {
    this.form.setValue({ status: '', query: '' });
    this.load();
  }

  protected statusLabel(status: string): string {
    return courseSuggestionStatusLabel(status);
  }

  protected statusChipClass(status: string): string {
    switch (status) {
      case 'approved':
      case 'intake_opened':
        return 'suggestion-status-chip--approved';
      case 'rejected':
        return 'suggestion-status-chip--rejected';
      case 'pending':
        return 'suggestion-status-chip--pending';
      default:
        return 'suggestion-status-chip--neutral';
    }
  }

  protected statusCardClass(status: string): string {
    switch (status) {
      case 'approved':
      case 'intake_opened':
        return 'suggestion-card--approved';
      case 'rejected':
        return 'suggestion-card--rejected';
      case 'pending':
        return 'suggestion-card--pending';
      default:
        return 'suggestion-card--neutral';
    }
  }

  protected statusIcon(status: string): string {
    switch (status) {
      case 'approved':
        return 'check_circle';
      case 'intake_opened':
        return 'task_alt';
      case 'rejected':
        return 'cancel';
      case 'pending':
        return 'schedule';
      default:
        return 'help';
    }
  }

  protected userLabel(userId: string | null | undefined): string {
    if (!userId) {
      return '—';
    }

    const currentUser = this.authState.currentUser();
    if (currentUser?.id === userId) {
      return identityUserDisplayName(currentUser);
    }

    const user = this.usersById()[userId];
    return user ? identityUserDisplayName(user) : `Сотрудник ${userId.slice(0, 8)}`;
  }

  private load(): void {
    this.loading.set(true);
    this.error.set(null);

    const status = normalizeText(this.form.controls.status.value);
    const query = (this.form.controls.query.value ?? '').trim().toLowerCase();
    const suggestions$ = this.canReview()
      ? this.suggestionsApi.list(status)
      : this.suggestionsApi.listMy().pipe(
          catchError(() => this.suggestionsApi.list(status)),
        );

    forkJoin({
      suggestions: suggestions$,
      users: this.canReview()
        ? this.usersApi.listAdminUsers().pipe(catchError(() => of([])))
        : of([]),
    }).subscribe({
      next: ({ suggestions, users }) => {
        const filtered = (suggestions ?? []).filter((item) => {
          const statusMatches = status ? item.status === status : true;
          const queryMatches = !query
            ? true
            : [item.title, item.description ?? '', item.provider_name ?? '', item.status]
                .join(' ')
                .toLowerCase()
                .includes(query);

          return statusMatches && queryMatches;
        });

        const usersById = (users ?? []).reduce<Record<string, IdentityUserView>>((accumulator, user) => {
          accumulator[user.id] = user;
          return accumulator;
        }, {});

        this.usersById.set(usersById);
        this.suggestions.set(filtered ?? []);
        this.loading.set(false);
      },
      error: () => {
        this.error.set('Не удалось загрузить предложения курсов.');
        this.loading.set(false);
      },
    });
  }
}

function normalizeText(value: string | null | undefined): string | undefined {
  const trimmed = value?.trim();
  return trimmed ? trimmed : undefined;
}
