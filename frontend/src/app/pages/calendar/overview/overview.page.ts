import { HttpErrorResponse } from '@angular/common/http';
import { CommonModule } from '@angular/common';
import { Component, OnInit, computed, inject, signal } from '@angular/core';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatProgressBarModule } from '@angular/material/progress-bar';
import { FullCalendarModule } from '@fullcalendar/angular';
import type { CalendarOptions, EventInput, EventMountArg } from '@fullcalendar/core';
import ruLocale from '@fullcalendar/core/locales/ru';
import dayGridPlugin from '@fullcalendar/daygrid';
import interactionPlugin from '@fullcalendar/interaction';
import timeGridPlugin from '@fullcalendar/timegrid';
import { RouterLink } from '@angular/router';
import { firstValueFrom } from 'rxjs';

import type { LearningPlanItem, MyLearningPlan } from '@core/api/contracts';
import { IntegrationsApiService } from '@core/api/integrations-api.service';
import { LearningPlanApiService } from '@core/api/learning-plan-api.service';
import { pickPreferredYougileConnection, resolveYougileTaskWindow } from '@entities/yougile';
import type { YougileConnection, YougileTask } from '@entities/yougile';

type CalendarSource = 'yougile' | 'learning';

interface TimelineRange {
  id: string;
  source: CalendarSource;
  title: string;
  badge: string;
  meta: string;
  start: Date;
  end: Date;
  allDay: boolean;
  backgroundColor: string;
  textColor: string;
}

interface LearningDeadlinePoint {
  id: string;
  title: string;
  status: string;
  statusLabel: string;
  deadline: Date;
  deadlineLabel: string;
}

interface CalendarRiskBlock {
  id: string;
  title: string;
  reason: string;
  window: string;
  severity: 'high' | 'medium';
}

interface LearningSetCard {
  id: string;
  title: string;
  status: string;
  statusLabel: string;
  bucket: 'in_progress' | 'upcoming';
  deadlineLabel: string;
  deadlineMs: number | null;
}

@Component({
  selector: 'app-page-calendar-overview',
  standalone: true,
  imports: [
    CommonModule,
    RouterLink,
    FullCalendarModule,
    MatButtonModule,
    MatIconModule,
    MatProgressBarModule,
  ],
  templateUrl: './overview.page.html',
  styleUrl: './overview.page.scss',
})
export class CalendarOverviewPageComponent implements OnInit {
  private readonly integrationsApi = inject(IntegrationsApiService);
  private readonly learningPlanApi = inject(LearningPlanApiService);

  private readonly dateFormatter = new Intl.DateTimeFormat('ru-RU', {
    day: '2-digit',
    month: 'short',
  });
  private readonly timeFormatter = new Intl.DateTimeFormat('ru-RU', {
    hour: '2-digit',
    minute: '2-digit',
  });
  private readonly dateTimeFormatter = new Intl.DateTimeFormat('ru-RU', {
    day: '2-digit',
    month: 'short',
    hour: '2-digit',
    minute: '2-digit',
  });

  protected readonly calendarLoading = signal(true);
  protected readonly yougileLoading = signal(true);
  protected readonly learningLoading = signal(true);
  protected readonly yougileError = signal<string | null>(null);
  protected readonly learningError = signal<string | null>(null);
  protected readonly yougileConnection = signal<YougileConnection | null>(null);
  protected readonly yougileTasks = signal<YougileTask[]>([]);
  protected readonly learningPlan = signal<MyLearningPlan | null>(null);

  protected readonly anyBusy = computed(
    () => this.calendarLoading() || this.yougileLoading() || this.learningLoading(),
  );
  protected readonly issueMessages = computed(() => {
    const messages = [this.yougileError(), this.learningError()].filter(
      (value): value is string => !!value,
    );
    return Array.from(new Set(messages));
  });
  protected readonly timelineRanges = computed(() => this.mapYougileRanges(this.yougileTasks()));
  protected readonly learningDeadlinePoints = computed(() => this.mapLearningDeadlinePoints());
  protected readonly learningSets = computed(() => this.mapLearningSets());
  protected readonly riskBlocks = computed(() => this.buildRiskBlocks());
  protected readonly calendarEvents = computed<EventInput[]>(() => this.buildCalendarEvents());
  protected readonly timeblockCount = computed(() => this.timelineRanges().length);
  protected readonly currentSetsCount = computed(() => this.learningSets().length);
  protected readonly riskCount = computed(() => this.riskBlocks().length);
  protected readonly calendarOptions: CalendarOptions = {
    plugins: [dayGridPlugin, timeGridPlugin, interactionPlugin],
    locale: ruLocale,
    initialView: 'timeGridWeek',
    headerToolbar: {
      left: 'prev,next today',
      center: 'title',
      right: 'timeGridWeek,timeGridDay,dayGridMonth',
    },
    buttonText: {
      today: 'Сегодня',
      week: 'Неделя',
      day: 'День',
      month: 'Месяц',
    },
    firstDay: 1,
    weekends: true,
    editable: false,
    selectable: false,
    dayMaxEventRows: 3,
    allDaySlot: true,
    nowIndicator: true,
    expandRows: true,
    stickyHeaderDates: true,
    height: 'auto',
    contentHeight: 'auto',
    eventDisplay: 'block',
    scrollTime: '08:00:00',
    slotMinTime: '06:00:00',
    slotMaxTime: '24:00:00',
    eventTimeFormat: {
      hour: '2-digit',
      minute: '2-digit',
      hour12: false,
    },
    eventDidMount: (arg) => this.decorateCalendarEvent(arg),
  };

  ngOnInit(): void {
    void this.refreshOverview();
  }

  protected async refreshOverview(): Promise<void> {
    this.calendarLoading.set(true);
    this.clearLocalIssues();

    try {
      await Promise.all([this.loadYougileOverview(), this.loadLearningPlan()]);
    } finally {
      this.calendarLoading.set(false);
    }
  }

  protected clearError(): void {
    this.yougileError.set(null);
    this.learningError.set(null);
  }

  protected yougileConnectionStatusLabel(status: string | null | undefined): string {
    switch ((status ?? '').toLowerCase()) {
      case 'active':
        return 'Активно';
      case 'sync_error':
        return 'Ошибка sync';
      case 'invalid':
        return 'Ключ недействителен';
      case 'revoked':
        return 'Отключено';
      case '':
        return 'Не подключено';
      default:
        return status || '—';
    }
  }

  protected learningStatusLabel(status: string): string {
    switch ((status ?? '').toLowerCase()) {
      case 'in_progress':
        return 'В процессе';
      case 'upcoming':
        return 'Скоро старт';
      case 'completed':
        return 'Завершен';
      case 'enrolled':
        return 'Записан';
      default:
        return status || '—';
    }
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

  protected taskBoardLabel(task: YougileTask): string {
    if (task.boardTitle?.trim()) {
      return task.boardTitle.trim();
    }

    return 'Без доски';
  }

  protected taskColumnLabel(task: YougileTask): string {
    if (task.columnTitle?.trim()) {
      return task.columnTitle.trim();
    }

    return 'Без колонки';
  }

  protected formatRiskSeverity(severity: 'high' | 'medium'): string {
    return severity === 'high' ? 'Высокий' : 'Средний';
  }

  private async loadYougileOverview(): Promise<void> {
    this.yougileLoading.set(true);
    this.yougileError.set(null);

    try {
      const response = await firstValueFrom(this.integrationsApi.listYougileConnections());
      const connection = pickPreferredYougileConnection(response.items ?? []);
      this.yougileConnection.set(connection);

      if (!connection) {
        this.yougileTasks.set([]);
        return;
      }

      try {
        const tasksResponse = await firstValueFrom(
          this.integrationsApi.listYougileTasks(connection.id, {
            includeDeleted: false,
            limit: 200,
          }),
        );

        this.yougileTasks.set(tasksResponse.content ?? []);
      } catch (error) {
        this.yougileTasks.set([]);
        this.yougileError.set(this.describeError(error, 'Не удалось загрузить Yougile задачи.'));
      }
    } catch (error) {
      this.yougileConnection.set(null);
      this.yougileTasks.set([]);
      this.yougileError.set(this.describeError(error, 'Не удалось загрузить Yougile задачи.'));
    } finally {
      this.yougileLoading.set(false);
    }
  }

  private async loadLearningPlan(): Promise<void> {
    this.learningLoading.set(true);
    this.learningError.set(null);

    try {
      const plan = await firstValueFrom(this.learningPlanApi.getMyPlan());
      this.learningPlan.set(this.normalizeLearningPlan(plan));
    } catch (error) {
      this.learningPlan.set(null);
      this.learningError.set(this.describeError(error, 'Не удалось загрузить текущие наборы.'));
    } finally {
      this.learningLoading.set(false);
    }
  }

  private mapYougileRanges(tasks: YougileTask[]): TimelineRange[] {
    return tasks.flatMap((task) => {
      const window = resolveYougileTaskWindow(task);
      if (!window) {
        return [];
      }

      return [
        {
          id: `yougile-${task.id}`,
          source: 'yougile' as CalendarSource,
          title: task.title?.trim() || 'Yougile task',
          badge: 'Yougile',
          meta: [this.taskBoardLabel(task), this.taskColumnLabel(task)]
            .filter((value) => !!value)
            .join(' · '),
          start: window.start,
          end: window.end,
          allDay: window.allDay,
          backgroundColor: window.allDay ? '#7c3aed' : '#0f766e',
          textColor: '#ffffff',
        },
      ];
    });
  }

  private mapLearningDeadlinePoints(): LearningDeadlinePoint[] {
    const plan = this.learningPlan();
    if (!plan) {
      return [];
    }

    return [
      ...this.learningPlanItems(plan.in_progress).map((item) =>
        this.mapLearningDeadlinePoint(item, 'in_progress'),
      ),
      ...this.learningPlanItems(plan.upcoming).map((item) =>
        this.mapLearningDeadlinePoint(item, 'upcoming'),
      ),
    ].filter((item): item is LearningDeadlinePoint => item !== null);
  }

  private mapLearningSets(): LearningSetCard[] {
    const plan = this.learningPlan();
    if (!plan) {
      return [];
    }

    return [
      ...this.learningPlanItems(plan.in_progress).map((item) =>
        this.mapLearningSet(item, 'in_progress'),
      ),
      ...this.learningPlanItems(plan.upcoming).map((item) => this.mapLearningSet(item, 'upcoming')),
    ].sort((left, right) => {
      const leftDeadline = left.deadlineMs ?? Number.POSITIVE_INFINITY;
      const rightDeadline = right.deadlineMs ?? Number.POSITIVE_INFINITY;
      return leftDeadline - rightDeadline;
    });
  }

  private buildCalendarEvents(): EventInput[] {
    const events: EventInput[] = [];

    for (const range of this.timelineRanges()) {
      events.push({
        id: range.id,
        title: range.title,
        start: range.allDay ? this.formatLocalDate(range.start) : range.start.toISOString(),
        end: range.allDay
          ? this.formatLocalDate(this.addDays(this.startOfDay(range.end), 1))
          : range.end.toISOString(),
        allDay: range.allDay,
        backgroundColor: range.backgroundColor,
        borderColor: range.backgroundColor,
        textColor: range.textColor,
        extendedProps: {
          badge: range.badge,
          meta: range.meta,
          source: range.source,
        },
      });
    }

    for (const point of this.learningDeadlinePoints()) {
      events.push({
        id: `learning-${point.id}`,
        title: point.title,
        start: this.formatLocalDate(point.deadline),
        end: this.formatLocalDate(this.addDays(this.startOfDay(point.deadline), 1)),
        allDay: true,
        backgroundColor: '#7c3aed',
        borderColor: '#7c3aed',
        textColor: '#ffffff',
        extendedProps: {
          badge: 'Набор',
          meta: point.statusLabel,
          source: 'learning',
        },
      });
    }

    return events;
  }

  private buildRiskBlocks(): CalendarRiskBlock[] {
    const risks = new Map<string, CalendarRiskBlock>();
    const ranges = this.timelineRanges();
    const points = this.learningDeadlinePoints();

    for (let leftIndex = 0; leftIndex < ranges.length; leftIndex += 1) {
      const left = ranges[leftIndex];

      for (let rightIndex = leftIndex + 1; rightIndex < ranges.length; rightIndex += 1) {
        const right = ranges[rightIndex];
        if (!this.rangesOverlap(left, right)) {
          continue;
        }

        const key = [left.id, right.id].sort().join('::');
        const overlapStart =
          left.start.getTime() >= right.start.getTime() ? left.start : right.start;
        const overlapEnd = left.end.getTime() <= right.end.getTime() ? left.end : right.end;
        risks.set(key, {
          id: key,
          title: `${left.title} и ${right.title}`,
          reason: `${left.badge} и ${right.badge} пересекаются по времени`,
          window: this.formatWindow(overlapStart, overlapEnd, left.allDay && right.allDay),
          severity: 'high',
        });
      }
    }

    for (const point of points) {
      for (const range of ranges) {
        if (
          point.deadline.getTime() < range.start.getTime() ||
          point.deadline.getTime() > range.end.getTime()
        ) {
          continue;
        }

        const key = `${point.id}::${range.id}`;
        risks.set(key, {
          id: key,
          title: `${point.title} и ${range.title}`,
          reason: `Дедлайн набора попадает в блок ${range.badge.toLowerCase()}`,
          window: this.formatWindow(range.start, range.end, range.allDay),
          severity: 'medium',
        });
      }
    }

    return Array.from(risks.values()).sort((left, right) => {
      const leftRank = left.severity === 'high' ? 0 : 1;
      const rightRank = right.severity === 'high' ? 0 : 1;
      if (leftRank !== rightRank) {
        return leftRank - rightRank;
      }

      return left.window.localeCompare(right.window, 'ru');
    });
  }

  private mapLearningDeadlinePoint(
    item: LearningPlanItem,
    bucket: 'in_progress' | 'upcoming',
  ): LearningDeadlinePoint | null {
    if (!item.deadline_at) {
      return null;
    }

    const deadline = this.parseDate(item.deadline_at);
    if (!deadline) {
      return null;
    }

    return {
      id: `${bucket}-${item.enrollment_id}`,
      title: item.title,
      status: item.status,
      statusLabel: this.learningStatusLabel(item.status),
      deadline,
      deadlineLabel: this.formatLearningDate(deadline),
    };
  }

  private mapLearningSet(
    item: LearningPlanItem,
    bucket: 'in_progress' | 'upcoming',
  ): LearningSetCard {
    const deadline = item.deadline_at ? this.parseDate(item.deadline_at) : null;

    return {
      id: `${bucket}-${item.enrollment_id}`,
      title: item.title,
      status: item.status,
      statusLabel: this.learningStatusLabel(item.status),
      bucket,
      deadlineLabel: deadline ? this.formatLearningDate(deadline) : 'Без дедлайна',
      deadlineMs: deadline?.getTime() ?? null,
    };
  }

  private formatLearningDate(value: string | Date): string {
    const parsed = value instanceof Date ? value : this.parseDate(value);
    if (!parsed) {
      return typeof value === 'string' ? value : '—';
    }

    return this.dateTimeFormatter.format(parsed);
  }

  private formatWindow(start: Date, end: Date, allDay: boolean): string {
    if (allDay) {
      return this.dateFormatter.format(start);
    }

    if (this.isSameDay(start, end)) {
      return `${this.dateFormatter.format(start)} · ${this.timeFormatter.format(start)}–${this.timeFormatter.format(end)}`;
    }

    return `${this.dateTimeFormatter.format(start)} → ${this.dateTimeFormatter.format(end)}`;
  }

  private rangesOverlap(left: TimelineRange, right: TimelineRange): boolean {
    return (
      left.start.getTime() <= right.end.getTime() && right.start.getTime() <= left.end.getTime()
    );
  }

  private describeError(error: unknown, fallback: string): string {
    if (error instanceof HttpErrorResponse) {
      const apiMessage = this.extractApiMessage(error.error);
      if (apiMessage) {
        return apiMessage;
      }

      if (error.status === 0) {
        return 'Backend недоступен.';
      }
    }

    return fallback;
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

  private normalizeLearningPlan(plan: MyLearningPlan | null | undefined): MyLearningPlan | null {
    if (!plan) {
      return null;
    }

    return {
      ...plan,
      in_progress: Array.isArray(plan.in_progress) ? plan.in_progress : [],
      upcoming: Array.isArray(plan.upcoming) ? plan.upcoming : [],
      completed_recently: Array.isArray(plan.completed_recently) ? plan.completed_recently : [],
      recommended: Array.isArray(plan.recommended) ? plan.recommended : [],
    };
  }

  private learningPlanItems(items: LearningPlanItem[] | null | undefined): LearningPlanItem[] {
    return Array.isArray(items) ? items : [];
  }

  private clearLocalIssues(): void {
    this.yougileError.set(null);
    this.learningError.set(null);
  }

  private parseDate(value: string | null | undefined): Date | null {
    if (!value) {
      return null;
    }

    const parsed = new Date(value);
    return Number.isNaN(parsed.getTime()) ? null : parsed;
  }

  private startOfDay(value: Date): Date {
    const copy = new Date(value);
    copy.setHours(0, 0, 0, 0);
    return copy;
  }

  private endOfDay(value: Date): Date {
    const copy = new Date(value);
    copy.setHours(23, 59, 59, 999);
    return copy;
  }

  private addDays(value: Date, days: number): Date {
    const copy = new Date(value);
    copy.setDate(copy.getDate() + days);
    return copy;
  }

  private formatLocalDate(value: Date): string {
    const year = value.getFullYear();
    const month = String(value.getMonth() + 1).padStart(2, '0');
    const day = String(value.getDate()).padStart(2, '0');
    return `${year}-${month}-${day}`;
  }

  private isSameDay(left: Date, right: Date): boolean {
    return (
      left.getFullYear() === right.getFullYear() &&
      left.getMonth() === right.getMonth() &&
      left.getDate() === right.getDate()
    );
  }

  private decorateCalendarEvent(arg: EventMountArg): void {
    const source = arg.event.extendedProps['source'];
    const meta = arg.event.extendedProps['meta'];
    if (source === 'yougile' || source === 'learning') {
      arg.el.setAttribute('data-source', source);
    }

    const titleParts = [arg.event.title];
    if (typeof meta === 'string' && meta.trim()) {
      titleParts.push(meta.trim());
    }
    arg.el.setAttribute('title', titleParts.join(' · '));
  }
}
