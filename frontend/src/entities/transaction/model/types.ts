import type { CategoryKind } from '@/entities/category/model/types'

export interface Transaction {
  id: string
  accountId: string
  transferAccountId: string | null
  categoryId: string | null
  kind: CategoryKind | null
  type: 'income' | 'expense' | 'transfer' | 'correction'
  direction: 'inflow' | 'outflow' | 'internal'
  amountMinor: number
  currency: string
  title: string
  occurredAt: string
  createdAt: string
  updatedAt: string
  note: string
}
