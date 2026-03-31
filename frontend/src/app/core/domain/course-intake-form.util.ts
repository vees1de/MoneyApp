import { toIsoFromDatetimeLocal } from './date-input.util';

export function normalizeText(value: string | null | undefined): string | undefined {
  const trimmed = value?.trim();
  return trimmed ? trimmed : undefined;
}

export function toPositiveNumber(value: string | number | null | undefined): number | undefined {
  if (value === null || value === undefined || value === '') {
    return undefined;
  }

  const parsed = Number(value);
  return Number.isFinite(parsed) && parsed > 0 ? parsed : undefined;
}

export function resolveEndDateFromWeeks(
  startDate: string | null | undefined,
  durationWeeks: number | null | undefined,
): string | undefined {
  if (!startDate || !durationWeeks || durationWeeks <= 0) {
    return undefined;
  }

  const date = new Date(`${startDate}T00:00:00`);
  if (Number.isNaN(date.getTime())) {
    return undefined;
  }

  date.setDate(date.getDate() + durationWeeks * 7);
  return toDateInputValue(date);
}

export function defaultDeadlineFromStartDate(
  startDate: string | null | undefined,
): string | undefined {
  if (!startDate) {
    return undefined;
  }

  const date = new Date(`${startDate}T00:00:00`);
  if (Number.isNaN(date.getTime())) {
    return undefined;
  }

  date.setDate(date.getDate() - 3);
  return `${toDateInputValue(date)}T23:59`;
}

export function buildIntakeSchedulePayload(input: {
  startDate: string | null | undefined;
  endDate: string | null | undefined;
  durationWeeks: number | null | undefined;
  applicationDeadline: string | null | undefined;
}): {
  start_date?: string;
  end_date?: string;
  duration_weeks?: number;
  application_deadline?: string;
} {
  const startDate = normalizeText(input.startDate);
  const durationWeeks = toPositiveNumber(input.durationWeeks);
  const endDate =
    durationWeeks && startDate
      ? resolveEndDateFromWeeks(startDate, durationWeeks)
      : normalizeText(input.endDate);
  const deadline =
    normalizeText(input.applicationDeadline) ?? defaultDeadlineFromStartDate(startDate);

  return {
    start_date: startDate,
    end_date: endDate,
    duration_weeks: durationWeeks,
    application_deadline: toIsoFromDatetimeLocal(deadline),
  };
}

function toDateInputValue(date: Date): string {
  const year = date.getFullYear();
  const month = `${date.getMonth() + 1}`.padStart(2, '0');
  const day = `${date.getDate()}`.padStart(2, '0');

  return `${year}-${month}-${day}`;
}
