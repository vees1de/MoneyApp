import type {
  ApiAccount,
  ApiAccountsResponse,
  ApiCategoriesResponse,
  ApiCategory,
  ApiTransaction,
  ApiTransactionsResponse,
} from '@/shared/api/contracts'
import { apiRequest } from '@/shared/api/http/client'

export function listAccounts() {
  return apiRequest<ApiAccountsResponse>('/accounts')
}

export function createAccount(input: {
  currency: string
  kind: 'cash' | 'bank_card' | 'bank_account' | 'savings' | 'virtual'
  name: string
  openingBalance: string
}) {
  return apiRequest<ApiAccount>('/accounts', {
    method: 'POST',
    body: JSON.stringify({
      name: input.name,
      kind: input.kind,
      currency: input.currency,
      opening_balance: input.openingBalance,
    }),
  })
}

export function updateAccount(accountId: string, input: { isArchived?: boolean; name?: string }) {
  return apiRequest<ApiAccount>(`/accounts/${accountId}`, {
    method: 'PATCH',
    body: JSON.stringify({
      name: input.name,
      is_archived: input.isArchived,
    }),
  })
}

export function listCategories() {
  return apiRequest<ApiCategoriesResponse>('/finance/categories')
}

export function createCategory(input: {
  color?: string | null
  kind: 'income' | 'expense'
  name: string
}) {
  return apiRequest<ApiCategory>('/finance/categories', {
    method: 'POST',
    body: JSON.stringify({
      kind: input.kind,
      name: input.name,
      color: input.color ?? null,
    }),
  })
}

export function deleteCategory(categoryId: string) {
  return apiRequest<void>(`/finance/categories/${categoryId}`, {
    method: 'DELETE',
  })
}

export function listTransactions(filters: {
  accountId?: string | null
  type?: 'all' | 'income' | 'expense' | 'transfer'
}) {
  return apiRequest<ApiTransactionsResponse>('/finance/transactions', {
    query: {
      account: filters.accountId ?? undefined,
      type: filters.type && filters.type !== 'all' ? filters.type : undefined,
    },
  })
}

export function createTransaction(input: {
  accountId: string
  amount: string
  categoryId?: string | null
  currency: string
  direction: 'inflow' | 'outflow' | 'internal'
  note: string
  occurredAt: string
  title?: string
  transferAccountId?: string | null
  type: 'income' | 'expense' | 'transfer'
}) {
  return apiRequest<ApiTransaction>('/finance/transactions', {
    method: 'POST',
    body: JSON.stringify({
      account_id: input.accountId,
      transfer_account_id: input.transferAccountId ?? null,
      category_id: input.categoryId ?? null,
      type: input.type,
      amount: input.amount,
      currency: input.currency,
      direction: input.direction,
      title: input.title || null,
      note: input.note || null,
      occurred_at: input.occurredAt,
    }),
  })
}

export function getTransaction(transactionId: string) {
  return apiRequest<ApiTransaction>(`/finance/transactions/${transactionId}`)
}

export function updateTransaction(
  transactionId: string,
  input: {
    accountId?: string
    amount?: string
    categoryId?: string | null
    currency?: string
    direction?: 'inflow' | 'outflow' | 'internal'
    note?: string
    occurredAt?: string
    title?: string
    transferAccountId?: string | null
    type?: 'income' | 'expense' | 'transfer'
  },
) {
  return apiRequest<ApiTransaction>(`/finance/transactions/${transactionId}`, {
    method: 'PATCH',
    body: JSON.stringify({
      account_id: input.accountId,
      transfer_account_id: input.transferAccountId,
      category_id: input.categoryId,
      type: input.type,
      amount: input.amount,
      currency: input.currency,
      direction: input.direction,
      title: input.title === undefined ? undefined : input.title || null,
      note: input.note === undefined ? undefined : input.note || null,
      occurred_at: input.occurredAt,
    }),
  })
}

export function deleteTransaction(transactionId: string) {
  return apiRequest<void>(`/finance/transactions/${transactionId}`, {
    method: 'DELETE',
  })
}
