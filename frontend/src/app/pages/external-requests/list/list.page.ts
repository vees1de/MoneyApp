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
import { catchError, map } from 'rxjs';

import { TrainingBudgetPreviewComponent } from '@app/shared/training-budget-preview/training-budget-preview.component';
import { ExternalRequestsApiService } from '@core/api/external-requests-api.service';
import type { ExternalRequest } from '@core/api/contracts';
import { AuthStateService } from '@core/auth/auth-state.service';
import { TrainingBudgetMockService } from '@core/domain/training-budget-mock.service';
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
    MatIconModule,
    MatInputModule,
    MatSelectModule,
    TrainingBudgetPreviewComponent,
  ],
  templateUrl: './list.page.html',
  styleUrl: './list.page.scss',
})
export class ExternalRequestsListPageComponent implements OnInit {
  private readonly api = inject(ExternalRequestsApiService);
  private readonly authState = inject(AuthStateService);
  private readonly budgetMock = inject(TrainingBudgetMockService);
  private readonly fb = inject(FormBuilder);

  private readonly actorContext = this.budgetMock.ensureActorContext({
    departmentId: this.authState.currentUser()?.employee_profile?.department_id ?? 'my-department',
    departmentName: 'Мой отдел',
    employeeId: this.authState.currentUser()?.id ?? 'current-user',
    employeeName: this.authState.displayName(),
    employeeEmail: this.authState.currentUser()?.email ?? null,
  });

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
  protected readonly currentBudget = computed(() =>
    this.budgetMock.previewForActor(this.actorContext, 0, 'RUB'),
  );

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

  protected requestBudgetLabel(requestId: string): string {
    const view = this.budgetMock.getRequestView(requestId);
    if (!view) {
      return 'Бюджет не рассчитан';
    }

    if (view.trackedRequest.requiresAdditionalApproval || view.validation.requiresAdditionalApproval) {
      return 'Нужно доп. согласование';
    }

    if (view.latestExpense?.status === 'reserved') {
      return 'Сумма в резерве';
    }

    if (view.latestExpense?.status === 'approved' || view.latestExpense?.status === 'completed') {
      return 'Сумма подтверждена';
    }

    return 'В пределах бюджета';
  }

  protected requestBudgetTone(requestId: string): 'ok' | 'warning' {
    const view = this.budgetMock.getRequestView(requestId);
    return view && (view.trackedRequest.requiresAdditionalApproval || view.validation.requiresAdditionalApproval)
      ? 'warning'
      : 'ok';
  }

  protected requestBudgetMeta(requestId: string): { department: number; employee: number } | null {
    const view = this.budgetMock.getRequestView(requestId);
    if (!view) return null;

    return {
      department: view.validation.departmentRemainingAmount,
      employee: view.validation.employeeRemainingAmount,
    };
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
        (requests ?? []).forEach((item) => this.budgetMock.syncRequest(item));
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
