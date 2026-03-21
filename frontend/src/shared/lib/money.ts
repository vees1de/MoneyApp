import { getLocaleTag } from '@/shared/i18n'

export function formatMoney(amountMinor: number | string, currency = 'RUB', locale = getLocaleTag()) {
  const numericAmount =
    typeof amountMinor === 'string' ? Number(amountMinor.replace(',', '.')) : amountMinor / 100

  return new Intl.NumberFormat(locale, {
    style: 'currency',
    currency,
    maximumFractionDigits: 2,
  }).format(numericAmount)
}

export function parseAmountToMinor(value: string | number) {
  const normalized = typeof value === 'number' ? value : Number(value.toString().replace(',', '.'))
  return Number.isFinite(normalized) ? Math.round(normalized * 100) : 0
}

export function parseAmountToMoneyString(value: string | number) {
  const normalized = typeof value === 'number' ? value : Number(value.toString().replace(',', '.'))
  return Number.isFinite(normalized) ? normalized.toFixed(2) : '0.00'
}

export function minorToMoneyString(amountMinor: number) {
  return (amountMinor / 100).toFixed(2)
}

export function moneyStringToMinor(value: string) {
  return parseAmountToMinor(value)
}

export function formatPercent(value: number, locale = getLocaleTag()) {
  return new Intl.NumberFormat(locale, {
    style: 'percent',
    maximumFractionDigits: 0,
  }).format(value)
}
