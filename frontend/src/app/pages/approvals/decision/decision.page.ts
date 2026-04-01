import { CommonModule } from '@angular/common';
import { Component, OnInit, computed, inject, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule } from '@angular/forms';
import { ActivatedRoute } from '@angular/router';
import { MatButtonModule } from '@angular/material/button';
import { MatCardModule } from '@angular/material/card';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';

import { ExternalRequestsApiService } from '@core/api/external-requests-api.service';
import type { ExternalRequest } from '@core/api/contracts';
import { AuthStateService } from '@core/auth/auth-state.service';
import {
  canDoExternalRequestAction,
  externalRequestStatusLabel,
} from '@core/domain/external-request.workflow';

@Component({
  selector: 'app-page-approvals-decision',
  standalone: true,
  imports: [
    CommonModule,
    ReactiveFormsModule,
    MatButtonModule,
    MatCardModule,
    MatFormFieldModule,
    MatInputModule,
  ],
  templateUrl: './decision.page.html',
  styleUrl: './decision.page.scss',
})
export class ApprovalsDecisionPageComponent implements OnInit {
  private readonly api = inject(ExternalRequestsApiService);
  private readonly authState = inject(AuthStateService);
  private readonly route = inject(ActivatedRoute);
  private readonly fb = inject(FormBuilder);

  protected readonly loading = signal(true);
  protected readonly acting = signal(false);
  protected readonly error = signal<string | null>(null);
  protected readonly request = signal<ExternalRequest | null>(null);

  protected readonly form = this.fb.group({
    comment: [''],
  });

  protected readonly role = computed(() => this.authState.currentUser()?.roles[0] ?? 'employee');

  ngOnInit(): void {
    const id = this.route.snapshot.paramMap.get('requestId');
    if (!id) {
      this.error.set('Не найден requestId');
      this.loading.set(false);
      return;
    }

    this.api.getById(id).subscribe({
      next: (request) => {
        this.request.set(request);
        this.loading.set(false);
      },
      error: () => {
        this.error.set('Не удалось загрузить заявку');
        this.loading.set(false);
      },
    });
  }

  protected statusLabel(status: string): string {
    return externalRequestStatusLabel(status);
  }

  protected canApprove(): boolean {
    const item = this.request();
    return !!item && canDoExternalRequestAction(this.role(), item.status, 'approve');
  }

  protected canReject(): boolean {
    const item = this.request();
    return !!item && canDoExternalRequestAction(this.role(), item.status, 'reject');
  }

  protected canRevision(): boolean {
    const item = this.request();
    return !!item && canDoExternalRequestAction(this.role(), item.status, 'request_revision');
  }

  protected approve(): void {
    const item = this.request();
    if (!item || this.acting()) return;

    this.acting.set(true);
    this.api.approve(item.id, { comment: this.form.controls.comment.value ?? '' }).subscribe({
      next: (updated) => {
        this.request.set(updated);
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
    this.api.reject(item.id, { comment: this.form.controls.comment.value ?? '' }).subscribe({
      next: (updated) => {
        this.request.set(updated);
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
    this.api.requestRevision(item.id, { comment: this.form.controls.comment.value ?? '' }).subscribe({
      next: (updated) => {
        this.request.set(updated);
        this.acting.set(false);
      },
      error: () => {
        this.error.set('Не удалось вернуть заявку на доработку');
        this.acting.set(false);
      },
    });
  }
}
