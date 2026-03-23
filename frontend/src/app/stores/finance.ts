import { computed, ref, watch } from 'vue'
import { defineStore } from 'pinia'

import { useAppUiStore } from '@/app/stores/app-ui'
import type { Account, AccountType } from '@/entities/account/model/types'
import type { Category, CategoryKind } from '@/entities/category/model/types'
import type { Transaction } from '@/entities/transaction/model/types'
import { mapApiAccount, mapApiCategory, mapApiTransaction } from '@/shared/api/mappers'
import {
  createAccount,
  createCategory,
  createTransaction,
  deleteCategory,
  deleteTransaction,
  getTransaction,
  listAccounts,
  listCategories,
  listTransactions,
  updateAccount,
  updateTransaction,
} from '@/shared/api/services/finance'
import { translate } from '@/shared/i18n'
import { createId } from '@/shared/lib/id'
import { minorToMoneyString } from '@/shared/lib/money'
import { readStorage, writeStorage } from '@/shared/lib/storage'

const FILTER_STORAGE_KEY = 'plos-finance-filters'
const SAVED_FILTERS_STORAGE_KEY = 'plos-finance-saved-filters'
const DELETE_UNDO_MS = 4200

type TransactionFilterType = CategoryKind | 'transfer' | 'all'

interface FinanceFilters {
  searchQuery: string
  selectedAccountId: string | null
  transactionKind: TransactionFilterType
}

interface SavedTransactionFilter {
  id: string
  label: string
  searchQuery: string
  selectedAccountId: string | null
  transactionKind: TransactionFilterType
}

interface PendingDeletion {
  index: number
  timerId: number
  transaction: Transaction
}

const defaultFilters: FinanceFilters = {
  searchQuery: '',
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

function matchesTransactionType(transaction: Transaction, filter: TransactionFilterType) {
  if (filter === 'all') {
    return true
  }

  return transaction.type === filter
}

function normalizeSearch(value: string) {
  return value.trim().toLowerCase()
}

function insertTransactionAtIndex(items: Transaction[], transaction: Transaction, index: number) {
  const nextItems = [...items]
  nextItems.splice(Math.min(index, nextItems.length), 0, transaction)
  return nextItems
}

export const useFinanceStore = defineStore('finance', () => {
  const appUiStore = useAppUiStore()

  const accounts = ref<Account[]>([])
  const accountDirectory = ref<Record<string, Account>>({})
  const categories = ref<Category[]>([])
  const transactions = ref<Transaction[]>([])
  const transactionDetails = ref<Record<string, Transaction>>({})
  const filters = ref<FinanceFilters>({ ...defaultFilters })
  const savedFilters = ref<SavedTransactionFilter[]>([])
  const pendingDeletionIds = ref<string[]>([])
  const loading = ref(false)
  const detailLoading = ref(false)
  const hydrated = ref(false)

  const deletionQueue = new Map<string, PendingDeletion>()

  function bootstrap() {
    filters.value = readStorage<FinanceFilters>(FILTER_STORAGE_KEY, { ...defaultFilters })
    savedFilters.value = readStorage<SavedTransactionFilter[]>(SAVED_FILTERS_STORAGE_KEY, [])
    hydrated.value = true
  }

  function getAccountById(accountId: string | null | undefined) {
    if (!accountId) {
      return null
    }

    return accountDirectory.value[accountId] ?? accounts.value.find((account) => account.id === accountId) ?? null
  }

  function getCategoryById(categoryId: string | null | undefined) {
    if (!categoryId) {
      return null
    }

    return categories.value.find((category) => category.id === categoryId) ?? null
  }

  function upsertAccount(mapped: Account) {
    accountDirectory.value = {
      ...accountDirectory.value,
      [mapped.id]: mapped,
    }

    if (mapped.isArchived) {
      accounts.value = accounts.value.filter((account) => account.id !== mapped.id)
      return
    }

    const existingIndex = accounts.value.findIndex((account) => account.id === mapped.id)
    if (existingIndex === -1) {
      accounts.value = [...accounts.value, mapped]
      return
    }

    const nextAccounts = [...accounts.value]
    nextAccounts.splice(existingIndex, 1, mapped)
    accounts.value = nextAccounts
  }

  function upsertTransactionRecord(mapped: Transaction) {
    transactionDetails.value = {
      ...transactionDetails.value,
      [mapped.id]: mapped,
    }

    const existingIndex = transactions.value.findIndex((transaction) => transaction.id === mapped.id)
    if (existingIndex === -1) {
      transactions.value = [mapped, ...transactions.value]
      return
    }

    const nextTransactions = [...transactions.value]
    nextTransactions.splice(existingIndex, 1, mapped)
    transactions.value = nextTransactions
  }

  function removeTransactionRecord(transactionId: string) {
    const index = transactions.value.findIndex((transaction) => transaction.id === transactionId)
    if (index === -1) {
      return null
    }

    const transaction = transactions.value[index]
    if (!transaction) {
      return null
    }

    transactions.value = transactions.value.filter((item) => item.id !== transactionId)

    const nextDetails = { ...transactionDetails.value }
    delete nextDetails[transactionId]
    transactionDetails.value = nextDetails

    return {
      index,
      transaction,
    }
  }

  function restoreTransactionRecord(transaction: Transaction, index: number) {
    transactionDetails.value = {
      ...transactionDetails.value,
      [transaction.id]: transaction,
    }
    transactions.value = insertTransactionAtIndex(transactions.value, transaction, index)
  }

  function buildSavedFilterLabel(source: FinanceFilters) {
    const accountName = source.selectedAccountId
      ? getAccountById(source.selectedAccountId)?.name ?? translate('common.account')
      : translate('common.allAccounts')

    const typeLabel =
      source.transactionKind === 'all'
        ? translate('transactions.kindAll')
        : source.transactionKind === 'transfer'
          ? translate('transactions.kindTransfer')
          : source.transactionKind === 'income'
            ? translate('transactions.kindIncome')
            : translate('transactions.kindExpense')

    const parts = [accountName, typeLabel]
    const query = source.searchQuery.trim()
    if (query) {
      parts.push(query)
    }

    return parts.join(' · ')
  }

  const sortedTransactions = computed(() =>
    [...transactions.value].sort((left, right) => {
      return new Date(right.occurredAt).getTime() - new Date(left.occurredAt).getTime()
    }),
  )

  const filteredTransactions = computed(() => {
    const search = normalizeSearch(filters.value.searchQuery)

    return sortedTransactions.value.filter((transaction) => {
      const matchesAccount =
        !filters.value.selectedAccountId || transaction.accountId === filters.value.selectedAccountId
      const matchesKind = matchesTransactionType(transaction, filters.value.transactionKind)

      if (!matchesAccount || !matchesKind) {
        return false
      }

      if (!search) {
        return true
      }

      const accountName = getAccountById(transaction.accountId)?.name ?? ''
      const transferAccountName = getAccountById(transaction.transferAccountId)?.name ?? ''
      const categoryName = getCategoryById(transaction.categoryId)?.name ?? ''

      return [transaction.title, transaction.note, accountName, transferAccountName, categoryName]
        .join(' ')
        .toLowerCase()
        .includes(search)
    })
  })

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
    const mappedAccounts = response.items.map(mapApiAccount)

    accountDirectory.value = Object.fromEntries(mappedAccounts.map((account) => [account.id, account]))
    accounts.value = mappedAccounts.filter((account) => !account.isArchived)

    if (!accounts.value.some((account) => account.id === filters.value.selectedAccountId)) {
      filters.value.selectedAccountId = accounts.value[0]?.id ?? null
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
    type?: TransactionFilterType
  }) {
    const response = await listTransactions({
      accountId: overrides?.accountId ?? null,
      type: overrides?.type ?? 'all',
    })

    transactions.value = response.items.map(mapApiTransaction)
    transactionDetails.value = Object.fromEntries(
      transactions.value.map((transaction) => [transaction.id, transaction]),
    )

    return transactions.value
  }

  async function fetchTransaction(transactionId: string) {
    detailLoading.value = true

    try {
      const response = await getTransaction(transactionId)
      const mapped = mapApiTransaction(response)
      upsertTransactionRecord(mapped)
      return mapped
    } finally {
      detailLoading.value = false
    }
  }

  async function hydrateFinance() {
    loading.value = true

    try {
      await Promise.all([
        fetchAccounts(),
        fetchCategories(),
        fetchTransactions({
          accountId: null,
          type: 'all',
        }),
      ])
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
    upsertAccount(mapped)

    if (!filters.value.selectedAccountId) {
      filters.value.selectedAccountId = mapped.id
    }

    return mapped
  }

  async function renameAccount(input: { id: string; name: string }) {
    const account = await updateAccount(input.id, {
      name: input.name.trim(),
    })

    const mapped = mapApiAccount(account)
    upsertAccount(mapped)
    return mapped
  }

  async function archiveAccount(accountId: string) {
    const account = await updateAccount(accountId, {
      isArchived: true,
    })

    const mapped = mapApiAccount(account)
    upsertAccount(mapped)

    if (filters.value.selectedAccountId === accountId) {
      filters.value.selectedAccountId = accounts.value[0]?.id ?? null
    }

    return mapped
  }

  async function addTransaction(input: {
    accountId: string
    amountMinor: number
    categoryId: string
    kind: CategoryKind
    note: string
    occurredAt: string
    title: string
  }) {
    const account = getAccountById(input.accountId)
    const transaction = await createTransaction({
      accountId: input.accountId,
      amount: minorToMoneyString(input.amountMinor),
      categoryId: input.categoryId,
      currency: account?.currency ?? 'RUB',
      direction: input.kind === 'income' ? 'inflow' : 'outflow',
      note: input.note,
      occurredAt: new Date(input.occurredAt).toISOString(),
      title: input.title.trim(),
      type: input.kind,
    })

    const mapped = mapApiTransaction(transaction)
    upsertTransactionRecord(mapped)
    await fetchAccounts()
    return mapped
  }

  async function addTransfer(input: {
    amountMinor: number
    fromAccountId: string
    note: string
    occurredAt: string
    title: string
    toAccountId: string
  }) {
    const fromAccount = getAccountById(input.fromAccountId)
    const transaction = await createTransaction({
      accountId: input.fromAccountId,
      amount: minorToMoneyString(input.amountMinor),
      currency: fromAccount?.currency ?? 'RUB',
      direction: 'internal',
      note: input.note,
      occurredAt: new Date(input.occurredAt).toISOString(),
      title: input.title.trim(),
      transferAccountId: input.toAccountId,
      type: 'transfer',
    })

    const mapped = mapApiTransaction(transaction)
    upsertTransactionRecord(mapped)
    await fetchAccounts()
    return mapped
  }

  async function updateTransactionEntry(input: {
    id: string
    accountId: string
    amountMinor: number
    categoryId: string | null
    note: string
    occurredAt: string
    title: string
    transferAccountId?: string | null
    type: 'income' | 'expense' | 'transfer'
  }) {
    const account = getAccountById(input.accountId)
    const transaction = await updateTransaction(input.id, {
      accountId: input.accountId,
      amount: minorToMoneyString(input.amountMinor),
      categoryId: input.type === 'transfer' ? null : input.categoryId,
      currency: account?.currency ?? 'RUB',
      direction:
        input.type === 'transfer'
          ? 'internal'
          : input.type === 'income'
            ? 'inflow'
            : 'outflow',
      note: input.note,
      occurredAt: new Date(input.occurredAt).toISOString(),
      title: input.title.trim(),
      transferAccountId: input.type === 'transfer' ? input.transferAccountId ?? null : null,
      type: input.type,
    })

    const mapped = mapApiTransaction(transaction)
    upsertTransactionRecord(mapped)
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

  function getRecentCategories(kind: CategoryKind, limit = 5) {
    const seen = new Set<string>()
    const nextCategories: Category[] = []

    for (const transaction of sortedTransactions.value) {
      if (transaction.type !== kind || !transaction.categoryId || seen.has(transaction.categoryId)) {
        continue
      }

      const category = getCategoryById(transaction.categoryId)
      if (!category) {
        continue
      }

      seen.add(category.id)
      nextCategories.push(category)

      if (nextCategories.length >= limit) {
        break
      }
    }

    return nextCategories
  }

  function setSelectedAccount(accountId: string | null) {
    filters.value.selectedAccountId = accountId
  }

  function setTransactionKindFilter(kind: TransactionFilterType) {
    filters.value.transactionKind = kind
  }

  function setSearchQuery(searchQuery: string) {
    filters.value.searchQuery = searchQuery
  }

  function saveCurrentFilter(label?: string) {
    const savedFilter = {
      id: createId('filter'),
      label: label?.trim() || buildSavedFilterLabel(filters.value),
      searchQuery: filters.value.searchQuery,
      selectedAccountId: filters.value.selectedAccountId,
      transactionKind: filters.value.transactionKind,
    } satisfies SavedTransactionFilter

    savedFilters.value = [savedFilter, ...savedFilters.value.filter((item) => item.label !== savedFilter.label)].slice(0, 6)
    return savedFilter
  }

  function applySavedFilter(filterId: string) {
    const savedFilter = savedFilters.value.find((item) => item.id === filterId)
    if (!savedFilter) {
      return
    }

    filters.value = {
      searchQuery: savedFilter.searchQuery,
      selectedAccountId: savedFilter.selectedAccountId,
      transactionKind: savedFilter.transactionKind,
    }
  }

  function removeSavedFilter(filterId: string) {
    savedFilters.value = savedFilters.value.filter((item) => item.id !== filterId)
  }

  function clearSearch() {
    filters.value.searchQuery = ''
  }

  function queueTransactionDeletion(transactionId: string) {
    if (deletionQueue.has(transactionId)) {
      return
    }

    const removed = removeTransactionRecord(transactionId)
    if (!removed) {
      return
    }

    pendingDeletionIds.value = [...pendingDeletionIds.value, transactionId]

    const timerId = window.setTimeout(async () => {
      deletionQueue.delete(transactionId)
      pendingDeletionIds.value = pendingDeletionIds.value.filter((id) => id !== transactionId)

      try {
        await deleteTransaction(transactionId)
        await fetchAccounts()
      } catch (error) {
        restoreTransactionRecord(removed.transaction, removed.index)
        pendingDeletionIds.value = pendingDeletionIds.value.filter((id) => id !== transactionId)
        appUiStore.pushToast(
          error instanceof Error ? error.message : translate('transactions.deleteFailed'),
          'warning',
        )
      }
    }, DELETE_UNDO_MS)

    deletionQueue.set(transactionId, {
      index: removed.index,
      timerId,
      transaction: removed.transaction,
    })

    appUiStore.pushToast(translate('transactions.deleteQueued'), 'warning', {
      actionLabel: translate('common.undo'),
      duration: DELETE_UNDO_MS,
      onAction: () => cancelPendingTransactionDeletion(transactionId),
    })
  }

  function cancelPendingTransactionDeletion(transactionId: string) {
    const pendingDeletion = deletionQueue.get(transactionId)
    if (!pendingDeletion) {
      return
    }

    window.clearTimeout(pendingDeletion.timerId)
    deletionQueue.delete(transactionId)
    pendingDeletionIds.value = pendingDeletionIds.value.filter((id) => id !== transactionId)
    restoreTransactionRecord(pendingDeletion.transaction, pendingDeletion.index)
    appUiStore.pushToast(translate('transactions.restored'), 'success')
  }

  function reset() {
    for (const pendingDeletion of deletionQueue.values()) {
      window.clearTimeout(pendingDeletion.timerId)
    }

    deletionQueue.clear()
    accounts.value = []
    accountDirectory.value = {}
    categories.value = []
    transactions.value = []
    transactionDetails.value = {}
    pendingDeletionIds.value = []
    filters.value = { ...defaultFilters }
    savedFilters.value = []
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

  watch(
    savedFilters,
    (nextFilters) => {
      if (hydrated.value) {
        writeStorage(SAVED_FILTERS_STORAGE_KEY, nextFilters)
      }
    },
    { deep: true },
  )

  return {
    accounts,
    accountDirectory,
    addAccount,
    addCategory,
    addTransaction,
    addTransfer,
    applySavedFilter,
    archiveAccount,
    bootstrap,
    cancelPendingTransactionDeletion,
    categories,
    clearSearch,
    detailLoading,
    expenseCategories,
    fetchAccounts,
    fetchCategories,
    fetchTransaction,
    fetchTransactions,
    filteredTransactions,
    filters,
    getAccountById,
    getCategoryById,
    getRecentCategories,
    hydrateFinance,
    incomeCategories,
    loading,
    pendingDeletionIds,
    queueTransactionDeletion,
    removeCategory,
    removeSavedFilter,
    renameAccount,
    reset,
    saveCurrentFilter,
    savedFilters,
    setSearchQuery,
    setSelectedAccount,
    setTransactionKindFilter,
    sortedTransactions,
    totalBalanceMinor,
    transactionDetails,
    transactions,
    updateTransactionEntry,
  }
})
