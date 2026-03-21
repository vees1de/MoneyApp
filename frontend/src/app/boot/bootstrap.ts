import type { Pinia } from 'pinia'

import { useAppUiStore } from '@/app/stores/app-ui'
import { useAuthStore } from '@/app/stores/auth'
import { useFinanceStore } from '@/app/stores/finance'
import { useUserStore } from '@/app/stores/user'
import { bootstrapI18n } from '@/shared/i18n'

export async function bootstrapApp(pinia: Pinia) {
  bootstrapI18n()
  useUserStore(pinia).bootstrap()
  await useAuthStore(pinia).bootstrap()
  useFinanceStore(pinia).bootstrap()
  useAppUiStore(pinia).bootstrap()
}
