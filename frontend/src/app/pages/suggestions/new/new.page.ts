import { CommonModule } from '@angular/common';
import { Component, OnInit, inject, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { Router, RouterLink } from '@angular/router';
import { MatButtonModule } from '@angular/material/button';
import { MatCardModule } from '@angular/material/card';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatIconModule } from '@angular/material/icon';
import { MatInputModule } from '@angular/material/input';
import { MatSelectModule } from '@angular/material/select';
import { catchError, of } from 'rxjs';

import { CourseSuggestionsApiService } from '@core/api/course-suggestions-api.service';
import { UsersApiService } from '@core/api/users-api.service';
import { AuthStateService } from '@core/auth/auth-state.service';
import { PERMISSIONS } from '@core/auth/permissions';
import { identityUserDisplayName } from '@core/domain/identity.util';
import type { IdentityUserView } from '@core/auth/auth.types';

@Component({
  selector: 'app-page-suggestions-new',
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
  templateUrl: './new.page.html',
  styleUrl: './new.page.scss',
})
export class SuggestionsNewPageComponent implements OnInit {
  private readonly suggestionsApi = inject(CourseSuggestionsApiService);
  private readonly usersApi = inject(UsersApiService);
  private readonly authState = inject(AuthStateService);
  private readonly fb = inject(FormBuilder);
  private readonly router = inject(Router);

  protected readonly loadingOptions = signal(true);
  protected readonly submitting = signal(false);
  protected readonly error = signal<string | null>(null);
  protected readonly directoryUnavailable = signal(false);
  protected readonly approvers = signal<IdentityUserView[]>([]);

  protected readonly hasDirectoryAccess =
    this.authState.hasPermission(PERMISSIONS.usersRead) || this.authState.hasAnyRole(['hr', 'admin']);

  protected readonly form = this.fb.group({
    title: ['', [Validators.required]],
    provider_name: [''],
    external_url: [''],
    description: [''],
    price: [''],
    price_currency: ['RUB', [Validators.required]],
    duration_hours: [''],
    approver_id: [''],
  });

  ngOnInit(): void {
    if (!this.hasDirectoryAccess) {
      this.directoryUnavailable.set(true);
      this.loadingOptions.set(false);
      return;
    }

    this.usersApi
      .listAdminUsers()
      .pipe(
        catchError(() => {
          this.directoryUnavailable.set(true);
          return of([]);
        }),
      )
      .subscribe({
        next: (users) => {
          this.approvers.set(users ?? []);
          this.loadingOptions.set(false);
        },
        error: () => {
          this.error.set('Не удалось подготовить форму предложения.');
          this.loadingOptions.set(false);
        },
      });
  }

  protected approverLabel(user: IdentityUserView): string {
    return identityUserDisplayName(user);
  }

  protected async submit(): Promise<void> {
    if (this.form.invalid || this.submitting()) {
      this.form.markAllAsTouched();
      return;
    }

    this.submitting.set(true);
    this.error.set(null);

    const values = this.form.getRawValue();
    const payload: Record<string, unknown> = {
      title: normalizeText(values.title),
      provider_name: normalizeText(values.provider_name),
      external_url: normalizeText(values.external_url),
      description: normalizeText(values.description),
      price: normalizeText(values.price),
      price_currency: normalizeText(values.price_currency) ?? 'RUB',
      duration_hours: normalizeText(values.duration_hours),
      approver_id: normalizeText(values.approver_id),
    };

    this.suggestionsApi.create(payload).subscribe({
      next: async (created) => {
        await this.router.navigateByUrl(`/suggestions/${created.id}`);
        this.submitting.set(false);
      },
      error: () => {
        this.error.set('Не удалось отправить предложение курса.');
        this.submitting.set(false);
      },
    });
  }
}

function normalizeText(value: string | null | undefined): string | undefined {
  const trimmed = value?.trim();
  return trimmed ? trimmed : undefined;
}
