import { CommonModule } from '@angular/common';
import { Component, OnInit, computed, inject, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule } from '@angular/forms';
import { ActivatedRoute, RouterLink } from '@angular/router';
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
  selector: 'app-page-external-requests-detail',
  standalone: true,
  imports: [
    CommonModule,
    ReactiveFormsModule,
    RouterLink,
    MatButtonModule,
    MatCardModule,
    MatFormFieldModule,
    MatInputModule,
  ],
  templateUrl: './detail.page.html',
  styleUrl: './detail.page.scss',
})
export class ExternalRequestsDetailPageComponent implements OnInit {
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

  protected submitRequest(): void {
    const item = this.request();
    if (!item || this.acting()) return;

    this.acting.set(true);
    this.error.set(null);

    this.api.submit(item.id).subscribe({
      next: (updated) => {
        this.request.set(updated);
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
    this.error.set(null);

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
    this.error.set(null);

    this.api
      .requestRevision(item.id, { comment: this.form.controls.comment.value ?? '' })
      .subscribe({
        next: (updated) => {
          this.request.set(updated);
          this.acting.set(false);
        },
        error: () => {
          this.error.set('Не удалось вернуть на доработку');
          this.acting.set(false);
        },
      });
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

    this.api.getById(id).subscribe({
      next: (item) => {
        this.request.set(item);
        this.loading.set(false);
      },
      error: () => {
        this.error.set('Не удалось загрузить заявку');
        this.loading.set(false);
      },
    });
  }
}
