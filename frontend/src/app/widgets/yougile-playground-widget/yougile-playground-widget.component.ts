import { CommonModule } from '@angular/common';
import { Component, OnInit, computed, inject, signal } from '@angular/core';
import { MatButtonModule } from '@angular/material/button';
import { firstValueFrom } from 'rxjs';

import type {
  BoardSummary,
  BoardSummaryBoardItem,
  BoardSummaryOverdueTask,
} from '@core/api/contracts';
import { IntegrationsApiService } from '@core/api/integrations-api.service';
import { WidgetShellComponent } from '@app/widgets/widget-shell/widget-shell.component';

@Component({
  selector: 'app-yougile-playground-widget',
  standalone: true,
  imports: [CommonModule, MatButtonModule, WidgetShellComponent],
  templateUrl: './yougile-playground-widget.component.html',
  styleUrl: './yougile-playground-widget.component.scss',
})
export class YougilePlaygroundWidgetComponent implements OnInit {
  private readonly integrationsApi = inject(IntegrationsApiService);
  private readonly dateFormatter = new Intl.DateTimeFormat('ru-RU', {
    dateStyle: 'medium',
  });

  protected readonly loading = signal(true);
  protected readonly error = signal<string | null>(null);
  protected readonly boardOptions = signal<BoardSummaryBoardItem[]>([]);
  protected readonly summary = signal<BoardSummary | null>(null);
  protected readonly selectedBoardId = signal<string | null>(null);

  protected readonly isEmpty = computed(() => {
    const summary = this.summary();
    return !summary || summary.status !== 'ready' || summary.boards.length === 0;
  });

  protected readonly activeBoard = computed(() => {
    const boardId = this.selectedBoardId();
    if (!boardId) {
      return null;
    }

    return this.boardOptions().find((item) => item.board_id === boardId) ?? null;
  });

  ngOnInit(): void {
    void this.loadSummary();
  }

  protected async reload(): Promise<void> {
    await this.loadSummary(this.selectedBoardId(), false);
  }

  protected async selectBoard(boardId: string | null): Promise<void> {
    if (this.selectedBoardId() === boardId) {
      return;
    }

    this.selectedBoardId.set(boardId);
    await this.loadSummary(boardId, false);
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

  protected boardLabel(item: BoardSummaryBoardItem | BoardSummaryOverdueTask): string {
    if ('board_title' in item) {
      return item.board_title?.trim() || 'Без доски';
    }

    return item.title;
  }

  private async loadSummary(
    boardId = this.selectedBoardId(),
    preserveBoardOptions = true,
  ): Promise<void> {
    this.loading.set(true);
    this.error.set(null);

    try {
      const summary = await firstValueFrom(
        this.integrationsApi.getAgileBoardSummary({
          board_id: boardId ?? undefined,
        }),
      );

      this.summary.set(summary);
      this.selectedBoardId.set(boardId);

      if (!preserveBoardOptions || this.boardOptions().length === 0 || !boardId) {
        this.boardOptions.set(summary.boards);
      }
    } catch {
      this.error.set('failed');
      this.summary.set(null);
    } finally {
      this.loading.set(false);
    }
  }
}
