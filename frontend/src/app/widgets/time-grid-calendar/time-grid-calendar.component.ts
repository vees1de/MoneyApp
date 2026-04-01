import { Component, computed, input, signal, ElementRef, viewChild, afterNextRender } from '@angular/core';

export interface TimeGridEvent {
  id: string;
  title: string;
  start: Date;
  end: Date;
  allDay: boolean;
  color: string;
  textColor: string;
  badge?: string;
  meta?: string;
}

type ViewMode = 'week' | 'day';

interface DayColumn {
  date: Date;
  key: string;
  dayName: string;
  dayNumber: number;
  isToday: boolean;
  timedEvents: PositionedEvent[];
  allDayEvents: TimeGridEvent[];
}

interface PositionedEvent extends TimeGridEvent {
  topPct: number;
  heightPct: number;
  timeLabel: string;
}

const HOURS = Array.from({ length: 24 }, (_, i) => i);

const DAY_SHORT = ['Вс', 'Пн', 'Вт', 'Ср', 'Чт', 'Пт', 'Сб'];
const MONTH_SHORT = [
  'янв', 'фев', 'мар', 'апр', 'май', 'июн',
  'июл', 'авг', 'сен', 'окт', 'ноя', 'дек',
];

@Component({
  selector: 'app-time-grid-calendar',
  standalone: true,
  templateUrl: './time-grid-calendar.component.html',
  styleUrl: './time-grid-calendar.component.scss',
})
export class TimeGridCalendarComponent {
  readonly events = input<TimeGridEvent[]>([]);

  protected readonly viewMode = signal<ViewMode>('week');
  protected readonly anchor = signal(startOfDay(new Date()));
  protected readonly hours = HOURS;

  private readonly bodyRef = viewChild<ElementRef<HTMLElement>>('calBody');

  private readonly timeFmt = new Intl.DateTimeFormat('ru-RU', {
    hour: '2-digit',
    minute: '2-digit',
  });

  protected readonly columns = computed<DayColumn[]>(() => {
    const events = this.events();
    const mode = this.viewMode();
    const base = this.anchor();
    const today = startOfDay(new Date());

    const dates = mode === 'week' ? weekDays(base) : [base];

    return dates.map((date) => {
      const key = dateKey(date);
      const dayEvents = events.filter((e) => eventOnDay(e, date));
      const timed = dayEvents
        .filter((e) => !e.allDay)
        .map((e) => this.positionEvent(e, date));
      const allDay = dayEvents.filter((e) => e.allDay);

      return {
        date,
        key,
        dayName: DAY_SHORT[date.getDay()],
        dayNumber: date.getDate(),
        isToday: key === dateKey(today),
        timedEvents: timed,
        allDayEvents: allDay,
      };
    });
  });

  protected readonly hasAllDay = computed(() =>
    this.columns().some((col) => col.allDayEvents.length > 0),
  );

  protected readonly title = computed(() => {
    const mode = this.viewMode();
    const base = this.anchor();

    if (mode === 'day') {
      return `${base.getDate()} ${MONTH_SHORT[base.getMonth()]} ${base.getFullYear()}`;
    }

    const days = weekDays(base);
    const first = days[0];
    const last = days[6];

    if (first.getMonth() === last.getMonth()) {
      return `${first.getDate()} – ${last.getDate()} ${MONTH_SHORT[first.getMonth()]} ${first.getFullYear()}`;
    }

    return `${first.getDate()} ${MONTH_SHORT[first.getMonth()]} – ${last.getDate()} ${MONTH_SHORT[last.getMonth()]} ${last.getFullYear()}`;
  });

  protected readonly nowPct = computed(() => {
    const now = new Date();
    return ((now.getHours() + now.getMinutes() / 60) / 24) * 100;
  });

  constructor() {
    afterNextRender(() => {
      const el = this.bodyRef()?.nativeElement;
      if (el) {
        const scrollTarget = (8 / 24) * el.scrollHeight;
        el.scrollTop = scrollTarget;
      }
    });
  }

  protected prev(): void {
    const step = this.viewMode() === 'week' ? 7 : 1;
    this.anchor.update((d) => addDays(d, -step));
  }

  protected next(): void {
    const step = this.viewMode() === 'week' ? 7 : 1;
    this.anchor.update((d) => addDays(d, step));
  }

  protected goToday(): void {
    this.anchor.set(startOfDay(new Date()));
  }

  protected setView(mode: ViewMode): void {
    this.viewMode.set(mode);
  }

  protected formatHour(hour: number): string {
    return `${String(hour).padStart(2, '0')}:00`;
  }

  protected trackById(_: number, item: { id: string }): string {
    return item.id;
  }

  protected trackByKey(_: number, item: { key: string }): string {
    return item.key;
  }

  private positionEvent(event: TimeGridEvent, day: Date): PositionedEvent {
    const dayStart = startOfDay(day).getTime();
    const dayEnd = dayStart + 24 * 60 * 60 * 1000;

    const clampedStart = Math.max(event.start.getTime(), dayStart);
    const clampedEnd = Math.min(event.end.getTime(), dayEnd);

    const startMinutes = (clampedStart - dayStart) / 60_000;
    const endMinutes = (clampedEnd - dayStart) / 60_000;
    const totalMinutes = 24 * 60;

    const topPct = (startMinutes / totalMinutes) * 100;
    const heightPct = Math.max(((endMinutes - startMinutes) / totalMinutes) * 100, 0.7);

    const timeLabel = `${this.timeFmt.format(event.start)} – ${this.timeFmt.format(event.end)}`;

    return { ...event, topPct, heightPct, timeLabel };
  }
}

function startOfDay(d: Date): Date {
  const copy = new Date(d);
  copy.setHours(0, 0, 0, 0);
  return copy;
}

function addDays(d: Date, n: number): Date {
  const copy = new Date(d);
  copy.setDate(copy.getDate() + n);
  copy.setHours(0, 0, 0, 0);
  return copy;
}

function weekDays(anchor: Date): Date[] {
  const d = new Date(anchor);
  const dow = d.getDay();
  const mondayOffset = dow === 0 ? -6 : 1 - dow;
  const monday = addDays(d, mondayOffset);
  return Array.from({ length: 7 }, (_, i) => addDays(monday, i));
}

function dateKey(d: Date): string {
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`;
}

function eventOnDay(event: TimeGridEvent, day: Date): boolean {
  const dayStart = startOfDay(day).getTime();
  const dayEnd = dayStart + 24 * 60 * 60 * 1000;

  if (event.allDay) {
    const evStart = startOfDay(event.start).getTime();
    const evEnd = startOfDay(event.end).getTime();
    return evStart <= dayStart && dayStart <= evEnd;
  }

  return event.start.getTime() < dayEnd && event.end.getTime() > dayStart;
}
