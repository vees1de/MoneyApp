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
import type { YougileConnection } from '@entities/yougile';

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

  protected readonly connectionsLoading = signal(true);
  protected readonly summaryLoading = signal(false);
  protected readonly creatingConnection = signal(false);
  protected readonly deletingConnectionId = signal<string | null>(null);
  protected readonly syncingConnectionId = signal<string | null>(null);
  protected readonly actionError = signal<string | null>(null);
  protected readonly actionNotice = signal<string | null>(null);
  protected readonly connections = signal<YougileConnection[]>([]);
  protected readonly selectedConnectionId = signal<string | null>(null);
  protected readonly boardOptions = signal<BoardSummaryBoardItem[]>([]);
  protected readonly summary = signal<BoardSummary | null>(null);
  protected readonly summaryError = signal<string | null>(null);
  protected readonly selectedBoardId = signal<string | null>(null);

  protected readonly connectForm = this.fb.nonNullable.group({
    title: [''],
    login: ['', [Validators.required]],
    password: ['', [Validators.required]],
    companyId: ['', [Validators.required]],
  });

  protected readonly selectedConnection = computed(() =>
    this.connections().find((item) => item.id === this.selectedConnectionId()) ?? null,
  );

  protected readonly activeBoard = computed(() => {
    const boardId = this.selectedBoardId();
    if (!boardId) {
      return null;
    }

    return this.boardOptions().find((item) => item.board_id === boardId) ?? null;
  });

  protected readonly summaryUnavailable = computed(() => {
    const summary = this.summary();
    return !this.summaryLoading() && !this.summaryError() && (!summary || summary.status !== 'ready');
  });

  ngOnInit(): void {
    void this.loadConnections();
  }

  protected async createConnection(): Promise<void> {
    if (this.connectForm.invalid || this.creatingConnection()) {
      this.connectForm.markAllAsTouched();
      return;
    }

    this.clearMessages();
    this.creatingConnection.set(true);

    try {
      const payload = this.connectForm.getRawValue();
      const key = await firstValueFrom(
        this.integrationsApi.createYougileKey({
          login: payload.login.trim(),
          password: payload.password,
          companyId: payload.companyId.trim(),
        }),
      );

      const connection = await firstValueFrom(
        this.integrationsApi.createYougileConnection({
          title: payload.title.trim() || null,
          apiBaseUrl: 'https://yougile.com',
          authMode: 'api_key',
          yougileCompanyId: key.companyId,
          apiKey: key.apiKey,
        }),
      );

      await this.startSync(connection.id, true);

      this.actionNotice.set(
        'Подключение создано. Full sync запущен, после завершения обновите summary.',
      );
      this.connectForm.reset({
        title: '',
        login: '',
        password: '',
        companyId: '',
      });
      await this.loadConnections(connection.id);
    } catch (error) {
      this.actionError.set(this.describeError(error));
    } finally {
      this.creatingConnection.set(false);
    }
  }

  protected async reloadConnections(): Promise<void> {
    await this.loadConnections(this.selectedConnectionId());
  }

  protected async selectConnection(connectionId: string): Promise<void> {
    if (this.selectedConnectionId() === connectionId) {
      return;
    }

    this.selectedConnectionId.set(connectionId);
    this.selectedBoardId.set(null);
    await this.loadSummary(connectionId, null, false);
  }

  protected async selectBoard(boardId: string | null): Promise<void> {
    if (this.selectedBoardId() === boardId) {
      return;
    }

    const connectionId = this.selectedConnectionId();
    if (!connectionId) {
      return;
    }

    this.selectedBoardId.set(boardId);
    await this.loadSummary(connectionId, boardId, boardId !== null);
  }

  protected async reloadSummary(): Promise<void> {
    const connectionId = this.selectedConnectionId();
    if (!connectionId) {
      return;
    }

    await this.loadSummary(connectionId, this.selectedBoardId(), this.selectedBoardId() !== null);
  }

  protected async syncSelectedConnection(): Promise<void> {
    const connectionId = this.selectedConnectionId();
    if (!connectionId) {
      return;
    }

    this.clearMessages();

    try {
      await this.startSync(connectionId, false);
      this.actionNotice.set('Full sync запущен. Обновите summary чуть позже.');
    } catch (error) {
      this.actionError.set(this.describeError(error));
    }
  }

  protected async deleteSelectedConnection(): Promise<void> {
    const connection = this.selectedConnection();
    if (!connection || this.deletingConnectionId()) {
      return;
    }

    this.clearMessages();
    this.deletingConnectionId.set(connection.id);

    try {
      await firstValueFrom(this.integrationsApi.deleteYougileConnection(connection.id));
      this.actionNotice.set('Подключение удалено.');
      await this.loadConnections(null);
    } catch (error) {
      this.actionError.set(this.describeError(error));
    } finally {
      this.deletingConnectionId.set(null);
    }
  }

  protected connectionLabel(connection: YougileConnection): string {
    return connection.title?.trim() || `Компания ${connection.yougile_company_id}`;
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

  private async loadConnections(preferredConnectionId = this.selectedConnectionId()): Promise<void> {
    this.connectionsLoading.set(true);
    this.summaryError.set(null);

    try {
      const response = await firstValueFrom(this.integrationsApi.listYougileConnections());
      const visibleConnections = response.items
        .filter((item) => item.status !== 'revoked')
        .sort((left, right) => Date.parse(right.updated_at) - Date.parse(left.updated_at));

      this.connections.set(visibleConnections);

      const nextSelectedId = this.resolveSelectedConnectionId(
        visibleConnections,
        preferredConnectionId,
      );

      this.selectedConnectionId.set(nextSelectedId);
      this.selectedBoardId.set(null);

      if (nextSelectedId) {
        await this.loadSummary(nextSelectedId, null, false);
      } else {
        this.summary.set(null);
        this.boardOptions.set([]);
      }
    } catch (error) {
      this.actionError.set(this.describeError(error));
      this.connections.set([]);
      this.selectedConnectionId.set(null);
      this.summary.set(null);
      this.boardOptions.set([]);
    } finally {
      this.connectionsLoading.set(false);
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

  private async startSync(connectionId: string, preserveNotice: boolean): Promise<void> {
    this.syncingConnectionId.set(connectionId);

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
      this.syncingConnectionId.set(null);
      if (!preserveNotice && this.actionNotice()) {
        this.actionNotice.set(null);
      }
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

  private clearMessages(): void {
    this.actionError.set(null);
    this.actionNotice.set(null);
  }

  private describeError(error: unknown): string {
    if (error instanceof HttpErrorResponse) {
      const apiMessage = this.extractApiMessage(error.error);
      if (apiMessage) {
        return apiMessage;
      }
      if (error.status === 0) {
        return 'Backend недоступен.';
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
