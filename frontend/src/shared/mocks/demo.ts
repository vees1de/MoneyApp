import type { Account } from '@/entities/account/model/types'
import type { Category } from '@/entities/category/model/types'
import type { SavingsGoal } from '@/entities/savings-goal/model/types'
import type { Transaction } from '@/entities/transaction/model/types'
import type { UserProfile } from '@/entities/user/model/types'
import type { WeeklyReview } from '@/entities/review/model/types'
import { addDays, startOfCurrentWeek, toIsoDate } from '@/shared/lib/date'

const weekStart = startOfCurrentWeek()

export const demoProfile: UserProfile = {
  id: 'user-1',
  fullName: 'Alex Petrov',
  handle: '@alex',
  currency: 'RUB',
  timezone: 'Asia/Yakutsk',
  provider: null,
  onboardingCompleted: false,
}

export const demoAccounts: Account[] = [
  {
    id: 'acc-main',
    name: 'Main card',
    type: 'bank',
    balanceMinor: 245_000,
    currency: 'RUB',
    isPrimary: true,
    updatedAt: new Date().toISOString(),
  },
  {
    id: 'acc-cash',
    name: 'Cash',
    type: 'cash',
    balanceMinor: 18_500,
    currency: 'RUB',
    isPrimary: false,
    updatedAt: new Date().toISOString(),
  },
]

export const demoCategories: Category[] = [
  { id: 'cat-salary', name: 'Salary', kind: 'income', scope: 'system', color: '#18744f' },
  { id: 'cat-freelance', name: 'Freelance', kind: 'income', scope: 'custom', color: '#2f855a' },
  { id: 'cat-food', name: 'Food', kind: 'expense', scope: 'system', color: '#c14638' },
  { id: 'cat-home', name: 'Home', kind: 'expense', scope: 'system', color: '#a56916' },
  { id: 'cat-transport', name: 'Transport', kind: 'expense', scope: 'system', color: '#7a5ccf' },
  { id: 'cat-health', name: 'Health', kind: 'expense', scope: 'custom', color: '#b45f8d' },
]

export const demoTransactions: Transaction[] = [
  {
    id: 'txn-1',
    accountId: 'acc-main',
    categoryId: 'cat-salary',
    kind: 'income',
    amountMinor: 180_000,
    currency: 'RUB',
    occurredAt: toIsoDate(addDays(weekStart, 0)),
    note: 'Monthly salary',
  },
  {
    id: 'txn-2',
    accountId: 'acc-main',
    categoryId: 'cat-food',
    kind: 'expense',
    amountMinor: 14_500,
    currency: 'RUB',
    occurredAt: toIsoDate(addDays(weekStart, 1)),
    note: 'Groceries',
  },
  {
    id: 'txn-3',
    accountId: 'acc-main',
    categoryId: 'cat-home',
    kind: 'expense',
    amountMinor: 22_000,
    currency: 'RUB',
    occurredAt: toIsoDate(addDays(weekStart, 2)),
    note: 'Utilities',
  },
  {
    id: 'txn-4',
    accountId: 'acc-cash',
    categoryId: 'cat-transport',
    kind: 'expense',
    amountMinor: 4_200,
    currency: 'RUB',
    occurredAt: toIsoDate(addDays(weekStart, 3)),
    note: 'Taxi',
  },
  {
    id: 'txn-5',
    accountId: 'acc-main',
    categoryId: 'cat-freelance',
    kind: 'income',
    amountMinor: 38_000,
    currency: 'RUB',
    occurredAt: toIsoDate(addDays(weekStart, 4)),
    note: 'Client payment',
  },
]

export const demoSavingsGoals: SavingsGoal[] = [
  {
    id: 'goal-emergency',
    name: 'Emergency fund',
    targetMinor: 300_000,
    savedMinor: 120_000,
    currency: 'RUB',
    targetDate: null,
    isCompleted: false,
  },
  {
    id: 'goal-trip',
    name: 'Summer trip',
    targetMinor: 120_000,
    savedMinor: 48_000,
    currency: 'RUB',
    targetDate: toIsoDate(addDays(new Date(), 120)),
    isCompleted: false,
  },
]

export const demoWeeklyReview: WeeklyReview = {
  id: 'review-current',
  periodStart: toIsoDate(weekStart),
  periodEnd: toIsoDate(addDays(weekStart, 6)),
  openingBalanceMinor: 85_000,
  actualBalanceMinor: null,
  status: 'pending',
  resolvedAt: null,
}
