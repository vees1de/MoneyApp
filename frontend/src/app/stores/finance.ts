import { computed, ref, watch } from 'vue'
import { defineStore } from 'pinia'

import type { Account, AccountType } from '@/entities/account/model/types'
import type { Category, CategoryKind } from '@/entities/category/model/types'
import type { Transaction } from '@/entities/transaction/model/types'
import { mapApiAccount, mapApiCategory, mapApiTransaction } from '@/shared/api/mappers'
import {
  createAccount,
  createCategory,
  createTransaction,
  deleteCategory,
  listAccounts,
  listCategories,
  listTransactions,
  updateAccount,
} from '@/shared/api/services/finance'
import { minorToMoneyString } from '@/shared/lib/money'
import { readStorage, writeStorage } from '@/shared/lib/storage'

const FILTER_STORAGE_KEY = 'plos-finance-filters'

interface FinanceFilters {
  selectedAccountId: string | null
  transactionKind: CategoryKind | 'all'
}

const defaultFilters: FinanceFilters = {
  selectedAccountId: null,
  transactionKind: 'all',
}

function mapAccountTypeToApi(type: AccountType) {
  switch (type) {
    case 'cash':
      return 'cash'
    case 'savings':
      return 'savings'
    default:
      return 'bank_card'
  }
}

export const useFinanceStore = defineStore('finance', () => {
  const accounts = ref<Account[]>([])
  const categories = ref<Category[]>([])
  const transactions = ref<Transaction[]>([])
  const filters = ref<FinanceFilters>({ ...defaultFilters })
  const loading = ref(false)
  const hydrated = ref(false)

  function bootstrap() {
    filters.value = readStorage<FinanceFilters>(FILTER_STORAGE_KEY, { ...defaultFilters })
    hydrated.value = true
  }

  const sortedTransactions = computed(() =>
    [...transactions.value].sort((left, right) => {
      return new Date(right.occurredAt).getTime() - new Date(left.occurredAt).getTime()
    }),
  )

  const filteredTransactions = computed(() =>
    sortedTransactions.value.filter((transaction) => {
      const matchesAccount =
        !filters.value.selectedAccountId || transaction.accountId === filters.value.selectedAccountId
      const matchesKind =
        filters.value.transactionKind === 'all' || transaction.kind === filters.value.transactionKind

      return matchesAccount && matchesKind
    }),
  )

  const expenseCategories = computed(() =>
    categories.value.filter((category) => category.kind === 'expense'),
  )

  const incomeCategories = computed(() =>
    categories.value.filter((category) => category.kind === 'income'),
  )

  const totalBalanceMinor = computed(() =>
    accounts.value.reduce((sum, account) => sum + account.balanceMinor, 0),
  )

  async function fetchAccounts() {
    const response = await listAccounts()
    accounts.value = response.items.filter((item) => !item.is_archived).map(mapApiAccount)

    if (!filters.value.selectedAccountId && accounts.value[0]) {
      filters.value.selectedAccountId = accounts.value[0].id
    }

    return accounts.value
  }

  async function fetchCategories() {
    const response = await listCategories()
    categories.value = response.items.filter((item) => !item.is_archived).map(mapApiCategory)
    return categories.value
  }

  async function fetchTransactions(overrides?: {
    accountId?: string | null
    type?: CategoryKind | 'all'
  }) {
    const response = await listTransactions({
      accountId: overrides?.accountId !== undefined ? overrides.accountId : filters.value.selectedAccountId,
      type: overrides?.type ?? filters.value.transactionKind,
    })
    transactions.value = response.items.map(mapApiTransaction)
    return transactions.value
  }

  async function hydrateFinance() {
    loading.value = true

    try {
      await Promise.all([fetchAccounts(), fetchCategories(), fetchTransactions({ accountId: null, type: 'all' })])
    } finally {
      loading.value = false
    }
  }

  async function addAccount(input: {
    name: string
    type: AccountType
    balanceMinor: number
    currency: string
  }) {
    const account = await createAccount({
      name: input.name.trim(),
      kind: mapAccountTypeToApi(input.type),
      currency: input.currency,
      openingBalance: minorToMoneyString(input.balanceMinor),
    })

    const mapped = mapApiAccount(account)
    accounts.value = [...accounts.value, mapped]

    if (!filters.value.selectedAccountId) {
      filters.value.selectedAccountId = mapped.id
    }

    return mapped
  }

  async function updateAccountName(input: { id: string; name: string }) {
    const account = await updateAccount(input.id, {
      name: input.name.trim(),
    })

    const mapped = mapApiAccount(account)
    accounts.value = accounts.value.map((item) => (item.id === mapped.id ? mapped : item))
    return mapped
  }

  async function addTransaction(input: {
    accountId: string
    amountMinor: number
    categoryId: string
    kind: CategoryKind
    note: string
    occurredAt: string
  }) {
    const account = accounts.value.find((item) => item.id === input.accountId)
    const transaction = await createTransaction({
      accountId: input.accountId,
      amount: minorToMoneyString(input.amountMinor),
      categoryId: input.categoryId,
      type: input.kind,
      currency: account?.currency ?? 'RUB',
      note: input.note,
      occurredAt: new Date(input.occurredAt).toISOString(),
    })

    const mapped = mapApiTransaction(transaction)
    transactions.value = [mapped, ...transactions.value]
    await fetchAccounts()
    return mapped
  }

  async function addCategory(input: { name: string; kind: CategoryKind; color: string }) {
    const category = await createCategory({
      color: input.color,
      kind: input.kind,
      name: input.name,
    })

    const mapped = mapApiCategory(category)
    categories.value = [mapped, ...categories.value]
    return mapped
  }

  async function removeCategory(categoryId: string) {
    await deleteCategory(categoryId)
    categories.value = categories.value.filter((category) => category.id !== categoryId)
  }

  function setSelectedAccount(accountId: string | null) {
    filters.value.selectedAccountId = accountId
  }

  function setTransactionKindFilter(kind: CategoryKind | 'all') {
    filters.value.transactionKind = kind
  }

  function reset() {
    accounts.value = []
    categories.value = []
    transactions.value = []
    filters.value = { ...defaultFilters }
  }

  watch(
    filters,
    (nextFilters) => {
      if (hydrated.value) {
        writeStorage(FILTER_STORAGE_KEY, nextFilters)
      }
    },
    { deep: true },
  )

  return {
    accounts,
    addAccount,
    addCategory,
    addTransaction,
    bootstrap,
    categories,
    expenseCategories,
    fetchAccounts,
    fetchCategories,
    fetchTransactions,
    filteredTransactions,
    filters,
    hydrateFinance,
    incomeCategories,
    loading,
    removeCategory,
    reset,
    setSelectedAccount,
    setTransactionKindFilter,
    sortedTransactions,
    totalBalanceMinor,
    transactions,
    updateAccountName,
  }
})
