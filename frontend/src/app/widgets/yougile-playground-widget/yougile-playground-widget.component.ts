import { HttpErrorResponse } from '@angular/common/http';
import { CommonModule } from '@angular/common';
import { Component, OnInit, computed, inject, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { MatButtonModule } from '@angular/material/button';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { firstValueFrom } from 'rxjs';

import { IntegrationsApiService } from '@core/api/integrations-api.service';
import { WidgetShellComponent } from '@app/widgets/widget-shell/widget-shell.component';
import type {
  YougileBoard,
  YougileCompanyOption,
  YougileConnection,
  YougileTask,
} from '@entities/yougile';

type YougileColumnCard = {
  columnId: string;
  title: string;
  boardTitle: string;
  tasksTotal: number;
};

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
  protected readonly tasksLoading = signal(false);
  protected readonly syncingConnection = signal(false);
  protected readonly deletingConnection = signal(false);
  protected readonly modalOpen = signal(false);
  protected readonly tasksModalOpen = signal(false);
  protected readonly reconnectMode = signal(false);
  protected readonly discoveringCompanies = signal(false);
  protected readonly connectingConnection = signal(false);
  protected readonly hasCompanySearch = signal(false);
  protected readonly availableCompanies = signal<YougileCompanyOption[]>([]);
  protected readonly selectedCompanyId = signal<string | null>(null);
  protected readonly currentConnection = signal<YougileConnection | null>(null);
  protected readonly boards = signal<YougileBoard[]>([]);
  protected readonly tasks = signal<YougileTask[]>([]);
  protected readonly selectedBoardId = signal<string | null>(null);
  protected readonly selectedColumnId = signal<string | null>(null);
  protected readonly tasksError = signal<string | null>(null);
  protected readonly actionError = signal<string | null>(null);
  protected readonly actionNotice = signal<string | null>(null);
  protected readonly modalError = signal<string | null>(null);

  protected readonly credentialsForm = this.fb.nonNullable.group({
    login: ['', [Validators.required]],
    password: ['', [Validators.required]],
    name: [''],
  });

  protected readonly selectedCompany = computed(
    () => this.availableCompanies().find((item) => item.id === this.selectedCompanyId()) ?? null,
  );

  protected readonly activeTasks = computed(() =>
    this.tasks().filter((item) => !item.deleted && !item.archived && !item.completed),
  );

  protected readonly filteredTasks = computed(() => {
    const selectedBoardId = this.selectedBoardId();
    const activeTasks = this.activeTasks();

    return selectedBoardId ? activeTasks.filter((item) => item.boardId === selectedBoardId) : activeTasks;
  });

  protected readonly visibleTasks = computed(() =>
    [
      ...(this.selectedColumnId()
        ? this.filteredTasks().filter((item) => item.columnId === this.selectedColumnId())
        : this.filteredTasks()),
    ].sort((left, right) => this.taskSortValue(left) - this.taskSortValue(right)),
  );

  protected readonly columnCards = computed((): YougileColumnCard[] => {
    const taskColumns = new Map<string, YougileColumnCard>();

    for (const task of this.filteredTasks()) {
      const columnId = task.columnId?.trim();
      if (!columnId) {
        continue;
      }

      const existing = taskColumns.get(columnId);
      if (existing) {
        existing.tasksTotal += 1;
        continue;
      }

      taskColumns.set(columnId, {
        columnId,
        title: task.columnTitle?.trim() || 'Без колонки',
        boardTitle: this.resolveBoardLabel(task.boardId, task.boardTitle),
        tasksTotal: 1,
      });
    }

    return [...taskColumns.values()].slice(0, 4);
  });

  protected readonly selectedColumn = computed(() => {
    const columnId = this.selectedColumnId();
    if (!columnId) {
      return null;
    }

    return this.columnCards().find((item) => item.columnId === columnId) ?? null;
  });

  ngOnInit(): void {
    void this.loadCurrentConnection();
  }

  protected openSettingsModal(): void {
    this.clearMessages();
    this.modalError.set(null);
    this.reconnectMode.set(false);
    this.modalOpen.set(true);
    this.resetModalState(true);
  }

  protected openConnectModal(reconnect = false): void {
    this.clearMessages();
    this.modalError.set(null);
    this.reconnectMode.set(reconnect);
    this.modalOpen.set(true);
    this.resetModalState(true);
  }

  protected switchToReconnect(): void {
    this.modalError.set(null);
    this.reconnectMode.set(true);
    this.resetModalState(true);
  }

  protected backToSettings(): void {
    if (!this.currentConnection()) {
      this.closeConnectModal();
      return;
    }

    this.modalError.set(null);
    this.reconnectMode.set(false);
    this.resetModalState(true);
  }

  protected closeConnectModal(force = false): void {
    if (!force && (this.discoveringCompanies() || this.connectingConnection())) {
      return;
    }

    this.modalOpen.set(false);
    this.reconnectMode.set(false);
    this.modalError.set(null);
    this.resetModalState(true);
  }

  protected openTasksModal(columnId: string | null = null): void {
    this.selectedColumnId.set(columnId);
    this.tasksModalOpen.set(true);
  }

  protected closeTasksModal(): void {
    this.tasksModalOpen.set(false);
    this.selectedColumnId.set(null);
  }

  protected async discoverCompanies(): Promise<void> {
    if (this.credentialsForm.invalid || this.discoveringCompanies()) {
      this.credentialsForm.markAllAsTouched();
      return;
    }

    this.modalError.set(null);
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
    } catch (error) {
      this.modalError.set(this.describeError(error));
      this.availableCompanies.set([]);
    } finally {
      this.discoveringCompanies.set(false);
    }
  }

  protected chooseCompany(companyId: string): void {
    this.selectedCompanyId.set(companyId);
    this.modalError.set(null);
  }

  protected async connectSelectedCompany(): Promise<void> {
    const company = this.selectedCompany();
    if (!company || this.connectingConnection()) {
      this.modalError.set('Выберите компанию.');
      return;
    }

    this.modalError.set(null);
    this.clearMessages();
    this.connectingConnection.set(true);

    try {
      const payload = this.credentialsForm.getRawValue();
      const connection = await firstValueFrom(
        this.integrationsApi.connectYougileConnection({
          login: payload.login.trim(),
          password: payload.password,
          companyId: company.id,
        }),
      );

      let notice = 'Подключение сохранено.';
      try {
        await this.runSync(connection.id, true);
        notice = 'Подключение сохранено. Full sync запущен.';
      } catch (syncError) {
        this.actionError.set(this.describeError(syncError));
        notice = 'Подключение сохранено. Sync можно запустить вручную.';
      }

      this.closeConnectModal(true);
      this.selectedBoardId.set(null);
      this.actionNotice.set(notice);
      await this.loadCurrentConnection(connection.id);
    } catch (error) {
      this.modalError.set(this.describeError(error));
    } finally {
      this.connectingConnection.set(false);
    }
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
      this.closeConnectModal(true);
      this.currentConnection.set(null);
      this.tasksError.set(null);
      this.boards.set([]);
      this.tasks.set([]);
      this.selectedBoardId.set(null);
      this.actionNotice.set('Подключение отключено.');
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
      this.closeConnectModal(true);
      this.actionNotice.set('Full sync запущен.');
    } catch (error) {
      this.actionError.set(this.describeError(error));
    }
  }

  protected selectBoard(boardId: string | null): void {
    if (this.selectedBoardId() === boardId) {
      return;
    }

    this.selectedBoardId.set(boardId);
  }

  protected async reloadData(): Promise<void> {
    const connection = this.currentConnection();
    if (!connection) {
      return;
    }

    await Promise.all([this.loadTasks(connection.id), this.loadBoards(connection.id)]);
  }

  protected connectionLabel(connection: YougileConnection): string {
    return connection.title?.trim() || `Компания ${connection.yougile_company_id}`;
  }

  protected connectionStatusLabel(status: string): string {
    switch (status) {
      case 'active':
        return 'Активно';
      case 'invalid':
        return 'Ключ недействителен';
      case 'sync_error':
        return 'Ошибка sync';
      case 'revoked':
        return 'Отключено';
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

  protected taskBoardLabel(task: YougileTask): string {
    if (task.boardTitle?.trim()) {
      return task.boardTitle.trim();
    }

    return this.resolveBoardLabel(task.boardId);
  }

  protected taskColumnLabel(task: YougileTask): string {
    return this.resolveColumnLabel(task.columnId, task.columnTitle);
  }

  protected taskIdentifier(task: YougileTask): string {
    return task.idTaskProject?.trim() || task.idTaskCommon?.trim() || task.id;
  }

  protected taskDeadlineLabel(task: YougileTask): string {
    const deadline = task.deadlineAt ?? task.deadline?.deadline;
    if (!deadline) {
      return 'Без дедлайна';
    }

    const timestamp = new Date(deadline);
    if (Number.isNaN(timestamp.getTime())) {
      return deadline;
    }

    if (task.deadline?.withTime) {
      return this.dateTimeFormatter.format(timestamp);
    }

    return this.dateFormatter.format(timestamp);
  }

  protected trackBoard(_: number, item: YougileBoard): string {
    return item.yougile_board_id;
  }

  protected trackColumn(_: number, item: YougileColumnCard): string {
    return item.columnId;
  }

  protected trackCompany(_: number, item: YougileCompanyOption): string {
    return item.id;
  }

  private async loadCurrentConnection(_preferredConnectionId?: string): Promise<void> {
    this.connectionLoading.set(true);
    this.tasksError.set(null);

    try {
      const response = await firstValueFrom(this.integrationsApi.listYougileConnections());
      const connection = response.items.find((item) => item.status !== 'revoked') ?? null;
      this.currentConnection.set(connection);

      if (!connection) {
        this.tasks.set([]);
        this.boards.set([]);
        this.selectedBoardId.set(null);
        return;
      }

      await Promise.all([this.loadBoards(connection.id), this.loadTasks(connection.id)]);
    } catch (error) {
      this.currentConnection.set(null);
      this.tasks.set([]);
      this.boards.set([]);
      this.selectedBoardId.set(null);
      this.actionError.set(this.describeError(error));
    } finally {
      this.connectionLoading.set(false);
    }
  }

  private async loadBoards(connectionId: string): Promise<void> {
    try {
      const boardsResponse = await firstValueFrom(this.integrationsApi.listYougileBoards(connectionId));

      this.boards.set(
        boardsResponse.items
          .filter((item) => !item.deleted)
          .sort((left, right) => left.title.localeCompare(right.title, 'ru')),
      );
    } catch {
      this.boards.set([]);
    }
  }

  private async loadTasks(connectionId: string): Promise<void> {
    this.tasksLoading.set(true);
    this.tasksError.set(null);

    try {
      const response = await firstValueFrom(
        this.integrationsApi.listYougileTasks(connectionId, {
          limit: 200,
          includeDeleted: false,
          mineOnly: true,
        }),
      );

      this.tasks.set(response.content);
    } catch (error) {
      this.tasks.set([]);
      this.tasksError.set(this.describeError(error));
    } finally {
      this.tasksLoading.set(false);
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

  private resetModalState(resetCredentials: boolean): void {
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
        return 'Нет доступа к компании.';
      }
      if (error.status === 429) {
        return 'Слишком много запросов. Повторите позже.';
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

  private resolveBoardLabel(boardId?: string | null, fallback?: string | null): string {
    if (fallback?.trim()) {
      return fallback.trim();
    }

    if (!boardId?.trim()) {
      return 'Без доски';
    }

    return this.boards().find((item) => item.yougile_board_id === boardId)?.title?.trim() ?? 'Без доски';
  }

  private resolveColumnLabel(columnId: string, fallback?: string | null): string {
    if (fallback?.trim()) {
      return fallback.trim();
    }

    const taskColumn = this.tasks().find((item) => item.columnId === columnId && item.columnTitle?.trim());
    if (taskColumn?.columnTitle?.trim()) {
      return taskColumn.columnTitle.trim();
    }

    const columnCard = this.columnCards().find((item) => item.columnId === columnId);
    return columnCard?.title?.trim() || 'Без колонки';
  }

  private taskSortValue(task: YougileTask): number {
    const deadlineValue = task.deadlineAt ?? task.deadline?.deadline;
    const deadline = deadlineValue ? Date.parse(deadlineValue) : Number.POSITIVE_INFINITY;
    if (Number.isFinite(deadline)) {
      return deadline;
    }

    const timestamp = task.timestamp ? Date.parse(task.timestamp) : Number.POSITIVE_INFINITY;
    return Number.isFinite(timestamp) ? timestamp : Number.POSITIVE_INFINITY;
  }
}
