import type { Pinia } from 'pinia'
import type { Router } from 'vue-router'

import { useAuthStore } from '@/app/stores/auth'
import { useUserStore } from '@/app/stores/user'

export function resolveHomeRoute(pinia: Pinia) {
  const userStore = useUserStore(pinia)
  return userStore.profile.onboardingCompleted ? { name: 'dashboard' } : { name: 'onboarding' }
}

export function applyNavigationGuards(router: Router, pinia: Pinia) {
  router.beforeEach((to) => {
    const authStore = useAuthStore(pinia)
    const userStore = useUserStore(pinia)

    if (!authStore.bootstrapped) {
      authStore.bootstrap()
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
    const pageTitle = to.meta.title ? `${to.meta.title} | Personal Life OS` : 'Personal Life OS'
    document.title = pageTitle
  })
}
