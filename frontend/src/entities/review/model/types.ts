export type ReviewStatus = 'pending' | 'completed' | 'skipped'

export interface WeeklyReview {
  id: string
  periodStart: string
  periodEnd: string
  openingBalanceMinor: number
  actualBalanceMinor: number | null
  status: ReviewStatus
  resolvedAt: string | null
}
