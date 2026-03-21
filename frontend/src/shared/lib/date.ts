const MS_IN_DAY = 24 * 60 * 60 * 1000

export function toIsoDate(value: Date | string) {
  const date = typeof value === 'string' ? new Date(value) : value
  return date.toISOString().slice(0, 10)
}

export function startOfCurrentWeek() {
  const now = new Date()
  const day = now.getDay()
  const mondayOffset = day === 0 ? -6 : 1 - day
  const date = new Date(now)

  date.setHours(0, 0, 0, 0)
  date.setDate(now.getDate() + mondayOffset)

  return date
}

export function endOfCurrentWeek() {
  const date = new Date(startOfCurrentWeek())

  date.setDate(date.getDate() + 6)
  date.setHours(23, 59, 59, 999)

  return date
}

export function isDateWithinRange(value: string, start: string, end: string) {
  const target = new Date(value).getTime()
  return target >= new Date(start).getTime() && target <= new Date(end).getTime()
}

export function addDays(value: Date, days: number) {
  return new Date(value.getTime() + days * MS_IN_DAY)
}

export function formatDate(value: string, locale = 'ru-RU') {
  return new Intl.DateTimeFormat(locale, {
    day: 'numeric',
    month: 'short',
  }).format(new Date(value))
}

export function formatDateTime(value: string, locale = 'ru-RU') {
  return new Intl.DateTimeFormat(locale, {
    day: 'numeric',
    month: 'short',
    hour: '2-digit',
    minute: '2-digit',
  }).format(new Date(value))
}

export function formatWeekRange(start: string, end: string, locale = 'ru-RU') {
  return `${formatDate(start, locale)} - ${formatDate(end, locale)}`
}
