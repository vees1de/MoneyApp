import { CommonModule } from '@angular/common';
import { Component, inject } from '@angular/core';
import { RouterLink } from '@angular/router';
import { CalendarSyncFacade } from '@features/calendar-sync';
import type { CalendarEvent } from '@entities/calendar-event';

@Component({
  selector: 'app-page-calendar-conflicts',
  standalone: true,
  imports: [CommonModule, RouterLink],
  templateUrl: './conflicts.page.html',
  styleUrl: './conflicts.page.scss',
})
export class CalendarConflictsPageComponent {
  private readonly facade = inject(CalendarSyncFacade);
  protected readonly routePath = '/calendar/conflicts';
  protected readonly entitySample: CalendarEvent[] = [];

  protected loadPage(): void {
    this.facade.load();
  }
}
