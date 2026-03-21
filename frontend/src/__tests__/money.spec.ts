import { describe, expect, it } from 'vitest'

import { formatMoney, parseAmountToMinor } from '@/shared/lib/money'

describe('money helpers', () => {
  it('formats minor units into localized money', () => {
    expect(formatMoney(123_45, 'RUB', 'ru-RU')).toContain('123')
  })

  it('parses decimal strings into minor units', () => {
    expect(parseAmountToMinor('12.34')).toBe(1234)
  })
})
