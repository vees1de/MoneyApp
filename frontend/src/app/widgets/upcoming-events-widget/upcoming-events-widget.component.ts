import { HttpErrorResponse } from '@angular/common/http';
import { CommonModule } from '@angular/common';
import { Component, OnInit, computed, inject, signal } from '@angular/core';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { RouterLink } from '@angular/router';
import { firstValueFrom } from 'rxjs';

import { IntegrationsApiService } from '@core/api/integrations-api.service';
import {
  pickPreferredYougileConnection,
  resolveYougileTaskWindow,
  type YougileTask,
} from '@entities/yougile';
import { WidgetShellComponent } from '@app/widgets/widget-shell/widget-shell.component';

type MiniCalendarEvent = {
  id: string;
  title: string;
  start: Date;
  end: Date;
  dayKey: string;
  timeLabel: string;
  primaryMetaLabel: string;
  primaryMetaValue: string;
  secondaryMetaLabel: string;
  secondaryMetaValue: string;
  visualLabel: string;
  visualTone: 'timed' | 'all-day' | 'completed';
};

type MiniCalendarConflict = {
  id: string;
  dayKey: string;
  timeLabel: string;
  title: string;
  reason: string;
};

type MiniCalendarDay = {
  key: string;
  name: string;
  number: string;
  isActive: boolean;
  hasItems: boolean;
  hasConflict: boolean;
};

const WEEKDAY_LABELS = ['Вс', 'Пн', 'Вт', 'Ср', 'Чт', 'Пт', 'Сб'];

@Component({
  selector: 'app-upcoming-events-widget',
  standalone: true,
  imports: [CommonModule, RouterLink, MatButtonModule, MatIconModule, WidgetShellComponent],
  templateUrl: './upcoming-events-widget.component.html',
  styleUrl: './upcoming-events-widget.component.scss',
})
export class UpcomingEventsWidgetComponent implements OnInit {
  private readonly integrationsApi = inject(IntegrationsApiService);

  protected readonly loading = signal(true);
  protected readonly error = signal<string | null>(null);
  protected readonly tasks = signal<YougileTask[]>([]);
  protected readonly selectedDayKey = signal(this.formatDateKey(new Date()));

  protected readonly calendarEvents = computed<MiniCalendarEvent[]>(() =>
    this.normalizeTasks(this.tasks()),
  );

  protected readonly conflicts = computed<MiniCalendarConflict[]>(() =>
    this.buildConflicts(this.calendarEvents()),
  );

  protected readonly selectedDayEvents = computed<MiniCalendarEvent[]>(() =>
    this.calendarEvents()
      .filter((event) => event.dayKey === this.selectedDayKey())
      .slice(0, 2),
  );

  protected readonly selectedDayConflict = computed<MiniCalendarConflict | null>(
    () => this.conflicts().find((item) => item.dayKey === this.selectedDayKey()) ?? null,
  );

  protected readonly visibleWeek = computed<MiniCalendarDay[]>(() => {
    const anchor = this.parseDateKey(this.selectedDayKey()) ?? new Date();
    const monday = this.startOfWeek(anchor);

    return Array.from({ length: 7 }, (_, index) => {
      const date = this.addDays(monday, index);
      const key = this.formatDateKey(date);

      return {
        key,
        name: WEEKDAY_LABELS[date.getDay()],
        number: String(date.getDate()),
        isActive: key === this.selectedDayKey(),
        hasItems: this.calendarEvents().some((event) => event.dayKey === key),
        hasConflict: this.conflicts().some((conflict) => conflict.dayKey === key),
      };
    });
  });

  protected readonly emptyDay = computed(
    () => !this.selectedDayEvents().length && !this.selectedDayConflict(),
  );
  protected readonly unscheduledCount = computed(
    () => this.tasks().filter((task) => !resolveYougileTaskWindow(task)).length,
  );

  ngOnInit(): void {
    void this.loadCalendar();
  }

  protected selectDay(dayKey: string): void {
    this.selectedDayKey.set(dayKey);
  }

  protected selectNextConflictDay(): void {
    const allConflicts = this.conflicts();
    if (!allConflicts.length) {
      return;
    }

    const currentKey = this.selectedDayKey();
    const currentIndex = allConflicts.findIndex((item) => item.dayKey === currentKey);
    const next = allConflicts[(currentIndex + 1 + allConflicts.length) % allConflicts.length];
    this.selectedDayKey.set(next.dayKey);
  }

  private alignSelection(): void {
    const current = this.selectedDayKey();
    if (
      this.calendarEvents().some((event) => event.dayKey === current) ||
      this.conflicts().some((conflict) => conflict.dayKey === current)
    ) {
      return;
    }

    const firstConflict = this.conflicts()[0];
    if (firstConflict) {
      this.selectedDayKey.set(firstConflict.dayKey);
      return;
    }

    const nextEvent = this.calendarEvents()[0];
    if (nextEvent) {
      this.selectedDayKey.set(nextEvent.dayKey);
      return;
    }

    this.selectedDayKey.set(this.formatDateKey(new Date()));
  }

  private async loadCalendar(): Promise<void> {
    this.loading.set(true);
    this.error.set(null);

    try {
      const response = await firstValueFrom(this.integrationsApi.listYougileConnections());
      const connection = pickPreferredYougileConnection(response.items ?? []);

      if (!connection) {
        this.tasks.set([]);
        this.alignSelection();
        return;
      }

      const tasksResponse = await firstValueFrom(
        this.integrationsApi.listYougileTasks(connection.id, {
          includeDeleted: false,
          limit: 200,
        }),
      );

      this.tasks.set(tasksResponse.content ?? []);
      this.alignSelection();
    } catch (error) {
      this.tasks.set([]);
      this.error.set(this.describeError(error));
    } finally {
      this.loading.set(false);
    }
  }

  private normalizeTasks(items: YougileTask[]): MiniCalendarEvent[] {
    return items
      .flatMap((task) => {
        const window = resolveYougileTaskWindow(task);
        if (!window) {
          return [];
        }

        return [
          {
            id: task.id,
            title: task.title?.trim() || 'Yougile task',
            start: window.start,
            end: window.end,
            dayKey: this.formatDateKey(window.start),
            timeLabel: window.allDay ? 'Весь день' : this.formatTimeRange(window.start, window.end),
            primaryMetaLabel: 'Колонка',
            primaryMetaValue: this.taskColumnLabel(task),
            secondaryMetaLabel: 'Доска',
            secondaryMetaValue: this.taskBoardLabel(task),
            visualLabel: this.taskVisualLabel(task),
            visualTone: this.taskVisualTone(task, window.allDay),
          },
        ];
      })
      .sort((left, right) => left.start.getTime() - right.start.getTime());
  }

  private buildConflicts(events: MiniCalendarEvent[]): MiniCalendarConflict[] {
    const conflicts = new Map<string, MiniCalendarConflict>();

    for (let leftIndex = 0; leftIndex < events.length; leftIndex += 1) {
      const left = events[leftIndex];

      for (let rightIndex = leftIndex + 1; rightIndex < events.length; rightIndex += 1) {
        const right = events[rightIndex];

        if (left.dayKey !== right.dayKey) {
          continue;
        }

        if (
          left.start.getTime() >= right.end.getTime() ||
          right.start.getTime() >= left.end.getTime()
        ) {
          continue;
        }

        const overlapStart =
          left.start.getTime() >= right.start.getTime() ? left.start : right.start;
        const overlapEnd = left.end.getTime() <= right.end.getTime() ? left.end : right.end;
        const id = [left.id, right.id].sort().join('::');

        conflicts.set(id, {
          id,
          dayKey: left.dayKey,
          title: 'Конфликт времени',
          timeLabel: this.formatTimeRange(overlapStart, overlapEnd),
          reason: `${left.title} пересекается с ${right.title}`,
        });
      }
    }

    return [...conflicts.values()].sort((left, right) =>
      left.timeLabel.localeCompare(right.timeLabel, 'ru'),
    );
  }

  private taskBoardLabel(task: YougileTask): string {
    if (task.boardTitle?.trim()) {
      return task.boardTitle.trim();
    }

    return 'Без доски';
  }

  private taskColumnLabel(task: YougileTask): string {
    if (task.columnTitle?.trim()) {
      return task.columnTitle.trim();
    }

    return 'Без колонки';
  }

  private taskIdentifier(task: YougileTask): string {
    return task.idTaskProject?.trim() || task.idTaskCommon?.trim() || task.id;
  }

  private taskVisualLabel(task: YougileTask): string {
    const identifier = this.taskIdentifier(task);
    const compactIdentifier = identifier.replace(/[^A-Za-zА-Яа-я0-9]/g, '');
    if (compactIdentifier) {
      return compactIdentifier.slice(0, 4).toUpperCase();
    }

    const initials = (task.title ?? '')
      .replace(/[^A-Za-zА-Яа-я0-9 ]/g, ' ')
      .split(/\s+/)
      .filter(Boolean)
      .slice(0, 2)
      .map((part) => part[0]?.toUpperCase() ?? '')
      .join('');

    return initials || 'YG';
  }

  private taskVisualTone(task: YougileTask, allDay: boolean): 'timed' | 'all-day' | 'completed' {
    if (task.completed || task.archived) {
      return 'completed';
    }

    return allDay ? 'all-day' : 'timed';
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
    }

    return 'Не удалось загрузить Yougile задачи.';
  }

  private extractApiMessage(payload: unknown): string | null {
    if (typeof payload === 'string' && payload.trim()) {
      return payload.trim();
    }

    if (!payload || typeof payload !== 'object') {
      return null;
    }

    const errorPayload = (payload as { error?: unknown }).error;
    if (!errorPayload || typeof errorPayload !== 'object') {
      return null;
    }

    const message = (errorPayload as { message?: unknown }).message;
    return typeof message === 'string' && message.trim() ? message.trim() : null;
  }

  private formatTimeRange(start: Date, end: Date): string {
    const formatter = new Intl.DateTimeFormat('ru-RU', {
      hour: '2-digit',
      minute: '2-digit',
    });
    return `${formatter.format(start)} - ${formatter.format(end)}`;
  }

  private formatDateKey(date: Date): string {
    const year = date.getFullYear();
    const month = String(date.getMonth() + 1).padStart(2, '0');
    const day = String(date.getDate()).padStart(2, '0');
    return `${year}-${month}-${day}`;
  }

  private parseDateKey(value: string | null | undefined): Date | null {
    if (!value) {
      return null;
    }

    const parsed = new Date(`${value}T00:00:00`);
    return Number.isNaN(parsed.getTime()) ? null : parsed;
  }

  private startOfWeek(date: Date): Date {
    const result = new Date(date);
    result.setHours(0, 0, 0, 0);
    const day = result.getDay();
    result.setDate(result.getDate() + (day === 0 ? -6 : 1 - day));
    return result;
  }

  private addDays(date: Date, days: number): Date {
    const result = new Date(date);
    result.setDate(result.getDate() + days);
    return result;
  }
}
