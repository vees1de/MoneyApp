import type { YougileConnection, YougileTask } from './yougile.model';

export interface YougileTaskWindow {
  start: Date;
  end: Date;
  allDay: boolean;
}

export function pickPreferredYougileConnection(
  connections: YougileConnection[],
): YougileConnection | null {
  const active = connections.find((item) => item.status === 'active');
  if (active) {
    return active;
  }

  return connections.find((item) => item.status !== 'revoked') ?? null;
}

export function resolveYougileTaskWindow(task: YougileTask): YougileTaskWindow | null {
  const deadlineValue = task.deadline?.deadline ?? task.deadlineAt ?? null;
  if (!deadlineValue) {
    return null;
  }

  const end = parseDate(deadlineValue);
  if (!end) {
    return null;
  }

  if (task.deadline?.withTime) {
    const start = task.deadline.startDate ? parseDate(task.deadline.startDate) : null;
    if (start) {
      return { start, end, allDay: false };
    }

    return {
      start: new Date(end.getTime() - 60 * 60 * 1000),
      end,
      allDay: false,
    };
  }

  return {
    start: startOfDay(end),
    end: endOfDay(end),
    allDay: true,
  };
}

function parseDate(value: string | null | undefined): Date | null {
  if (!value) {
    return null;
  }

  const parsed = new Date(value);
  return Number.isNaN(parsed.getTime()) ? null : parsed;
}

function startOfDay(value: Date): Date {
  const copy = new Date(value);
  copy.setHours(0, 0, 0, 0);
  return copy;
}

function endOfDay(value: Date): Date {
  const copy = new Date(value);
  copy.setHours(23, 59, 59, 999);
  return copy;
}
