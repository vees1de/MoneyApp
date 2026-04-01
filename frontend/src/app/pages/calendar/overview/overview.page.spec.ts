import { TestBed } from '@angular/core/testing';
import { provideRouter } from '@angular/router';
import { of } from 'rxjs';

import { IntegrationsApiService } from '@core/api/integrations-api.service';
import { LearningPlanApiService } from '@core/api/learning-plan-api.service';
import type { YougileTask } from '@entities/yougile';

import { CalendarOverviewPageComponent } from './overview.page';

describe('CalendarOverviewPageComponent', () => {
  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [CalendarOverviewPageComponent],
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
        {
          provide: LearningPlanApiService,
          useValue: {
            getMyPlan: () => of(null),
          },
        },
      ],
    }).compileComponents();
  });

  it('turns a Yougile task with startDate and deadline into a timed timeblock', () => {
    const fixture = TestBed.createComponent(CalendarOverviewPageComponent);
    const component = fixture.componentInstance as CalendarOverviewPageComponent & {
      yougileTasks: { set(value: YougileTask[]): void };
      learningPlan: { set(value: null): void };
      calendarEvents: () => Array<{
        id?: string;
        title?: string;
        start?: string;
        end?: string;
        allDay?: boolean;
        backgroundColor?: string;
        borderColor?: string;
        textColor?: string;
        extendedProps?: {
          badge?: string;
          meta?: string;
          source?: string;
        };
      }>;
      unscheduledYougileTasks: () => unknown[];
    };

    const testTimeblock: YougileTask = {
      id: 'task-1',
      deleted: false,
      title: 'Тестовый таймблок',
      timestamp: '2026-04-01T09:49:47.426Z',
      columnId: 'col-1',
      columnTitle: 'В РАБОТЕ',
      boardId: 'board-1',
      boardTitle: 'Новая доска',
      archived: false,
      completed: false,
      createdBy: 'user-1',
      deadline: {
        deadline: '2026-04-02T00:15:00.602Z',
        startDate: '2026-04-01T23:15:00.602Z',
        withTime: true,
      },
      deadlineAt: '2026-04-02T00:15:00.602Z',
      idTaskCommon: 'ID-5',
      idTaskProject: 'DEV-5',
      type: 'task',
    };

    component.yougileTasks.set([testTimeblock]);
    component.learningPlan.set(null);

    const events = component.calendarEvents();
    const timeblock = events.find((event) => event.id === 'yougile-task-1');

    expect(events).toHaveLength(1);
    expect(timeblock).toBeDefined();
    expect(timeblock).toMatchObject({
      id: 'yougile-task-1',
      title: 'Тестовый таймблок',
      start: '2026-04-01T23:15:00.602Z',
      end: '2026-04-02T00:15:00.602Z',
      allDay: false,
      backgroundColor: '#0f766e',
      borderColor: '#0f766e',
      textColor: '#ffffff',
      extendedProps: {
        badge: 'Yougile',
        meta: 'Новая доска · В РАБОТЕ',
        source: 'yougile',
      },
    });
    expect(component.unscheduledYougileTasks()).toEqual([]);
  });
});
