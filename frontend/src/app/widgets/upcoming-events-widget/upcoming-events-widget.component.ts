import { CommonModule } from '@angular/common';
import { Component, OnInit, inject, signal } from '@angular/core';
import { RouterLink } from '@angular/router';
import { MatListModule } from '@angular/material/list';

import { CalendarApiService } from '@core/api/calendar-api.service';
import type { CalendarUpcomingEvent } from '@core/api/contracts';
import { WidgetShellComponent } from '@app/widgets/widget-shell/widget-shell.component';

@Component({
  selector: 'app-upcoming-events-widget',
  standalone: true,
  imports: [CommonModule, RouterLink, MatListModule, WidgetShellComponent],
  templateUrl: './upcoming-events-widget.component.html',
  styleUrl: './upcoming-events-widget.component.scss',
})
export class UpcomingEventsWidgetComponent implements OnInit {
  private readonly api = inject(CalendarApiService);

  protected readonly loading = signal(true);
  protected readonly error = signal<string | null>(null);
  protected readonly events = signal<CalendarUpcomingEvent[]>([]);

  ngOnInit(): void {
    this.api.getUpcoming(5).subscribe({
      next: (events) => {
        this.events.set(events ?? []);
        this.loading.set(false);
      },
      error: () => {
        this.error.set('failed');
        this.loading.set(false);
      },
    });
  }

  protected eventSubtitle(event: CalendarUpcomingEvent): string {
    const dt = new Date(event.start_at);
    if (Number.isNaN(dt.getTime())) {
      return 'Дата не указана';
    }
    return dt.toLocaleString('ru-RU', {
      day: '2-digit',
      month: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
    });
  }
}
