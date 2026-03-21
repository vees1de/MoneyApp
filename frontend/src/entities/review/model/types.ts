export type ReviewStatus = 'pending' | 'matched' | 'discrepancy_found' | 'resolved' | 'skipped'

export interface WeeklyReview {
  id: string
  periodStart: string
  periodEnd: string
  expectedBalanceMinor: number
  actualBalanceMinor: number | null
  deltaMinor: number | null
  resolutionNote: string | null
  status: ReviewStatus
  resolvedAt: string | null
}
