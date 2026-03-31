import { HttpErrorResponse } from '@angular/common/http';
import { CommonModule } from '@angular/common';
import { Component, OnInit, computed, inject, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { MatButtonModule } from '@angular/material/button';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { firstValueFrom } from 'rxjs';

import type {
  BoardSummary,
  BoardSummaryBoardItem,
  BoardSummaryOverdueTask,
} from '@core/api/contracts';
import { IntegrationsApiService } from '@core/api/integrations-api.service';
import { WidgetShellComponent } from '@app/widgets/widget-shell/widget-shell.component';
import type { YougileCompanyOption, YougileConnection } from '@entities/yougile';

@Component({
  selector: 'app-yougile-playground-widget',
  standalone: true,
  imports: [
    CommonModule,
    ReactiveFormsModule,
    MatButtonModule,
    MatFormFieldModule,
    MatInputModule,
    WidgetShellComponent,
  ],
  templateUrl: './yougile-playground-widget.component.html',
  styleUrl: './yougile-playground-widget.component.scss',
})
export class YougilePlaygroundWidgetComponent implements OnInit {
  private readonly fb = inject(FormBuilder);
  private readonly integrationsApi = inject(IntegrationsApiService);
  private readonly dateFormatter = new Intl.DateTimeFormat('ru-RU', {
    dateStyle: 'medium',
  });
  private readonly dateTimeFormatter = new Intl.DateTimeFormat('ru-RU', {
    dateStyle: 'short',
    timeStyle: 'short',
  });

  protected readonly connectionLoading = signal(true);
  protected readonly discoveringCompanies = signal(false);
  protected readonly connectingConnection = signal(false);
  protected readonly deletingConnection = signal(false);
  protected readonly syncingConnection = signal(false);
  protected readonly reconnectMode = signal(false);
  protected readonly hasCompanySearch = signal(false);
  protected readonly availableCompanies = signal<YougileCompanyOption[]>([]);
  protected readonly selectedCompanyId = signal<string | null>(null);
  protected readonly currentConnection = signal<YougileConnection | null>(null);
  protected readonly boardOptions = signal<BoardSummaryBoardItem[]>([]);
  protected readonly selectedBoardId = signal<string | null>(null);
  protected readonly summary = signal<BoardSummary | null>(null);
  protected readonly summaryLoading = signal(false);
  protected readonly summaryError = signal<string | null>(null);
  protected readonly actionError = signal<string | null>(null);
  protected readonly actionNotice = signal<string | null>(null);

  protected readonly credentialsForm = this.fb.nonNullable.group({
    login: ['', [Validators.required]],
    password: ['', [Validators.required]],
    name: [''],
  });

  protected readonly showWizard = computed(
    () => !this.connectionLoading() && (!this.currentConnection() || this.reconnectMode()),
  );

  protected readonly selectedCompany = computed(
    () => this.availableCompanies().find((item) => item.id === this.selectedCompanyId()) ?? null,
  );

  protected readonly activeBoard = computed(() => {
    const boardId = this.selectedBoardId();
    if (!boardId) {
      return null;
    }

    return this.boardOptions().find((item) => item.board_id === boardId) ?? null;
  });

  protected readonly summaryUnavailable = computed(() => {
    if (!this.currentConnection()) {
      return false;
    }

    const summary = this.summary();
    return (
      !this.summaryLoading() && !this.summaryError() && (!summary || summary.status !== 'ready')
    );
  });

  ngOnInit(): void {
    void this.loadCurrentConnection();
  }

  protected async discoverCompanies(): Promise<void> {
    if (this.credentialsForm.invalid || this.discoveringCompanies()) {
      this.credentialsForm.markAllAsTouched();
      return;
    }

    this.clearMessages();
    this.discoveringCompanies.set(true);
    this.hasCompanySearch.set(true);
    this.availableCompanies.set([]);
    this.selectedCompanyId.set(null);

    try {
      const payload = this.credentialsForm.getRawValue();
      const response = await firstValueFrom(
        this.integrationsApi.discoverYougileCompanies({
          login: payload.login.trim(),
          password: payload.password,
          name: payload.name.trim() || undefined,
        }),
      );

      this.availableCompanies.set(response.content);
      if (response.content.length === 1) {
        this.selectedCompanyId.set(response.content[0].id);
      }
      if (!response.content.length) {
        this.actionNotice.set('Компании не найдены. Уточните фильтр или проверьте доступ.');
      }
    } catch (error) {
      this.actionError.set(this.describeError(error));
      this.availableCompanies.set([]);
    } finally {
      this.discoveringCompanies.set(false);
    }
  }

  protected chooseCompany(companyId: string): void {
    this.selectedCompanyId.set(companyId);
    this.clearMessages();
  }

  protected async connectSelectedCompany(): Promise<void> {
    const companyId = this.selectedCompanyId();
    if (!companyId || this.connectingConnection()) {
      if (!companyId) {
        this.actionError.set('Выберите компанию YouGile.');
      }
      return;
    }

    this.clearMessages();
    this.connectingConnection.set(true);

    try {
      const payload = this.credentialsForm.getRawValue();
      const connection = await firstValueFrom(
        this.integrationsApi.connectYougileConnection({
          login: payload.login.trim(),
          password: payload.password,
          companyId,
        }),
      );

      let syncMessage = 'Подключение сохранено.';
      try {
        await this.runSync(connection.id, true);
        syncMessage = 'Подключение сохранено. Full sync запущен, summary обновится после импорта.';
      } catch (syncError) {
        this.actionError.set(this.describeError(syncError));
        syncMessage = 'Подключение сохранено. Full sync не стартовал, его можно запустить вручную.';
      }

      this.resetWizard(true);
      this.reconnectMode.set(false);
      this.selectedBoardId.set(null);
      this.actionNotice.set(syncMessage);
      await this.loadCurrentConnection(connection.id);
    } catch (error) {
      this.actionError.set(this.describeError(error));
    } finally {
      this.connectingConnection.set(false);
    }
  }

  protected beginReconnect(): void {
    this.clearMessages();
    this.resetWizard(true);
    this.reconnectMode.set(true);
  }

  protected cancelReconnect(): void {
    this.clearMessages();
    this.resetWizard(true);
    this.reconnectMode.set(false);
  }

  protected async deleteCurrentConnection(): Promise<void> {
    const connection = this.currentConnection();
    if (!connection || this.deletingConnection()) {
      return;
    }

    this.clearMessages();
    this.deletingConnection.set(true);

    try {
      await firstValueFrom(this.integrationsApi.deleteYougileConnection(connection.id));
      this.currentConnection.set(null);
      this.summary.set(null);
      this.summaryError.set(null);
      this.boardOptions.set([]);
      this.selectedBoardId.set(null);
      this.reconnectMode.set(false);
      this.resetWizard(true);
      this.actionNotice.set('Подключение удалено.');
    } catch (error) {
      this.actionError.set(this.describeError(error));
    } finally {
      this.deletingConnection.set(false);
    }
  }

  protected async syncCurrentConnection(): Promise<void> {
    const connection = this.currentConnection();
    if (!connection) {
      return;
    }

    this.clearMessages();

    try {
      await this.runSync(connection.id, false);
      this.actionNotice.set('Full sync запущен. Обновите summary через несколько секунд.');
    } catch (error) {
      this.actionError.set(this.describeError(error));
    }
  }

  protected async selectBoard(boardId: string | null): Promise<void> {
    if (this.selectedBoardId() === boardId) {
      return;
    }

    const connection = this.currentConnection();
    if (!connection) {
      return;
    }

    this.selectedBoardId.set(boardId);
    await this.loadSummary(connection.id, boardId, boardId !== null);
  }

  protected async reloadSummary(): Promise<void> {
    const connection = this.currentConnection();
    if (!connection) {
      return;
    }

    await this.loadSummary(connection.id, this.selectedBoardId(), this.selectedBoardId() !== null);
  }

  protected connectionLabel(connection: YougileConnection): string {
    return connection.title?.trim() || `Компания ${connection.yougile_company_id}`;
  }

  protected connectionStatusLabel(status: string): string {
    switch (status) {
      case 'active':
        return 'Активно';
      case 'invalid':
        return 'Нужна проверка ключа';
      case 'sync_error':
        return 'Ошибка sync';
      case 'revoked':
        return 'Удалено';
      default:
        return status;
    }
  }

  protected formatDeadline(value?: string | null): string {
    if (!value) {
      return 'Без дедлайна';
    }

    const timestamp = new Date(value);
    if (Number.isNaN(timestamp.getTime())) {
      return value;
    }

    return this.dateFormatter.format(timestamp);
  }

  protected formatSyncTime(value?: string | null): string {
    if (!value) {
      return 'Не было';
    }

    const timestamp = new Date(value);
    if (Number.isNaN(timestamp.getTime())) {
      return value;
    }

    return this.dateTimeFormatter.format(timestamp);
  }

  protected boardLabel(item: BoardSummaryBoardItem | BoardSummaryOverdueTask): string {
    if ('board_title' in item) {
      return item.board_title?.trim() || 'Без доски';
    }

    return item.title;
  }

  private async loadCurrentConnection(_preferredConnectionId?: string): Promise<void> {
    this.connectionLoading.set(true);
    this.summaryError.set(null);

    try {
      const response = await firstValueFrom(this.integrationsApi.listYougileConnections());
      const connection = response.items.find((item) => item.status !== 'revoked') ?? null;
      this.currentConnection.set(connection);

      if (!connection) {
        this.summary.set(null);
        this.boardOptions.set([]);
        this.selectedBoardId.set(null);
        return;
      }

      await this.loadSummary(
        connection.id,
        this.selectedBoardId(),
        this.selectedBoardId() !== null,
      );
    } catch (error) {
      this.currentConnection.set(null);
      this.summary.set(null);
      this.boardOptions.set([]);
      this.selectedBoardId.set(null);
      this.actionError.set(this.describeError(error));
    } finally {
      this.connectionLoading.set(false);
    }
  }

  private async loadSummary(
    connectionId: string,
    boardId: string | null,
    preserveBoardOptions: boolean,
  ): Promise<void> {
    this.summaryLoading.set(true);
    this.summaryError.set(null);

    try {
      const summary = await firstValueFrom(
        this.integrationsApi.getAgileBoardSummary({
          connection_id: connectionId,
          board_id: boardId ?? undefined,
        }),
      );

      this.summary.set(summary);
      if (!preserveBoardOptions || boardId === null) {
        this.boardOptions.set(summary.boards);
      }
    } catch (error) {
      this.summaryError.set(this.describeError(error));
      this.summary.set(null);
      if (!preserveBoardOptions) {
        this.boardOptions.set([]);
      }
    } finally {
      this.summaryLoading.set(false);
    }
  }

  private async runSync(connectionId: string, preserveNotice: boolean): Promise<void> {
    this.syncingConnection.set(true);

    try {
      await firstValueFrom(
        this.integrationsApi.startYougileSync(connectionId, {
          mode: 'full',
          includeUsers: true,
          includeStructure: true,
          includeTasks: true,
        }),
      );
    } finally {
      this.syncingConnection.set(false);
      if (!preserveNotice && this.actionNotice()) {
        this.actionNotice.set(null);
      }
    }
  }

  private resetWizard(resetCredentials: boolean): void {
    this.hasCompanySearch.set(false);
    this.availableCompanies.set([]);
    this.selectedCompanyId.set(null);

    if (resetCredentials) {
      this.credentialsForm.reset({
        login: '',
        password: '',
        name: '',
      });
    }
  }

  private clearMessages(): void {
    this.actionError.set(null);
    this.actionNotice.set(null);
  }

  private describeError(error: unknown): string {
    if (error instanceof HttpErrorResponse) {
      if (error.status === 401) {
        return 'Неверный login или password.';
      }
      if (error.status === 403) {
        return 'Доступ к компании запрещён или у пользователя нет нужных прав.';
      }
      if (error.status === 429) {
        return 'YouGile временно ограничил запросы. Повторите позже.';
      }

      const apiMessage = this.extractApiMessage(error.error);
      if (apiMessage) {
        return apiMessage;
      }

      if (error.status === 0) {
        return 'Backend недоступен.';
      }
    }

    return 'Операция завершилась ошибкой.';
  }

  private extractApiMessage(payload: unknown): string | null {
    if (typeof payload === 'string' && payload.trim()) {
      return payload;
    }

    if (!payload || typeof payload !== 'object') {
      return null;
    }

    const errorPayload = (payload as { error?: unknown }).error;
    if (!errorPayload || typeof errorPayload !== 'object') {
      return null;
    }

    const message = (errorPayload as { message?: unknown }).message;
    return typeof message === 'string' && message.trim() ? message : null;
  }
}
