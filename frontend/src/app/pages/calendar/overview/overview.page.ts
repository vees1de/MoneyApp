import { CommonModule } from '@angular/common';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { Component, DestroyRef, OnInit, computed, inject } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { ActivatedRoute, Router, RouterLink } from '@angular/router';
import { MatButtonModule } from '@angular/material/button';
import { MatCardModule } from '@angular/material/card';
import { MatDividerModule } from '@angular/material/divider';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatIconModule } from '@angular/material/icon';
import { MatInputModule } from '@angular/material/input';
import { MatProgressBarModule } from '@angular/material/progress-bar';
import { MatSlideToggleModule } from '@angular/material/slide-toggle';

import type { OutlookEventRecord, OutlookMessageRecord } from '@core/api/contracts';
import { CalendarSyncFacade } from '@features/calendar-sync';

@Component({
  selector: 'app-page-calendar-overview',
  standalone: true,
  imports: [
    CommonModule,
    ReactiveFormsModule,
    RouterLink,
    MatButtonModule,
    MatCardModule,
    MatDividerModule,
    MatFormFieldModule,
    MatIconModule,
    MatInputModule,
    MatProgressBarModule,
    MatSlideToggleModule,
  ],
  templateUrl: './overview.page.html',
  styleUrl: './overview.page.scss',
})
export class CalendarOverviewPageComponent implements OnInit {
  private readonly facade = inject(CalendarSyncFacade);
  private readonly fb = inject(FormBuilder);
  private readonly route = inject(ActivatedRoute);
  private readonly router = inject(Router);
  private readonly destroyRef = inject(DestroyRef);

  protected readonly routePath = '/calendar/overview';
  protected readonly loading = this.facade.loading;
  protected readonly connecting = this.facade.connecting;
  protected readonly syncing = this.facade.syncing;
  protected readonly updatingSettings = this.facade.updatingSettings;
  protected readonly sendingTestEmail = this.facade.sendingTestEmail;
  protected readonly error = this.facade.error;
  protected readonly notice = this.facade.notice;
  protected readonly status = this.facade.status;
  protected readonly messages = this.facade.messages;
  protected readonly events = this.facade.events;

  protected readonly account = computed(() => this.status()?.account ?? null);
  protected readonly connected = computed(() => this.status()?.connected ?? false);
  protected readonly graphConfigured = computed(() => this.status()?.graph_configured ?? false);
  protected readonly anyBusy = computed(
    () =>
      this.loading() ||
      this.connecting() ||
      this.syncing() ||
      this.updatingSettings() ||
      this.sendingTestEmail(),
  );

  protected readonly manualForm = this.fb.nonNullable.group({
    access_token: ['', [Validators.required]],
    refresh_token: [''],
    system_email_enabled: [true],
  });

  ngOnInit(): void {
    this.facade.load();

    this.route.queryParamMap.pipe(takeUntilDestroyed(this.destroyRef)).subscribe((params) => {
      const outlookStatus = params.get('outlook');
      const outlookMessage = params.get('outlook_message');
      if (!outlookStatus && !outlookMessage) {
        return;
      }

      if (outlookMessage) {
        this.facade.setNotice(outlookMessage);
      }

      if (outlookStatus === 'connected') {
        void this.facade.sync();
      }

      void this.router.navigate([], {
        relativeTo: this.route,
        queryParams: {},
        replaceUrl: true,
      });
    });
  }

  protected async startMicrosoftConnect(): Promise<void> {
    const authUrl = await this.facade.getMicrosoftConnectUrl();
    if (authUrl) {
      globalThis.location.assign(authUrl);
    }
  }

  protected async connectWithToken(): Promise<void> {
    if (this.manualForm.invalid) {
      this.manualForm.markAllAsTouched();
      return;
    }

    const values = this.manualForm.getRawValue();
    const connected = await this.facade.connectManual({
      access_token: values.access_token.trim(),
      refresh_token: values.refresh_token.trim() || null,
      system_email_enabled: values.system_email_enabled,
    });

    if (connected) {
      this.manualForm.patchValue({
        access_token: '',
        refresh_token: '',
      });
    }
  }

  protected async runSync(): Promise<void> {
    await this.facade.sync();
  }

  protected async sendTestEmail(): Promise<void> {
    await this.facade.sendTestEmail();
  }

  protected async disconnectOutlook(): Promise<void> {
    await this.facade.disconnect();
  }

  protected async updateSystemEmailEnabled(enabled: boolean): Promise<void> {
    const ok = await this.facade.updateSystemEmailEnabled(enabled);
    if (!ok) {
      await this.facade.refresh();
    }
  }

  protected clearError(): void {
    this.facade.clearError();
  }

  protected clearNotice(): void {
    this.facade.clearMessage();
  }

  protected authModeLabel(mode: string | null | undefined): string {
    switch ((mode ?? '').toLowerCase()) {
      case 'oauth':
        return 'Microsoft account';
      case 'access_token':
        return 'Graph access token';
      default:
        return mode || '—';
    }
  }

  protected connectionStatusLabel(): string {
    if (this.connected()) {
      return 'Подключено';
    }
    if (this.account()) {
      return 'Требует внимания';
    }
    return 'Не подключено';
  }

  protected formatDateTime(value: string | null | undefined): string {
    if (!value) {
      return '—';
    }

    const parsed = new Date(value);
    if (Number.isNaN(parsed.getTime())) {
      return value;
    }

    return parsed.toLocaleString('ru-RU', {
      day: '2-digit',
      month: 'short',
      hour: '2-digit',
      minute: '2-digit',
    });
  }

  protected messageSenderLabel(item: OutlookMessageRecord): string {
    return item.sender_name || item.sender_email || 'Без отправителя';
  }

  protected eventMetaLabel(item: OutlookEventRecord): string {
    return [item.organizer_name || item.organizer_email, item.location]
      .filter((value) => !!value)
      .join(' · ');
  }
}
