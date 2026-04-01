import { CommonModule } from '@angular/common';
import { Component, OnInit, computed, inject, signal } from '@angular/core';
import { RouterLink } from '@angular/router';
import { MatButtonModule } from '@angular/material/button';

import { CalendarApiService } from '@core/api/calendar-api.service';
import type { CalendarUpcomingEvent } from '@core/api/contracts';
import { WidgetShellComponent } from '@app/widgets/widget-shell/widget-shell.component';

@Component({
  selector: 'app-upcoming-events-widget',
  standalone: true,
  imports: [CommonModule, RouterLink, MatButtonModule, WidgetShellComponent],
  templateUrl: './upcoming-events-widget.component.html',
  styleUrl: './upcoming-events-widget.component.scss',
})
export class UpcomingEventsWidgetComponent implements OnInit {
  private readonly api = inject(CalendarApiService);

  protected readonly loading = signal(true);
  protected readonly error = signal<string | null>(null);
  protected readonly events = signal<CalendarUpcomingEvent[]>([]);
  protected readonly activeDate = signal<string | null>(null);
  protected readonly weekDays = computed(() =>
    this.buildWeekDays(this.activeDate() ?? this.dateKey(new Date())),
  );
  protected readonly filteredEvents = computed(() => {
    const selected = this.activeDate();
    if (!selected) {
      return this.events().slice(0, 3);
    }
    return this.events()
      .filter((event) => this.dateKey(event.start_at) === selected)
      .slice(0, 3);
  });

  ngOnInit(): void {
    this.api.getUpcoming(5).subscribe({
      next: (events) => {
        const items = events ?? [];
        this.events.set(items);
        this.activeDate.set(
          items.length > 0 ? this.dateKey(items[0].start_at) : this.dateKey(new Date()),
        );
        this.loading.set(false);
      },
      error: () => {
        this.error.set('failed');
        this.activeDate.set(this.dateKey(new Date()));
        this.loading.set(false);
      },
    });
  }

  protected setActiveDate(value: string): void {
    this.activeDate.set(value);
  }

  protected isDayActive(value: string): boolean {
    return this.activeDate() === value;
  }

  protected eventSubtitle(event: CalendarUpcomingEvent): string {
    const start = new Date(event.start_at);
    const end = new Date(event.end_at);
    if (Number.isNaN(start.getTime())) {
      return 'Дата не указана';
    }
    if (Number.isNaN(end.getTime())) {
      return start.toLocaleString('ru-RU', {
        day: '2-digit',
        month: '2-digit',
        hour: '2-digit',
        minute: '2-digit',
      });
    }
    return `${this.timeLabel(start)} - ${this.timeLabel(end)}`;
  }

  private dateKey(value: string | Date): string {
    const date = value instanceof Date ? value : new Date(value);
    if (Number.isNaN(date.getTime())) {
      return '';
    }
    const year = date.getFullYear();
    const month = String(date.getMonth() + 1).padStart(2, '0');
    const day = String(date.getDate()).padStart(2, '0');
    return `${year}-${month}-${day}`;
  }

  private timeLabel(date: Date): string {
    return date.toLocaleTimeString('ru-RU', {
      hour: '2-digit',
      minute: '2-digit',
    });
  }

  private buildWeekDays(centerDate: string): Array<{ date: string; weekday: string; dayNumber: string }> {
    const base = new Date(centerDate);
    if (Number.isNaN(base.getTime())) {
      return [];
    }

    const days: Array<{ date: string; weekday: string; dayNumber: string }> = [];
    for (let offset = 0; offset < 7; offset += 1) {
      const current = new Date(base);
      current.setDate(base.getDate() + offset);

      days.push({
        date: this.dateKey(current),
        weekday: current.toLocaleDateString('ru-RU', { weekday: 'short' }),
        dayNumber: current.toLocaleDateString('ru-RU', { day: 'numeric' }),
      });
    }

    return days;
  }
}
