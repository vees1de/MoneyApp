import type { Pinia } from 'pinia'

import { useAppUiStore } from '@/app/stores/app-ui'
import { useAuthStore } from '@/app/stores/auth'
import { useFinanceStore } from '@/app/stores/finance'
import { useReviewStore } from '@/app/stores/review'
import { useSavingsStore } from '@/app/stores/savings'
import { useUserStore } from '@/app/stores/user'

export async function bootstrapApp(pinia: Pinia) {
  useUserStore(pinia).bootstrap()
  useAuthStore(pinia).bootstrap()
  useFinanceStore(pinia).bootstrap()
  useSavingsStore(pinia).bootstrap()
  useReviewStore(pinia).bootstrap()
  useAppUiStore(pinia).bootstrap()
}
