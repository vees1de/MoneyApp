import { CommonModule } from '@angular/common';
import { HttpErrorResponse } from '@angular/common/http';
import { Component, inject, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { Router } from '@angular/router';
import { MatButtonModule } from '@angular/material/button';
import { MatCardModule } from '@angular/material/card';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { firstValueFrom } from 'rxjs';

import { AuthApiService } from '@core/auth/auth-api.service';
import { AuthSessionService } from '@core/auth/auth-session.service';
import { AuthStateService } from '@core/auth/auth-state.service';
import type { LoginResponse } from '@core/auth/auth.types';

interface AuthApiErrorPayload {
  error?: {
    code?: string;
    message?: string;
  };
  message?: string;
}

@Component({
  selector: 'app-page-public-login',
  standalone: true,
  imports: [
    CommonModule,
    ReactiveFormsModule,
    MatButtonModule,
    MatCardModule,
    MatFormFieldModule,
    MatInputModule,
  ],
  templateUrl: './login.page.html',
  styleUrl: './login.page.scss',
})
export class PublicLoginPageComponent {
  private readonly fb = inject(FormBuilder);
  private readonly authApi = inject(AuthApiService);
  private readonly authSession = inject(AuthSessionService);
  private readonly authState = inject(AuthStateService);
  private readonly router = inject(Router);

  protected readonly submitting = signal(false);
  protected readonly authError = signal<string | null>(null);

  protected readonly form = this.fb.group({
    email: ['', [Validators.required, Validators.email]],
    password: ['', [Validators.required]],
  });

  protected get emailControl() {
    return this.form.controls.email;
  }

  protected get passwordControl() {
    return this.form.controls.password;
  }

  protected emailError(): string {
    if (!this.emailControl.touched && !this.emailControl.dirty) {
      return '';
    }
    if (this.emailControl.hasError('required')) {
      return 'Введите email.';
    }
    if (this.emailControl.hasError('email')) {
      return 'Введите корректный email.';
    }
    return '';
  }

  protected passwordError(): string {
    if (!this.passwordControl.touched && !this.passwordControl.dirty) {
      return '';
    }
    if (this.passwordControl.hasError('required')) {
      return 'Введите пароль.';
    }
    return '';
  }

  protected async submit(): Promise<void> {
    if (this.form.invalid || this.submitting()) {
      this.form.markAllAsTouched();
      return;
    }

    this.submitting.set(true);
    this.authError.set(null);

    try {
      const payload = this.form.getRawValue();
      const response = await firstValueFrom(
        this.authApi.login({
          email: payload.email ?? '',
          password: payload.password ?? '',
        }),
      );

      await this.completeAuthentication(response);
    } catch (error) {
      this.authError.set(this.resolveAuthError(error));
    } finally {
      this.submitting.set(false);
    }
  }

  private async completeAuthentication(response: LoginResponse): Promise<void> {
    this.authSession.setTokens(response.tokens);
    this.authState.setCurrentUser(response.user);
    this.authSession.setUserSnapshot(response.user);

    await this.router.navigateByUrl('/dashboard');
  }

  private resolveAuthError(error: unknown): string {
    const fallback = 'Не удалось войти. Проверьте логин и пароль.';
    if (!(error instanceof HttpErrorResponse)) {
      return fallback;
    }

    const apiError = this.extractApiError(error);
    if (error.status === 401 || apiError?.code === 'invalid_credentials') {
      return fallback;
    }

    return apiError?.message ?? fallback;
  }

  private extractApiError(error: HttpErrorResponse): AuthApiErrorPayload['error'] | null {
    const payload = error.error;
    if (!payload || typeof payload !== 'object') {
      return null;
    }

    const typedPayload = payload as AuthApiErrorPayload;
    if (typedPayload.error) {
      return typedPayload.error;
    }
    if (typedPayload.message) {
      return { message: typedPayload.message };
    }

    return null;
  }
}
