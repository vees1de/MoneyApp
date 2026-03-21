export function formatMoney(amountMinor: number, currency = 'RUB', locale = 'ru-RU') {
  return new Intl.NumberFormat(locale, {
    style: 'currency',
    currency,
    maximumFractionDigits: 2,
  }).format(amountMinor / 100)
}

export function parseAmountToMinor(value: string | number) {
  const normalized = typeof value === 'number' ? value : Number(value.toString().replace(',', '.'))
  return Number.isFinite(normalized) ? Math.round(normalized * 100) : 0
}

export function formatPercent(value: number, locale = 'ru-RU') {
  return new Intl.NumberFormat(locale, {
    style: 'percent',
    maximumFractionDigits: 0,
  }).format(value)
}
