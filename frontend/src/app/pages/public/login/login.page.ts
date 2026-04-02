import { CommonModule } from '@angular/common';
import { HttpErrorResponse } from '@angular/common/http';
import { Component, computed, inject, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, ValidatorFn, Validators } from '@angular/forms';
import { Router } from '@angular/router';
import { MatButtonModule } from '@angular/material/button';
import { MatCardModule } from '@angular/material/card';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { firstValueFrom } from 'rxjs';

import { AuthApiService } from '@core/auth/auth-api.service';
import { AuthSessionService } from '@core/auth/auth-session.service';
import { AuthStateService } from '@core/auth/auth-state.service';
import type { LoginRequest, LoginResponse } from '@core/auth/auth.types';

type AuthMode = 'login' | 'register';
type DemoAccount = 'employee' | 'hr';
type AuthAction = 'form' | DemoAccount;

interface AuthApiErrorPayload {
  error?: {
    code?: string;
    message?: string;
  };
  message?: string;
}

const DEMO_CREDENTIALS: Record<DemoAccount, LoginRequest> = {
  employee: {
    email: 'employee.demo@moneyapp.local',
    password: 'DemoEmployee123!',
  },
  hr: {
    email: 'hr.demo@moneyapp.local',
    password: 'DemoHR123!',
  },
};

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
    MatProgressSpinnerModule,
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
  private readonly passwordMatchValidator: ValidatorFn = (group) => {
    if (!this.isRegisterMode()) {
      return null;
    }

    const password = group.get('password')?.value ?? '';
    const confirmPassword = group.get('confirmPassword')?.value ?? '';
    if (!password || !confirmPassword || password === confirmPassword) {
      return null;
    }

    return { passwordMismatch: true };
  };

  protected readonly mode = signal<AuthMode>('login');
  protected readonly isRegisterMode = computed(() => this.mode() === 'register');
  protected readonly submitting = signal(false);
  protected readonly activeAction = signal<AuthAction | null>(null);
  protected readonly authError = signal<string | null>(null);

  protected readonly form = this.fb.group(
    {
      email: ['', [Validators.required, Validators.email]],
      password: ['', Validators.required],
      confirmPassword: [''],
      firstName: [''],
      lastName: [''],
      middleName: [''],
      positionTitle: [''],
    },
    { validators: this.passwordMatchValidator },
  );

  constructor() {
    this.syncModeValidators();
  }

  protected get emailControl() {
    return this.form.controls.email;
  }

  protected get passwordControl() {
    return this.form.controls.password;
  }

  protected get confirmPasswordControl() {
    return this.form.controls.confirmPassword;
  }

  protected get firstNameControl() {
    return this.form.controls.firstName;
  }

  protected get lastNameControl() {
    return this.form.controls.lastName;
  }

  protected setMode(mode: AuthMode): void {
    if (this.mode() === mode) {
      return;
    }

    this.mode.set(mode);
    this.authError.set(null);
    this.syncModeValidators();
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
    if (this.passwordControl.hasError('minlength')) {
      return 'Пароль должен быть не короче 8 символов.';
    }
    return '';
  }

  protected confirmPasswordError(): string {
    if (!this.isRegisterMode()) {
      return '';
    }
    if (!this.confirmPasswordControl.touched && !this.confirmPasswordControl.dirty) {
      return '';
    }
    if (this.confirmPasswordControl.hasError('required')) {
      return 'Подтвердите пароль.';
    }
    if (this.form.hasError('passwordMismatch')) {
      return 'Пароли не совпадают.';
    }
    return '';
  }

  protected firstNameError(): string {
    if (!this.isRegisterMode()) {
      return '';
    }
    if (!this.firstNameControl.touched && !this.firstNameControl.dirty) {
      return '';
    }
    if (this.firstNameControl.hasError('required')) {
      return 'Введите имя.';
    }
    return '';
  }

  protected lastNameError(): string {
    if (!this.isRegisterMode()) {
      return '';
    }
    if (!this.lastNameControl.touched && !this.lastNameControl.dirty) {
      return '';
    }
    if (this.lastNameControl.hasError('required')) {
      return 'Введите фамилию.';
    }
    return '';
  }

  protected async submit(): Promise<void> {
    this.syncModeValidators();

    if (this.submitting()) {
      return;
    }

    if (this.form.invalid) {
      this.form.markAllAsTouched();
      return;
    }

    const payload = this.form.getRawValue();
    const authMode = this.mode();

    await this.runAuthAction(
      'form',
      () =>
        authMode === 'register'
          ? firstValueFrom(
              this.authApi.register({
                email: payload.email ?? '',
                password: payload.password ?? '',
                first_name: payload.firstName?.trim() ?? '',
                last_name: payload.lastName?.trim() ?? '',
                middle_name: this.optionalValue(payload.middleName),
                position_title: this.optionalValue(payload.positionTitle),
              }),
            )
          : firstValueFrom(
              this.authApi.login({
                email: payload.email ?? '',
                password: payload.password ?? '',
              }),
            ),
      authMode,
    );
  }

  protected async loginAsDemo(account: DemoAccount): Promise<void> {
    if (this.submitting()) {
      return;
    }

    await this.runAuthAction(
      account,
      () => firstValueFrom(this.authApi.login(DEMO_CREDENTIALS[account])),
      'login',
    );
  }

  protected isActionPending(action: AuthAction): boolean {
    return this.activeAction() === action;
  }

  private syncModeValidators(): void {
    const registerMode = this.isRegisterMode();

    this.passwordControl.setValidators(
      registerMode ? [Validators.required, Validators.minLength(8)] : [Validators.required],
    );
    this.confirmPasswordControl.setValidators(registerMode ? [Validators.required] : []);
    this.firstNameControl.setValidators(registerMode ? [Validators.required] : []);
    this.lastNameControl.setValidators(registerMode ? [Validators.required] : []);

    this.passwordControl.updateValueAndValidity({ emitEvent: false });
    this.confirmPasswordControl.updateValueAndValidity({ emitEvent: false });
    this.firstNameControl.updateValueAndValidity({ emitEvent: false });
    this.lastNameControl.updateValueAndValidity({ emitEvent: false });
    this.form.updateValueAndValidity({ emitEvent: false });
  }

  private optionalValue(value: string | null | undefined): string | null {
    const trimmed = value?.trim();
    return trimmed ? trimmed : null;
  }

  private async runAuthAction(
    action: AuthAction,
    request: () => Promise<LoginResponse>,
    authMode: AuthMode,
  ): Promise<void> {
    this.submitting.set(true);
    this.activeAction.set(action);
    this.authError.set(null);

    try {
      const response = await request();
      await this.completeAuthentication(response);
    } catch (error) {
      this.authError.set(this.resolveAuthError(error, authMode));
    } finally {
      this.submitting.set(false);
      this.activeAction.set(null);
    }
  }

  private async completeAuthentication(response: LoginResponse): Promise<void> {
    this.authSession.setTokens(response.tokens);
    this.authState.setCurrentUser(response.user);
    this.authSession.setUserSnapshot(response.user);

    await this.router.navigateByUrl('/dashboard');
  }

  private resolveAuthError(error: unknown, authMode: AuthMode): string {
    const fallback = authMode === 'register'
      ? 'Не удалось зарегистрироваться. Проверьте поля формы и попробуйте снова.'
      : 'Не удалось войти. Проверьте логин и пароль.';

    if (!(error instanceof HttpErrorResponse)) {
      return fallback;
    }

    const apiError = this.extractApiError(error);
    if (authMode === 'register') {
      if (error.status === 409 || apiError?.code === 'email_taken') {
        return 'Пользователь с таким email уже существует.';
      }
      if (error.status === 400 || apiError?.code === 'validation_error') {
        return 'Проверьте обязательные поля и длину пароля.';
      }
    }

    if (authMode === 'login' && (error.status === 401 || apiError?.code === 'invalid_credentials')) {
      return 'Не удалось войти. Проверьте логин и пароль.';
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
