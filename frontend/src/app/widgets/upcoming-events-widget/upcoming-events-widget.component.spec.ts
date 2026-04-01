import { TestBed } from '@angular/core/testing';
import { provideRouter } from '@angular/router';
import { of } from 'rxjs';

import { IntegrationsApiService } from '@core/api/integrations-api.service';
import type { YougileTask } from '@entities/yougile';
import { UpcomingEventsWidgetComponent } from './upcoming-events-widget.component';

describe('UpcomingEventsWidgetComponent', () => {
  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [UpcomingEventsWidgetComponent],
      providers: [
        provideRouter([]),
        {
          provide: IntegrationsApiService,
          useValue: {
            listYougileConnections: () => of({ items: [] }),
            listYougileTasks: () =>
              of({ paging: { count: 0, limit: 200, offset: 0, next: false }, content: [] }),
          },
        },
      ],
    }).compileComponents();
  });

  it('maps a timed yougile task into a mini-calendar entry', () => {
    const fixture = TestBed.createComponent(UpcomingEventsWidgetComponent);
    const component = fixture.componentInstance as UpcomingEventsWidgetComponent & {
      tasks: { set(value: YougileTask[]): void };
      selectedDayKey: { set(value: string): void };
      selectedDayEvents: () => Array<{
        title: string;
        timeLabel: string;
        primaryMetaValue: string;
        secondaryMetaValue: string;
      }>;
    };

    component.tasks.set([
      {
        id: 'task-1',
        deleted: false,
        title: 'Фокус-сессия',
        timestamp: '2026-04-23T12:00:00Z',
        columnId: 'col-1',
        columnTitle: 'В работе',
        boardId: 'board-1',
        boardTitle: 'Продукт',
        archived: false,
        completed: false,
        deadline: {
          deadline: '2026-04-23T17:00:00',
          startDate: '2026-04-23T15:00:00',
          withTime: true,
        },
        deadlineAt: '2026-04-23T17:00:00',
      },
    ]);

    component.selectedDayKey.set('2026-04-23');

    expect(component.selectedDayEvents()).toHaveLength(1);
    expect(component.selectedDayEvents()[0]).toMatchObject({
      title: 'Фокус-сессия',
      timeLabel: '15:00 - 17:00',
      primaryMetaValue: 'В работе',
      secondaryMetaValue: 'Продукт',
    });
  });

  it('builds a time conflict for overlapping yougile tasks on the same day', () => {
    const fixture = TestBed.createComponent(UpcomingEventsWidgetComponent);
    const component = fixture.componentInstance as UpcomingEventsWidgetComponent & {
      tasks: { set(value: YougileTask[]): void };
      selectedDayKey: { set(value: string): void };
      conflicts: () => Array<{ title: string; reason: string; dayKey: string; timeLabel: string }>;
      selectedDayConflict: () => {
        title: string;
        reason: string;
        dayKey: string;
        timeLabel: string;
      } | null;
    };

    component.tasks.set([
      {
        id: 'task-1',
        deleted: false,
        title: 'Курс Python - Basic',
        timestamp: '2026-04-23T12:00:00Z',
        columnId: 'col-1',
        archived: false,
        completed: false,
        deadline: {
          deadline: '2026-04-23T17:00:00',
          startDate: '2026-04-23T15:00:00',
          withTime: true,
        },
        deadlineAt: '2026-04-23T17:00:00',
      },
      {
        id: 'task-2',
        deleted: false,
        title: 'Командный созвон',
        timestamp: '2026-04-23T12:30:00Z',
        columnId: 'col-2',
        archived: false,
        completed: false,
        deadline: {
          deadline: '2026-04-23T17:00:00',
          startDate: '2026-04-23T16:00:00',
          withTime: true,
        },
        deadlineAt: '2026-04-23T17:00:00',
      },
    ]);

    component.selectedDayKey.set('2026-04-23');

    expect(component.conflicts()).toHaveLength(1);
    expect(component.selectedDayConflict()).toMatchObject({
      title: 'Конфликт времени',
      dayKey: '2026-04-23',
      timeLabel: '16:00 - 17:00',
    });
    expect(component.selectedDayConflict()?.reason).toContain('Курс Python - Basic');
    expect(component.selectedDayConflict()?.reason).toContain('Командный созвон');
  });
});
