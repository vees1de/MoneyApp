export interface ApiErrorPayload {
  error: {
    code: string
    details?: unknown
    message: string
  }
}

export interface ApiUser {
  id: string
  email: string | null
  display_name: string | null
  avatar_url: string | null
  timezone: string
  base_currency: string
  onboarding_completed: boolean
  weekly_review_weekday: number
  weekly_review_hour: number
  created_at: string
  updated_at: string
}

export interface ApiTokens {
  access_token: string
  expires_in: number
  refresh_token: string
}

export interface ApiAuthResponse {
  meta?: Record<string, unknown>
  tokens: ApiTokens
  user: ApiUser
}

export interface ApiMeResponse {
  user: ApiUser
}

export interface ApiAccount {
  id: string
  user_id: string
  name: string
  kind: 'cash' | 'bank_card' | 'bank_account' | 'savings' | 'virtual'
  currency: string
  opening_balance: string
  current_balance: string
  is_archived: boolean
  created_at: string
  updated_at: string
}

export interface ApiAccountsResponse {
  items: ApiAccount[]
}

export interface ApiCategory {
  id: string
  user_id?: string | null
  kind: 'income' | 'expense'
  name: string
  color?: string | null
  icon?: string | null
  parent_id?: string | null
  is_system: boolean
  is_archived: boolean
  created_at: string
  updated_at: string
}

export interface ApiCategoriesResponse {
  items: ApiCategory[]
}

export interface ApiTransaction {
  id: string
  user_id: string
  account_id: string
  transfer_account_id?: string | null
  type: 'income' | 'expense' | 'transfer' | 'correction'
  category_id?: string | null
  amount: string
  currency: string
  direction: 'inflow' | 'outflow' | 'internal'
  title?: string | null
  note?: string | null
  occurred_at: string
  created_at: string
  updated_at: string
}

export interface ApiTransactionsResponse {
  items: ApiTransaction[]
}

export interface ApiSavingsGoal {
  id: string
  user_id: string
  title: string
  target_amount: string
  current_amount: string
  currency: string
  target_date?: string | null
  priority: 'low' | 'medium' | 'high'
  status: 'active' | 'paused' | 'completed' | 'archived'
  created_at: string
  updated_at: string
}

export interface ApiGoalProgress {
  goal: ApiSavingsGoal
  progress_percent: string
  recommended_monthly_contribution: string
}

export interface ApiSavingsSummary {
  total_target: string
  total_current: string
  reserved_this_month: string
  safe_to_spend: string
  goals: ApiGoalProgress[]
}

export interface ApiSavingsGoalsResponse {
  items: ApiSavingsGoal[]
}

export interface ApiWeeklyReview {
  id: string
  user_id: string
  account_id?: string | null
  period_start: string
  period_end: string
  expected_balance: string
  actual_balance?: string | null
  delta?: string | null
  status: 'pending' | 'matched' | 'discrepancy_found' | 'resolved' | 'skipped'
  resolution_note?: string | null
  created_at: string
  completed_at?: string | null
}

export interface ApiTopCategory {
  category_id?: string | null
  name: string
  amount: string
}

export interface ApiFinanceDashboard {
  current_balance: string
  monthly_income: string
  monthly_expense: string
  saved_this_month: string
  safe_to_spend: string
  top_categories: ApiTopCategory[]
  savings: ApiGoalProgress[]
  weekly_review: ApiWeeklyReview
  insights: string[]
}
