import 'vue-router'
import type { MessageKey } from '@/shared/i18n/messages'

declare module 'vue-router' {
  interface RouteMeta {
    guestOnly?: boolean
    navKey?: 'dashboard' | 'accounts' | 'transactions' | 'savings' | 'review' | 'settings'
    requiresAuth?: boolean
    titleKey?: MessageKey
  }
}
