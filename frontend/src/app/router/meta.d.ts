import 'vue-router'

declare module 'vue-router' {
  interface RouteMeta {
    guestOnly?: boolean
    navKey?: 'dashboard' | 'accounts' | 'transactions' | 'savings' | 'review' | 'settings'
    requiresAuth?: boolean
    title?: string
  }
}
