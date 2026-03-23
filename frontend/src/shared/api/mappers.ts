import type { Account } from '@/entities/account/model/types'
import type { Category } from '@/entities/category/model/types'
import type { TopCategorySummary } from '@/entities/dashboard/model/types'
import type { SavingsGoal } from '@/entities/savings-goal/model/types'
import type { Transaction } from '@/entities/transaction/model/types'
import type { UserProfile } from '@/entities/user/model/types'
import type { WeeklyReview } from '@/entities/review/model/types'
import type {
  ApiAccount,
  ApiCategory,
  ApiFinanceDashboard,
  ApiSavingsGoal,
  ApiUser,
  ApiWeeklyReview,
  ApiTransaction,
} from './contracts'
import { moneyStringToMinor } from '@/shared/lib/money'

function mapAccountKind(kind: ApiAccount['kind']): Account['type'] {
  switch (kind) {
    case 'cash':
      return 'cash'
    case 'savings':
      return 'savings'
    default:
      return 'bank'
  }
}

export function mapApiUser(user: ApiUser): UserProfile {
  return {
    id: user.id,
    fullName: user.display_name ?? user.email ?? 'User',
    handle: user.email ?? '',
    currency: user.base_currency,
    timezone: user.timezone,
    provider: null,
    onboardingCompleted: user.onboarding_completed,
  }
}

export function mapApiAccount(account: ApiAccount): Account {
  return {
    id: account.id,
    name: account.name,
    type: mapAccountKind(account.kind),
    balanceMinor: moneyStringToMinor(account.current_balance),
    currency: account.currency,
    isArchived: account.is_archived,
    isPrimary: false,
    updatedAt: account.updated_at,
  }
}

export function mapApiCategory(category: ApiCategory): Category {
  return {
    id: category.id,
    name: category.name,
    kind: category.kind,
    scope: category.is_system ? 'system' : 'custom',
    color: category.color ?? '#185d43',
  }
}

export function mapApiTransaction(transaction: ApiTransaction): Transaction {
  return {
    id: transaction.id,
    accountId: transaction.account_id,
    transferAccountId: transaction.transfer_account_id ?? null,
    categoryId: transaction.category_id ?? null,
    kind:
      transaction.type === 'income'
        ? 'income'
        : transaction.type === 'expense'
          ? 'expense'
          : null,
    type: transaction.type,
    direction: transaction.direction,
    amountMinor: moneyStringToMinor(transaction.amount),
    currency: transaction.currency,
    title: transaction.title ?? '',
    occurredAt: transaction.occurred_at,
    createdAt: transaction.created_at,
    updatedAt: transaction.updated_at,
    note: transaction.note ?? '',
  }
}

export function mapApiSavingsGoal(goal: ApiSavingsGoal): SavingsGoal {
  return {
    id: goal.id,
    name: goal.title,
    targetMinor: moneyStringToMinor(goal.target_amount),
    savedMinor: moneyStringToMinor(goal.current_amount),
    currency: goal.currency,
    targetDate: goal.target_date ?? null,
    isCompleted: goal.status === 'completed',
  }
}

export function mapApiWeeklyReview(review: ApiWeeklyReview): WeeklyReview {
  return {
    id: review.id,
    periodStart: review.period_start,
    periodEnd: review.period_end,
    expectedBalanceMinor: moneyStringToMinor(review.expected_balance),
    actualBalanceMinor: review.actual_balance ? moneyStringToMinor(review.actual_balance) : null,
    deltaMinor: review.delta ? moneyStringToMinor(review.delta) : null,
    status: review.status,
    resolutionNote: review.resolution_note ?? null,
    resolvedAt: review.completed_at ?? null,
  }
}

export function mapApiTopCategory(category: ApiFinanceDashboard['top_categories'][number]): TopCategorySummary {
  return {
    categoryId: category.category_id ?? category.name,
    label: category.name,
    amountMinor: moneyStringToMinor(category.amount),
  }
}
