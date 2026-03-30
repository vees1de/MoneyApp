import { CommonModule } from '@angular/common';
import { Component, inject } from '@angular/core';
import { RouterLink } from '@angular/router';
import { CalendarSyncFacade } from '@features/calendar-sync';
import type { CalendarEvent } from '@entities/calendar-event';

@Component({
  selector: 'app-page-calendar-overview',
  standalone: true,
  imports: [CommonModule, RouterLink],
  templateUrl: './overview.page.html',
  styleUrl: './overview.page.scss',
})
export class CalendarOverviewPageComponent {
  private readonly facade = inject(CalendarSyncFacade);
  protected readonly routePath = '/calendar/overview';
  protected readonly entitySample: CalendarEvent[] = [];

  protected loadPage(): void {
    this.facade.load();
  }
}
