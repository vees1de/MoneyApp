import { computed, ref, watch } from 'vue'
import { defineStore } from 'pinia'

import type { Account, AccountType } from '@/entities/account/model/types'
import type { Category, CategoryKind } from '@/entities/category/model/types'
import type { Transaction } from '@/entities/transaction/model/types'
import { createId } from '@/shared/lib/id'
import { demoAccounts, demoCategories, demoTransactions } from '@/shared/mocks/demo'
import { readStorage, writeStorage } from '@/shared/lib/storage'

const ACCOUNTS_STORAGE_KEY = 'plos-finance-accounts'
const CATEGORIES_STORAGE_KEY = 'plos-finance-categories'
const TRANSACTIONS_STORAGE_KEY = 'plos-finance-transactions'
const FILTER_STORAGE_KEY = 'plos-finance-filters'

interface FinanceFilters {
  selectedAccountId: string | null
  transactionKind: CategoryKind | 'all'
}

function cloneAccounts() {
  return demoAccounts.map((account) => ({ ...account }))
}

function cloneCategories() {
  return demoCategories.map((category) => ({ ...category }))
}

function cloneTransactions() {
  return demoTransactions.map((transaction) => ({ ...transaction }))
}

const defaultFilters: FinanceFilters = {
  selectedAccountId: demoAccounts[0]?.id ?? null,
  transactionKind: 'all',
}

export const useFinanceStore = defineStore('finance', () => {
  const accounts = ref<Account[]>(cloneAccounts())
  const categories = ref<Category[]>(cloneCategories())
  const transactions = ref<Transaction[]>(cloneTransactions())
  const filters = ref<FinanceFilters>({ ...defaultFilters })
  const hydrated = ref(false)

  function bootstrap() {
    accounts.value = readStorage<Account[]>(ACCOUNTS_STORAGE_KEY, cloneAccounts())
    categories.value = readStorage<Category[]>(CATEGORIES_STORAGE_KEY, cloneCategories())
    transactions.value = readStorage<Transaction[]>(TRANSACTIONS_STORAGE_KEY, cloneTransactions())
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

  function addAccount(input: {
    name: string
    type: AccountType
    balanceMinor: number
    currency: string
  }) {
    const account: Account = {
      id: createId('acc'),
      name: input.name.trim(),
      type: input.type,
      balanceMinor: input.balanceMinor,
      currency: input.currency,
      isPrimary: accounts.value.length === 0,
      updatedAt: new Date().toISOString(),
    }

    accounts.value = [...accounts.value, account]

    if (!filters.value.selectedAccountId) {
      filters.value.selectedAccountId = account.id
    }

    return account
  }

  function updateAccount(input: { id: string; name: string }) {
    const account = accounts.value.find((item) => item.id === input.id)

    if (!account) {
      return
    }

    account.name = input.name.trim()
    account.updatedAt = new Date().toISOString()
  }

  function addTransaction(input: {
    accountId: string
    amountMinor: number
    categoryId: string
    kind: CategoryKind
    note: string
    occurredAt: string
  }) {
    const transaction: Transaction = {
      id: createId('txn'),
      accountId: input.accountId,
      amountMinor: input.amountMinor,
      categoryId: input.categoryId,
      kind: input.kind,
      note: input.note.trim(),
      occurredAt: input.occurredAt,
      currency: 'RUB',
    }

    transactions.value = [transaction, ...transactions.value]

    const account = accounts.value.find((item) => item.id === input.accountId)

    if (account) {
      account.balanceMinor += input.kind === 'income' ? input.amountMinor : -input.amountMinor
      account.updatedAt = new Date().toISOString()
    }

    return transaction
  }

  function addCategory(input: { name: string; kind: CategoryKind; color: string }) {
    const category: Category = {
      id: createId('cat'),
      name: input.name.trim(),
      kind: input.kind,
      scope: 'custom',
      color: input.color,
    }

    categories.value = [category, ...categories.value]

    return category
  }

  function deleteCategory(categoryId: string) {
    const hasTransactions = transactions.value.some((transaction) => transaction.categoryId === categoryId)

    if (hasTransactions) {
      return false
    }

    categories.value = categories.value.filter((category) => category.id !== categoryId)
    return true
  }

  function setSelectedAccount(accountId: string | null) {
    filters.value.selectedAccountId = accountId
  }

  function setTransactionKindFilter(kind: CategoryKind | 'all') {
    filters.value.transactionKind = kind
  }

  function reset() {
    accounts.value = cloneAccounts()
    categories.value = cloneCategories()
    transactions.value = cloneTransactions()
    filters.value = { ...defaultFilters }
  }

  watch(
    accounts,
    (nextAccounts) => {
      if (hydrated.value) {
        writeStorage(ACCOUNTS_STORAGE_KEY, nextAccounts)
      }
    },
    { deep: true },
  )
  watch(
    categories,
    (nextCategories) => {
      if (hydrated.value) {
        writeStorage(CATEGORIES_STORAGE_KEY, nextCategories)
      }
    },
    { deep: true },
  )
  watch(
    transactions,
    (nextTransactions) => {
      if (hydrated.value) {
        writeStorage(TRANSACTIONS_STORAGE_KEY, nextTransactions)
      }
    },
    { deep: true },
  )
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
    deleteCategory,
    expenseCategories,
    filteredTransactions,
    filters,
    incomeCategories,
    reset,
    setSelectedAccount,
    setTransactionKindFilter,
    sortedTransactions,
    totalBalanceMinor,
    transactions,
    updateAccount,
  }
})
