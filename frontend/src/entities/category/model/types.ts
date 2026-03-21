export type CategoryKind = 'income' | 'expense'
export type CategoryScope = 'system' | 'custom'

export interface Category {
  id: string
  name: string
  kind: CategoryKind
  scope: CategoryScope
  color: string
}
