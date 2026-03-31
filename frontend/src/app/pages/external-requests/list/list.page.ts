import { CommonModule } from '@angular/common';
import { Component, OnInit, inject, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule } from '@angular/forms';
import { RouterLink } from '@angular/router';
import { MatButtonModule } from '@angular/material/button';
import { MatCardModule } from '@angular/material/card';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatSelectModule } from '@angular/material/select';
import { catchError, map } from 'rxjs';

import { ExternalRequestsApiService } from '@core/api/external-requests-api.service';
import type { ExternalRequest } from '@core/api/contracts';
import { AuthStateService } from '@core/auth/auth-state.service';
import {
  externalRequestStatusLabel,
  externalRequestsScopeByRole,
} from '@core/domain/external-request.workflow';

@Component({
  selector: 'app-page-external-requests-list',
  standalone: true,
  imports: [
    CommonModule,
    ReactiveFormsModule,
    RouterLink,
    MatButtonModule,
    MatCardModule,
    MatFormFieldModule,
    MatInputModule,
    MatSelectModule,
  ],
  templateUrl: './list.page.html',
  styleUrl: './list.page.scss',
})
export class ExternalRequestsListPageComponent implements OnInit {
  private readonly api = inject(ExternalRequestsApiService);
  private readonly authState = inject(AuthStateService);
  private readonly fb = inject(FormBuilder);

  protected readonly loading = signal(true);
  protected readonly error = signal<string | null>(null);
  protected readonly requests = signal<ExternalRequest[]>([]);

  protected readonly statusOptions = [
    { value: 'draft', label: 'Черновик' },
    { value: 'manager_approval', label: 'На согласовании у руководителя' },
    { value: 'hr_approval', label: 'На согласовании у HR' },
    { value: 'approved', label: 'Согласовано' },
    { value: 'rejected', label: 'Отклонено' },
    { value: 'revision_requested', label: 'Требует доработки' },
    { value: 'completed', label: 'Завершено' },
  ];

  protected readonly form = this.fb.group({
    statuses: this.fb.control<string[]>([]),
    assignee: this.fb.control(''),
  });

  protected readonly role = signal<string>('employee');

  ngOnInit(): void {
    const primaryRole = this.authState.currentUser()?.roles[0] ?? 'employee';
    this.role.set(primaryRole);
    this.load();
  }

  protected applyFilters(): void {
    this.load();
  }

  protected showAwaitingApprovalsForEmployee(): void {
    this.form.patchValue({ statuses: ['manager_approval', 'hr_approval'] });
    this.load();
  }

  protected clearFilters(): void {
    this.form.setValue({ statuses: [], assignee: '' });
    this.load();
  }

  protected statusLabel(status: string): string {
    return externalRequestStatusLabel(status);
  }

  private load(): void {
    this.loading.set(true);
    this.error.set(null);

    const role = this.role();
    const scope = externalRequestsScopeByRole(role);
    const statuses = this.form.controls.statuses.value ?? [];
    const assignee = this.form.controls.assignee.value?.trim() || undefined;

    const request$ = this.api.listFiltered({ scope, statuses, assignee });

    const source$ =
      role === 'employee' && statuses.length > 0
        ? request$.pipe(
            catchError(() =>
              this.api
                .listMy()
                .pipe(map((rows) => rows.filter((row) => statuses.includes(row.status)))),
            ),
          )
        : request$;

    source$.subscribe({
      next: (requests) => {
        this.requests.set(requests ?? []);
        this.loading.set(false);
      },
      error: () => {
        this.error.set('Не удалось загрузить заявки');
        this.loading.set(false);
      },
    });
  }
}
