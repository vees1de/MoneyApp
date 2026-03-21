import type { CategoryKind } from '@/entities/category/model/types'

export interface Transaction {
  id: string
  accountId: string
  categoryId: string
  kind: CategoryKind
  type: 'income' | 'expense' | 'transfer' | 'correction'
  direction: 'inflow' | 'outflow' | 'internal'
  amountMinor: number
  currency: string
  occurredAt: string
  note: string
}
