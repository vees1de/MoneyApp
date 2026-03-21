export type AccountType = 'cash' | 'bank' | 'savings'

export interface Account {
  id: string
  name: string
  type: AccountType
  balanceMinor: number
  currency: string
  isPrimary: boolean
  updatedAt: string
}
