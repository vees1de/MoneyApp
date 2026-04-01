import { HttpErrorResponse } from '@angular/common/http';
import { Injectable, inject, signal } from '@angular/core';
import { firstValueFrom } from 'rxjs';

import type {
  OutlookEventRecord,
  OutlookIntegrationStatus,
  OutlookManualConnectRequest,
  OutlookMessageRecord,
} from '@core/api/contracts';
import { IntegrationsApiService } from '@core/api/integrations-api.service';

@Injectable({ providedIn: 'root' })
export class CalendarSyncFacade {
  private readonly integrationsApi = inject(IntegrationsApiService);

  readonly loading = signal(false);
  readonly connecting = signal(false);
  readonly syncing = signal(false);
  readonly updatingSettings = signal(false);
  readonly sendingTestEmail = signal(false);
  readonly error = signal<string | null>(null);
  readonly notice = signal<string | null>(null);

  readonly status = signal<OutlookIntegrationStatus | null>(null);
  readonly messages = signal<OutlookMessageRecord[]>([]);
  readonly events = signal<OutlookEventRecord[]>([]);

  load(): void {
    if (this.loading()) {
      return;
    }

    void this.refresh();
  }

  async refresh(): Promise<void> {
    this.loading.set(true);
    this.error.set(null);

    try {
      const status = await firstValueFrom(this.integrationsApi.getOutlookStatus());
      this.applyStatus(status);
      if (status.connected) {
        await this.loadMailboxData();
      }
    } catch (error) {
      this.error.set(this.extractErrorMessage(error, 'Не удалось загрузить Outlook-интеграцию.'));
    } finally {
      this.loading.set(false);
    }
  }

  async getMicrosoftConnectUrl(): Promise<string | null> {
    this.connecting.set(true);
    this.error.set(null);

    try {
      const response = await firstValueFrom(this.integrationsApi.getOutlookConnectLink());
      return response.auth_url;
    } catch (error) {
      this.error.set(this.extractErrorMessage(error, 'Не удалось начать Microsoft OAuth.'));
      return null;
    } finally {
      this.connecting.set(false);
    }
  }

  async connectManual(payload: OutlookManualConnectRequest): Promise<boolean> {
    this.connecting.set(true);
    this.error.set(null);
    this.notice.set(null);

    try {
      const status = await firstValueFrom(this.integrationsApi.connectOutlookManual(payload));
      this.applyStatus(status);
      this.notice.set('Токен Microsoft сохранён. Синхронизирую почту и календарь.');
      return await this.sync();
    } catch (error) {
      this.error.set(this.extractErrorMessage(error, 'Не удалось сохранить Microsoft access token.'));
      return false;
    } finally {
      this.connecting.set(false);
    }
  }

  async sync(): Promise<boolean> {
    this.syncing.set(true);
    this.error.set(null);

    try {
      const response = await firstValueFrom(this.integrationsApi.syncOutlook());
      this.applyStatus(response.status);
      await this.loadMailboxData();
      this.notice.set(
        `Синхронизировано: ${response.messages_synced} писем и ${response.events_synced} событий.`,
      );
      return true;
    } catch (error) {
      this.error.set(this.extractErrorMessage(error, 'Не удалось синхронизировать Outlook.'));
      return false;
    } finally {
      this.syncing.set(false);
    }
  }

  async updateSystemEmailEnabled(enabled: boolean): Promise<boolean> {
    this.updatingSettings.set(true);
    this.error.set(null);

    try {
      const status = await firstValueFrom(
        this.integrationsApi.updateOutlookSettings({ system_email_enabled: enabled }),
      );
      this.applyStatus(status);
      this.notice.set(
        enabled
          ? 'Системные письма будут отправляться в привязанную Outlook-почту.'
          : 'Системные письма в Outlook отключены.',
      );
      return true;
    } catch (error) {
      this.error.set(this.extractErrorMessage(error, 'Не удалось обновить настройки Outlook.'));
      return false;
    } finally {
      this.updatingSettings.set(false);
    }
  }

  async sendTestEmail(): Promise<boolean> {
    this.sendingTestEmail.set(true);
    this.error.set(null);

    try {
      const response = await firstValueFrom(this.integrationsApi.sendOutlookTestEmail());
      this.notice.set(`Тестовое письмо отправлено на ${response.recipient}.`);
      return true;
    } catch (error) {
      this.error.set(this.extractErrorMessage(error, 'Не удалось отправить тестовое письмо.'));
      return false;
    } finally {
      this.sendingTestEmail.set(false);
    }
  }

  async disconnect(): Promise<boolean> {
    this.connecting.set(true);
    this.error.set(null);

    try {
      await firstValueFrom(this.integrationsApi.disconnectOutlook());
      this.status.set({
        graph_configured: this.status()?.graph_configured ?? false,
        connected: false,
        account: null,
      });
      this.messages.set([]);
      this.events.set([]);
      this.notice.set('Outlook отключён.');
      return true;
    } catch (error) {
      this.error.set(this.extractErrorMessage(error, 'Не удалось отключить Outlook.'));
      return false;
    } finally {
      this.connecting.set(false);
    }
  }

  setNotice(message: string | null): void {
    this.notice.set(message);
  }

  clearError(): void {
    this.error.set(null);
  }

  clearMessage(): void {
    this.notice.set(null);
  }

  private async loadMailboxData(): Promise<void> {
    const [messagesResponse, eventsResponse] = await Promise.all([
      firstValueFrom(this.integrationsApi.listOutlookMessages()),
      firstValueFrom(this.integrationsApi.listOutlookEvents()),
    ]);

    this.messages.set(messagesResponse.items ?? []);
    this.events.set(eventsResponse.items ?? []);
  }

  private applyStatus(status: OutlookIntegrationStatus): void {
    this.status.set(status);
    if (!status.connected) {
      this.messages.set([]);
      this.events.set([]);
    }
  }

  private extractErrorMessage(error: unknown, fallback: string): string {
    if (error instanceof HttpErrorResponse) {
      const apiMessage = error.error?.error?.message;
      if (typeof apiMessage === 'string' && apiMessage.trim()) {
        return apiMessage.trim();
      }
    }

    return fallback;
  }
}
