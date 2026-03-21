import { watch } from 'vue'
import type { Pinia } from 'pinia'
import type { Router } from 'vue-router'

import { useAuthStore } from '@/app/stores/auth'
import { useUserStore } from '@/app/stores/user'
import { env } from '@/shared/config/env'
import { currentLocale, translate } from '@/shared/i18n'
import type { MessageKey } from '@/shared/i18n/messages'

export function resolveHomeRoute(pinia: Pinia) {
  const userStore = useUserStore(pinia)
  return userStore.profile.onboardingCompleted ? { name: 'dashboard' } : { name: 'onboarding' }
}

export function applyNavigationGuards(router: Router, pinia: Pinia) {
  const syncDocumentTitle = (titleKey?: MessageKey) => {
    document.title = titleKey ? `${translate(titleKey)} | ${env.appName}` : env.appName
  }

  router.beforeEach(async (to) => {
    const authStore = useAuthStore(pinia)
    const userStore = useUserStore(pinia)

    if (!authStore.bootstrapped) {
      await authStore.bootstrap()
    }

    if (to.meta.requiresAuth && !authStore.isAuthenticated) {
      return { name: 'login' }
    }

    if (to.meta.guestOnly && authStore.isAuthenticated) {
      return resolveHomeRoute(pinia)
    }

    if (
      authStore.isAuthenticated &&
      !userStore.profile.onboardingCompleted &&
      to.name !== 'onboarding'
    ) {
      return { name: 'onboarding' }
    }

    if (authStore.isAuthenticated && userStore.profile.onboardingCompleted && to.name === 'onboarding') {
      return { name: 'dashboard' }
    }

    return true
  })

  router.afterEach((to) => {
    syncDocumentTitle(to.meta.titleKey)
  })

  watch(currentLocale, () => {
    syncDocumentTitle(router.currentRoute.value.meta.titleKey)
  })
}
