export interface SavingsGoal {
  id: string
  name: string
  targetMinor: number
  savedMinor: number
  currency: string
  targetDate: string | null
  isCompleted: boolean
}
