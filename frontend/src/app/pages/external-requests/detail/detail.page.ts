import { CommonModule } from '@angular/common';
import { Component, OnInit, computed, inject, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule } from '@angular/forms';
import { ActivatedRoute, RouterLink } from '@angular/router';
import { MatButtonModule } from '@angular/material/button';
import { MatCardModule } from '@angular/material/card';
import { MatCheckboxModule } from '@angular/material/checkbox';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatIconModule } from '@angular/material/icon';
import { MatInputModule } from '@angular/material/input';

import { TrainingBudgetPreviewComponent } from '@app/shared/training-budget-preview/training-budget-preview.component';
import { ExternalRequestsApiService } from '@core/api/external-requests-api.service';
import type { ExternalRequest } from '@core/api/contracts';
import { AuthStateService } from '@core/auth/auth-state.service';
import { TrainingBudgetMockService } from '@core/domain/training-budget-mock.service';
import {
  canDoExternalRequestAction,
  externalRequestStatusLabel,
} from '@core/domain/external-request.workflow';

@Component({
  selector: 'app-page-external-requests-detail',
  standalone: true,
  imports: [
    CommonModule,
    ReactiveFormsModule,
    RouterLink,
    MatButtonModule,
    MatCardModule,
    MatCheckboxModule,
    MatFormFieldModule,
    MatIconModule,
    MatInputModule,
    TrainingBudgetPreviewComponent,
  ],
  templateUrl: './detail.page.html',
  styleUrl: './detail.page.scss',
})
export class ExternalRequestsDetailPageComponent implements OnInit {
  private readonly api = inject(ExternalRequestsApiService);
  private readonly authState = inject(AuthStateService);
  private readonly budgetMock = inject(TrainingBudgetMockService);
  private readonly route = inject(ActivatedRoute);
  private readonly fb = inject(FormBuilder);

  protected readonly loading = signal(true);
  protected readonly acting = signal(false);
  protected readonly error = signal<string | null>(null);
  protected readonly info = signal<string | null>(null);
  protected readonly request = signal<ExternalRequest | null>(null);

  protected readonly form = this.fb.group({
    comment: [''],
    requires_additional_approval: [false],
    additional_approval_comment: [''],
  });

  protected readonly role = computed(() => this.authState.currentUser()?.roles[0] ?? 'employee');
  protected readonly budgetView = computed(() => {
    const item = this.request();
    return item ? this.budgetMock.getRequestView(item.id) : null;
  });

  protected readonly submitBlockedByBudget = computed(() => {
    const view = this.budgetView();
    return !!view?.validation.requiresAdditionalApproval && !this.form.controls.requires_additional_approval.value;
  });

  ngOnInit(): void {
    this.load();
  }

  protected statusLabel(status: string): string {
    return externalRequestStatusLabel(status);
  }

  protected canSubmit(): boolean {
    const item = this.request();
    return !!item && canDoExternalRequestAction(this.role(), item.status, 'submit');
  }

  protected canApprove(): boolean {
    const item = this.request();
    return !!item && canDoExternalRequestAction(this.role(), item.status, 'approve');
  }

  protected canReject(): boolean {
    const item = this.request();
    return !!item && canDoExternalRequestAction(this.role(), item.status, 'reject');
  }

  protected canRequestRevision(): boolean {
    const item = this.request();
    return !!item && canDoExternalRequestAction(this.role(), item.status, 'request_revision');
  }

  protected canMarkCancelled(): boolean {
    const item = this.request();
    return (
      !!item &&
      this.role() === 'employee' &&
      ['approved', 'completed'].includes(item.status) &&
      this.budgetMock.hasActiveAllocation(item.id)
    );
  }

  protected submitRequest(): void {
    const item = this.request();
    const view = this.budgetView();
    if (!item || !view || this.acting()) return;

    if (view.validation.requiresAdditionalApproval && !this.form.controls.requires_additional_approval.value) {
      this.error.set('Бюджета недостаточно. Отметьте дополнительное согласование перед отправкой.');
      this.form.controls.requires_additional_approval.markAsTouched();
      return;
    }

    if (
      view.validation.requiresAdditionalApproval &&
      this.form.controls.requires_additional_approval.value &&
      !this.form.controls.additional_approval_comment.value?.trim()
    ) {
      this.error.set('Добавьте комментарий, почему нужно дополнительное согласование.');
      this.form.controls.additional_approval_comment.markAsTouched();
      return;
    }

    this.acting.set(true);
    this.error.set(null);
    this.info.set(null);

    this.api.submit(item.id).subscribe({
      next: (updated) => {
        this.request.set(updated);
        this.budgetMock.syncRequest(updated, {
          departmentId: view.trackedRequest.departmentId,
          departmentName: view.trackedRequest.departmentName,
          employeeId: view.trackedRequest.employeeId,
          employeeName: view.trackedRequest.employeeName,
          employeeEmail: view.trackedRequest.employeeEmail,
          requiresAdditionalApproval: this.form.controls.requires_additional_approval.value ?? undefined,
          additionalApprovalComment: this.form.controls.additional_approval_comment.value,
        });
        this.acting.set(false);
      },
      error: () => {
        this.error.set('Не удалось отправить заявку');
        this.acting.set(false);
      },
    });
  }

  protected approve(): void {
    const item = this.request();
    if (!item || this.acting()) return;

    this.acting.set(true);
    this.error.set(null);
    this.info.set(null);

    this.api.approve(item.id, { comment: this.form.controls.comment.value ?? '' }).subscribe({
      next: (updated) => {
        this.request.set(updated);
        this.budgetMock.syncRequest(updated);
        this.acting.set(false);
      },
      error: () => {
        this.error.set('Не удалось согласовать заявку');
        this.acting.set(false);
      },
    });
  }

  protected reject(): void {
    const item = this.request();
    if (!item || this.acting()) return;

    this.acting.set(true);
    this.error.set(null);
    this.info.set(null);

    this.api.reject(item.id, { comment: this.form.controls.comment.value ?? '' }).subscribe({
      next: (updated) => {
        this.request.set(updated);
        this.budgetMock.syncRequest(updated);
        this.info.set('Сумма возвращена в квоту сотрудника и бюджет отдела.');
        this.acting.set(false);
      },
      error: () => {
        this.error.set('Не удалось отклонить заявку');
        this.acting.set(false);
      },
    });
  }

  protected requestRevision(): void {
    const item = this.request();
    if (!item || this.acting()) return;

    this.acting.set(true);
    this.error.set(null);
    this.info.set(null);

    this.api
      .requestRevision(item.id, { comment: this.form.controls.comment.value ?? '' })
      .subscribe({
        next: (updated) => {
          this.request.set(updated);
          this.budgetMock.syncRequest(updated);
          this.info.set('Бюджетный резерв снят, сумма возвращена в доступные лимиты.');
          this.acting.set(false);
        },
        error: () => {
          this.error.set('Не удалось вернуть на доработку');
          this.acting.set(false);
        },
      });
  }

  protected markCancelled(): void {
    const item = this.request();
    if (!item) return;

    this.budgetMock.markCourseCancelled(item.id, this.form.controls.comment.value ?? null);
    this.info.set('Mock: курс отменён, сумма вернулась в бюджет отдела и персональную квоту.');
    this.error.set(null);
  }

  private load(): void {
    const id = this.route.snapshot.paramMap.get('requestId');
    if (!id) {
      this.error.set('Не найден идентификатор заявки');
      this.loading.set(false);
      return;
    }

    this.loading.set(true);
    this.error.set(null);
    this.info.set(null);

    this.api.getById(id).subscribe({
      next: (item) => {
        this.request.set(item);
        this.budgetMock.syncRequest(item);

        const view = this.budgetMock.getRequestView(item.id);
        if (view) {
          this.form.patchValue({
            requires_additional_approval: view.trackedRequest.requiresAdditionalApproval,
            additional_approval_comment: view.trackedRequest.additionalApprovalComment ?? '',
          });
        }

        this.loading.set(false);
      },
      error: () => {
        this.error.set('Не удалось загрузить заявку');
        this.loading.set(false);
      },
    });
  }
}
