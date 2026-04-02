import { CommonModule } from '@angular/common';
import { Component, computed, inject, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { Router, RouterLink } from '@angular/router';
import { MatButtonModule } from '@angular/material/button';
import { MatCardModule } from '@angular/material/card';
import { MatCheckboxModule } from '@angular/material/checkbox';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatIconModule } from '@angular/material/icon';
import { MatInputModule } from '@angular/material/input';

import { TrainingBudgetPreviewComponent } from '@app/shared/training-budget-preview/training-budget-preview.component';
import { ExternalRequestsApiService } from '@core/api/external-requests-api.service';
import { AuthStateService } from '@core/auth/auth-state.service';
import { TrainingBudgetMockService } from '@core/domain/training-budget-mock.service';

@Component({
  selector: 'app-page-external-requests-new',
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
  templateUrl: './new.page.html',
  styleUrl: './new.page.scss',
})
export class ExternalRequestsNewPageComponent {
  private readonly api = inject(ExternalRequestsApiService);
  private readonly authState = inject(AuthStateService);
  private readonly budgetMock = inject(TrainingBudgetMockService);
  private readonly fb = inject(FormBuilder);
  private readonly router = inject(Router);

  private readonly actorContext = this.budgetMock.ensureActorContext({
    departmentId: this.authState.currentUser()?.employee_profile?.department_id ?? 'my-department',
    departmentName: 'Мой отдел',
    employeeId: this.authState.currentUser()?.id ?? 'current-user',
    employeeName: this.authState.displayName(),
    employeeEmail: this.authState.currentUser()?.email ?? null,
  });

  protected readonly submitting = signal(false);
  protected readonly error = signal<string | null>(null);

  protected readonly form = this.fb.group({
    title: ['', [Validators.required]],
    provider_name: ['', [Validators.required]],
    course_url: ['', [Validators.required]],
    program_description: ['', [Validators.required]],
    planned_start_date: [''],
    planned_end_date: [''],
    duration_hours: [''],
    cost_amount: ['', [Validators.required]],
    currency: ['RUB', [Validators.required]],
    business_goal: [''],
    employee_comment: [''],
    requires_additional_approval: [false],
    additional_approval_comment: [''],
  });

  protected readonly amountValue = computed(() => {
    const raw = this.form.controls.cost_amount.value ?? '0';
    const normalized = Number.parseFloat(String(raw).replace(/\s+/g, '').replace(',', '.'));
    return Number.isFinite(normalized) && normalized > 0 ? Math.round(normalized) : 0;
  });

  protected readonly budgetPreview = computed(() =>
    this.budgetMock.previewForActor(
      this.actorContext,
      this.amountValue(),
      this.form.controls.currency.value || 'RUB',
    ),
  );

  protected readonly requiresAdditionalApproval = computed(
    () => this.budgetPreview().validation.requiresAdditionalApproval,
  );

  protected async submit(): Promise<void> {
    if (this.form.invalid || this.submitting()) {
      this.form.markAllAsTouched();
      return;
    }

    if (
      this.requiresAdditionalApproval() &&
      !this.form.controls.requires_additional_approval.value
    ) {
      this.error.set(
        'Budget is exceeded. Mark additional approval before the request can move forward.',
      );
      this.form.controls.requires_additional_approval.markAsTouched();
      return;
    }

    if (
      this.requiresAdditionalApproval() &&
      this.form.controls.requires_additional_approval.value &&
      !this.form.controls.additional_approval_comment.value?.trim()
    ) {
      this.error.set('Add a short justification for the extra approval.');
      this.form.controls.additional_approval_comment.markAsTouched();
      return;
    }

    this.submitting.set(true);
    this.error.set(null);

    const {
      requires_additional_approval,
      additional_approval_comment,
      ...payload
    } = this.form.getRawValue();

    this.api.create(payload).subscribe({
      next: async (created) => {
        this.budgetMock.storeDraftRequest({
          requestId: created.id,
          title: created.title,
          providerName: created.provider_name,
          amount: this.amountValue(),
          currency: created.currency ?? payload.currency ?? 'RUB',
          departmentId: created.department_id ?? this.actorContext.departmentId,
          departmentName: created.department_name ?? this.actorContext.departmentName,
          employeeId: created.employee_user_id ?? this.actorContext.employeeId,
          employeeName:
            created.employee_full_name ||
            created.employee_email ||
            this.actorContext.employeeName,
          employeeEmail: created.employee_email ?? this.actorContext.employeeEmail,
          requiresAdditionalApproval: requires_additional_approval ?? undefined,
          additionalApprovalComment: additional_approval_comment,
          lastKnownStatus: created.status,
        });

        await this.router.navigateByUrl(`/external-requests/${created.id}`);
        this.submitting.set(false);
      },
      error: () => {
        this.error.set('Не удалось создать заявку');
        this.submitting.set(false);
      },
    });
  }
}
