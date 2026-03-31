import { HttpErrorResponse } from '@angular/common/http';
import { CommonModule } from '@angular/common';
import { Component, computed, inject, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { MatButtonModule } from '@angular/material/button';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { firstValueFrom, forkJoin } from 'rxjs';

import { WidgetShellComponent } from '@app/widgets/widget-shell/widget-shell.component';
import { IntegrationsApiService } from '@core/api/integrations-api.service';
import { AuthStateService } from '@core/auth/auth-state.service';
import { PERMISSIONS } from '@core/auth/permissions';
import type {
  YougileBoard,
  YougileColumn,
  YougileConnection,
  YougileImportedUser,
  YougileMappingView,
  YougileProject,
  YougileSyncJob,
} from '@entities/yougile';

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
export class YougilePlaygroundWidgetComponent {
  private readonly fb = inject(FormBuilder);
  private readonly integrationsApi = inject(IntegrationsApiService);
  private readonly authState = inject(AuthStateService);
  private readonly dateTimeFormatter = new Intl.DateTimeFormat('ru-RU', {
    dateStyle: 'short',
    timeStyle: 'short',
  });

  private syncPollTimer: ReturnType<typeof setTimeout> | null = null;

  protected readonly canManageSettings = computed(() =>
    this.authState.hasPermission(PERMISSIONS.settingsManage),
  );
  protected readonly loadingConnections = signal(true);
  protected readonly overviewLoading = signal(false);
  protected readonly testingKey = signal(false);
  protected readonly creatingConnection = signal(false);
  protected readonly importingUsers = signal(false);
  protected readonly importingStructure = signal(false);
  protected readonly autoMatching = signal(false);
  protected readonly syncing = signal(false);
  protected readonly syncJobLoading = signal(false);
  protected readonly loadError = signal<string | null>(null);
  protected readonly actionError = signal<string | null>(null);
  protected readonly actionNotice = signal<string | null>(null);
  protected readonly connections = signal<YougileConnection[]>([]);
  protected readonly users = signal<YougileImportedUser[]>([]);
  protected readonly projects = signal<YougileProject[]>([]);
  protected readonly boards = signal<YougileBoard[]>([]);
  protected readonly columns = signal<YougileColumn[]>([]);
  protected readonly mappings = signal<YougileMappingView[]>([]);
  protected readonly selectedConnectionId = signal<string | null>(null);
  protected readonly syncJob = signal<YougileSyncJob | null>(null);

  protected readonly selectedConnection = computed(() =>
    this.connections().find((item) => item.id === this.selectedConnectionId()) ?? null,
  );
  protected readonly usersPreview = computed(() => this.users().slice(0, 5));
  protected readonly boardsPreview = computed(() => this.boards().slice(0, 5));
  protected readonly mappingsPreview = computed(() => this.mappings().slice(0, 5));
  protected readonly selectedConnectionRevoked = computed(
    () => this.selectedConnection()?.status === 'revoked',
  );

  protected readonly connectionForm = this.fb.nonNullable.group({
    title: [''],
    apiBaseUrl: ['https://yougile.com'],
    yougileCompanyId: ['', [Validators.required]],
    apiKey: ['', [Validators.required]],
  });

  constructor() {
    if (this.canManageSettings()) {
      void this.loadConnections();
      return;
    }

    this.loadingConnections.set(false);
  }

  ngOnDestroy(): void {
    this.stopSyncPolling();
  }

  protected async refreshConnections(): Promise<void> {
    await this.loadConnections(this.selectedConnectionId(), true);
  }

  protected async testKey(): Promise<void> {
    if (!this.connectionForm.controls.apiKey.value.trim()) {
      this.connectionForm.controls.apiKey.markAsTouched();
      return;
    }

    this.clearMessages();
    this.testingKey.set(true);

    try {
      const payload = this.connectionForm.getRawValue();
      const response = await firstValueFrom(
        this.integrationsApi.testYougileKey({
          apiBaseUrl: payload.apiBaseUrl.trim(),
          apiKey: payload.apiKey.trim(),
        }),
      );

      this.actionNotice.set(
        `${response.message}. Лимит: ${response.rateLimitMode}. Доступ к компании: ${
          response.companyAccessible ? 'да' : 'нет'
        }.`,
      );
    } catch (error) {
      this.actionError.set(this.describeError(error));
    } finally {
      this.testingKey.set(false);
    }
  }

  protected async createConnection(): Promise<void> {
    if (this.connectionForm.invalid || this.creatingConnection()) {
      this.connectionForm.markAllAsTouched();
      return;
    }

    this.clearMessages();
    this.creatingConnection.set(true);

    try {
      const payload = this.connectionForm.getRawValue();
      const created = await firstValueFrom(
        this.integrationsApi.createYougileConnection({
          title: payload.title.trim() || null,
          apiBaseUrl: payload.apiBaseUrl.trim(),
          authMode: 'api_key',
          yougileCompanyId: payload.yougileCompanyId.trim(),
          apiKey: payload.apiKey.trim(),
        }),
      );

      this.actionNotice.set('Подключение создано. Можно запускать импорт и sync.');
      this.connectionForm.patchValue({
        title: '',
        yougileCompanyId: '',
        apiKey: '',
      });
      this.connectionForm.markAsPristine();
      this.connectionForm.markAsUntouched();
      await this.loadConnections(created.id, true);
    } catch (error) {
      this.actionError.set(this.describeError(error));
    } finally {
      this.creatingConnection.set(false);
    }
  }

  protected selectConnection(id: string): void {
    if (this.selectedConnectionId() === id) {
      return;
    }

    this.selectedConnectionId.set(id);
    this.syncJob.set(null);
    this.stopSyncPolling();
    this.clearMessages();
    void this.refreshSelectedConnectionData(id);
  }

  protected async importUsers(): Promise<void> {
    const connection = this.selectedConnection();
    if (!connection || this.importingUsers()) {
      return;
    }

    this.clearMessages();
    this.importingUsers.set(true);

    try {
      const response = await firstValueFrom(this.integrationsApi.importYougileUsers(connection.id));
      this.actionNotice.set(
        `Пользователи импортированы: +${response.imported}, обновлено ${response.updated}, ошибок ${response.failed}.`,
      );
      await this.loadConnections(connection.id, true);
    } catch (error) {
      this.actionError.set(this.describeError(error));
    } finally {
      this.importingUsers.set(false);
    }
  }

  protected async importStructure(): Promise<void> {
    const connection = this.selectedConnection();
    if (!connection || this.importingStructure()) {
      return;
    }

    this.clearMessages();
    this.importingStructure.set(true);

    try {
      const response = await firstValueFrom(
        this.integrationsApi.importYougileStructure(connection.id),
      );
      this.actionNotice.set(
        `Структура обновлена: проектов ${response.projectsImported}, бордов ${response.boardsImported}, колонок ${response.columnsImported}.`,
      );
      await this.loadConnections(connection.id, true);
    } catch (error) {
      this.actionError.set(this.describeError(error));
    } finally {
      this.importingStructure.set(false);
    }
  }

  protected async autoMatch(): Promise<void> {
    const connection = this.selectedConnection();
    if (!connection || this.autoMatching()) {
      return;
    }

    this.clearMessages();
    this.autoMatching.set(true);

    try {
      const response = await firstValueFrom(
        this.integrationsApi.autoMatchYougileMappings(connection.id, { strategy: 'email' }),
      );
      this.actionNotice.set(
        `Автосопоставление завершено: matched ${response.matched}, без пары сотрудников ${response.unmatchedInternal}, без пары в YouGile ${response.unmatchedYougile}.`,
      );
      await this.refreshSelectedConnectionData(connection.id);
    } catch (error) {
      this.actionError.set(this.describeError(error));
    } finally {
      this.autoMatching.set(false);
    }
  }

  protected async startFullSync(): Promise<void> {
    const connection = this.selectedConnection();
    if (!connection || this.syncing()) {
      return;
    }

    this.clearMessages();
    this.syncing.set(true);

    try {
      const response = await firstValueFrom(
        this.integrationsApi.startYougileSync(connection.id, {
          mode: 'full',
          includeUsers: true,
          includeStructure: true,
          includeTasks: true,
        }),
      );

      this.actionNotice.set(`Full sync запущен. Job: ${response.jobId}.`);
      await this.loadSyncJob(response.jobId, true);
    } catch (error) {
      this.actionError.set(this.describeError(error));
    } finally {
      this.syncing.set(false);
    }
  }

  protected async refreshSyncJob(jobId = this.syncJob()?.id ?? null): Promise<void> {
    if (!jobId || this.syncJobLoading()) {
      return;
    }

    await this.loadSyncJob(jobId, false);
  }

  protected formatDateTime(value?: string | null): string {
    if (!value) {
      return 'Не было';
    }

    const timestamp = new Date(value);
    if (Number.isNaN(timestamp.getTime())) {
      return value;
    }

    return this.dateTimeFormatter.format(timestamp);
  }

  protected connectionLabel(connection: YougileConnection): string {
    return connection.title?.trim() || `Компания ${connection.yougile_company_id}`;
  }

  protected connectionStatusLabel(status: string): string {
    switch (status) {
      case 'active':
        return 'Активно';
      case 'sync_error':
        return 'Sync error';
      case 'invalid':
        return 'Невалидно';
      case 'revoked':
        return 'Отозвано';
      default:
        return status;
    }
  }

  protected syncStatusLabel(status: string): string {
    switch (status) {
      case 'pending':
        return 'В очереди';
      case 'processing':
        return 'В работе';
      case 'done':
        return 'Завершено';
      case 'failed':
        return 'Ошибка';
      default:
        return status;
    }
  }

  private async loadConnections(
    preferredConnectionId = this.selectedConnectionId(),
    refreshSelection = true,
  ): Promise<void> {
    if (!this.canManageSettings()) {
      this.loadingConnections.set(false);
      return;
    }

    this.loadingConnections.set(true);
    this.loadError.set(null);

    try {
      const response = await firstValueFrom(this.integrationsApi.listYougileConnections());
      const connections = response.items
        .slice()
        .sort((left, right) => Date.parse(right.updated_at) - Date.parse(left.updated_at));

      this.connections.set(connections);

      const selectedId = this.resolveSelectedConnectionId(connections, preferredConnectionId);
      this.selectedConnectionId.set(selectedId);

      if (!refreshSelection) {
        return;
      }

      if (selectedId) {
        await this.refreshSelectedConnectionData(selectedId);
      } else {
        this.resetSelectedCollections();
      }
    } catch (error) {
      this.loadError.set(this.describeError(error));
      this.resetSelectedCollections();
    } finally {
      this.loadingConnections.set(false);
    }
  }

  private async refreshSelectedConnectionData(connectionId = this.selectedConnectionId()): Promise<void> {
    if (!connectionId) {
      this.resetSelectedCollections();
      return;
    }

    this.overviewLoading.set(true);

    try {
      const response = await firstValueFrom(
        forkJoin({
          users: this.integrationsApi.listYougileUsers(connectionId),
          projects: this.integrationsApi.listYougileProjects(connectionId),
          boards: this.integrationsApi.listYougileBoards(connectionId),
          columns: this.integrationsApi.listYougileColumns(connectionId),
          mappings: this.integrationsApi.listYougileMappings(connectionId),
        }),
      );

      this.users.set(response.users.items);
      this.projects.set(response.projects.items);
      this.boards.set(response.boards.items);
      this.columns.set(response.columns.items);
      this.mappings.set(response.mappings.content);
    } catch (error) {
      this.actionError.set(this.describeError(error));
      this.resetSelectedCollections();
    } finally {
      this.overviewLoading.set(false);
    }
  }

  private async loadSyncJob(jobId: string, keepPolling: boolean): Promise<void> {
    this.syncJobLoading.set(true);

    try {
      const job = await firstValueFrom(this.integrationsApi.getYougileSyncJob(jobId));
      this.syncJob.set(job);

      if (keepPolling && this.isActiveSyncStatus(job.status)) {
        this.scheduleSyncPoll(jobId);
        return;
      }

      this.stopSyncPolling();

      if (!this.isActiveSyncStatus(job.status)) {
        await this.loadConnections(job.connection_id, true);
      }
    } catch (error) {
      this.actionError.set(this.describeError(error));
      this.stopSyncPolling();
    } finally {
      this.syncJobLoading.set(false);
    }
  }

  private resolveSelectedConnectionId(
    connections: YougileConnection[],
    preferredConnectionId: string | null,
  ): string | null {
    if (preferredConnectionId && connections.some((item) => item.id === preferredConnectionId)) {
      return preferredConnectionId;
    }

    return connections[0]?.id ?? null;
  }

  private resetSelectedCollections(): void {
    this.users.set([]);
    this.projects.set([]);
    this.boards.set([]);
    this.columns.set([]);
    this.mappings.set([]);
    this.syncJob.set(null);
    this.stopSyncPolling();
  }

  private scheduleSyncPoll(jobId: string): void {
    this.stopSyncPolling();
    this.syncPollTimer = setTimeout(() => {
      void this.loadSyncJob(jobId, true);
    }, 2500);
  }

  private stopSyncPolling(): void {
    if (this.syncPollTimer === null) {
      return;
    }

    clearTimeout(this.syncPollTimer);
    this.syncPollTimer = null;
  }

  private isActiveSyncStatus(status: string): boolean {
    return status === 'pending' || status === 'processing';
  }

  private clearMessages(): void {
    this.actionError.set(null);
    this.actionNotice.set(null);
  }

  private describeError(error: unknown): string {
    if (error instanceof HttpErrorResponse) {
      const apiMessage = this.extractApiMessage(error.error);

      if (error.status === 403) {
        return apiMessage ?? 'Нужен permission settings.manage.';
      }
      if (error.status === 0) {
        return 'Backend недоступен. Проверьте API и CORS.';
      }
      if (apiMessage) {
        return apiMessage;
      }
      if (error.message) {
        return error.message;
      }
    }

    return 'Операция завершилась ошибкой.';
  }

  private extractApiMessage(payload: unknown): string | null {
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
