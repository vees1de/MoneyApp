export type AuthProvider = 'telegram' | 'yandex'

export interface UserProfile {
  id: string
  fullName: string
  handle: string
  currency: string
  timezone: string
  provider: AuthProvider | null
  onboardingCompleted: boolean
}
