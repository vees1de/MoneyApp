import type { Pinia } from 'pinia'
import { createRouter, createWebHistory } from 'vue-router'

import { applyNavigationGuards, resolveHomeRoute } from '@/app/guards/navigation'

export function createAppRouter(pinia: Pinia) {
  const router = createRouter({
    history: createWebHistory(import.meta.env.BASE_URL),
    routes: [
      {
        path: '/',
        redirect: () => resolveHomeRoute(pinia),
      },
      {
        path: '/login',
        component: () => import('@/app/layouts/AuthLayout.vue'),
        children: [
          {
            path: '',
            name: 'login',
            component: () => import('@/pages/auth/LoginPage.vue'),
            meta: {
              guestOnly: true,
              titleKey: 'route.login',
            },
          },
        ],
      },
      {
        path: '/onboarding',
        component: () => import('@/app/layouts/AuthLayout.vue'),
        children: [
          {
            path: '',
            name: 'onboarding',
            component: () => import('@/pages/onboarding/OnboardingPage.vue'),
            meta: {
              requiresAuth: true,
              titleKey: 'route.onboarding',
            },
          },
        ],
      },
      {
        path: '/',
        component: () => import('@/app/layouts/MobileAppLayout.vue'),
        children: [
          {
            path: 'dashboard',
            name: 'dashboard',
            component: () => import('@/pages/dashboard/DashboardPage.vue'),
            meta: {
              navKey: 'dashboard',
              requiresAuth: true,
              titleKey: 'route.dashboard',
            },
          },
          {
            path: 'accounts',
            name: 'accounts',
            component: () => import('@/pages/accounts/AccountsPage.vue'),
            meta: {
              navKey: 'accounts',
              requiresAuth: true,
              titleKey: 'route.accounts',
            },
          },
          {
            path: 'transactions',
            name: 'transactions',
            component: () => import('@/pages/transactions/TransactionsPage.vue'),
            meta: {
              navKey: 'transactions',
              requiresAuth: true,
              titleKey: 'route.transactions',
            },
          },
          {
            path: 'transactions/new',
            name: 'transaction-new',
            component: () => import('@/pages/transactions/TransactionEditorPage.vue'),
            meta: {
              requiresAuth: true,
              titleKey: 'route.transactionNew',
            },
          },
          {
            path: 'transactions/transfer',
            name: 'transaction-transfer',
            component: () => import('@/pages/transactions/TransferPage.vue'),
            meta: {
              requiresAuth: true,
              titleKey: 'route.transfer',
            },
          },
          {
            path: 'transactions/:id',
            name: 'transaction-edit',
            component: () => import('@/pages/transactions/TransactionEditorPage.vue'),
            meta: {
              requiresAuth: true,
              titleKey: 'route.transactionEdit',
            },
          },
          {
            path: 'categories',
            name: 'categories',
            component: () => import('@/pages/categories/CategoriesPage.vue'),
            meta: {
              requiresAuth: true,
              titleKey: 'route.categories',
            },
          },
          {
            path: 'savings',
            name: 'savings',
            component: () => import('@/pages/savings/SavingsPage.vue'),
            meta: {
              navKey: 'savings',
              requiresAuth: true,
              titleKey: 'route.savings',
            },
          },
          {
            path: 'review',
            name: 'review',
            component: () => import('@/pages/review/ReviewPage.vue'),
            meta: {
              navKey: 'review',
              requiresAuth: true,
              titleKey: 'route.review',
            },
          },
          {
            path: 'settings',
            name: 'settings',
            component: () => import('@/pages/settings/SettingsPage.vue'),
            meta: {
              navKey: 'settings',
              requiresAuth: true,
              titleKey: 'route.settings',
            },
          },
        ],
      },
      {
        path: '/:pathMatch(.*)*',
        name: 'not-found',
        component: () => import('@/pages/common/NotFoundPage.vue'),
        meta: {
          titleKey: 'route.notFound',
        },
      },
    ],
  })

  applyNavigationGuards(router, pinia)

  return router
}
