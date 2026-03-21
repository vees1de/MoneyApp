import { ref, watch } from 'vue'
import { defineStore } from 'pinia'

import type { AuthProvider, UserProfile } from '@/entities/user/model/types'
import { demoProfile } from '@/shared/mocks/demo'
import { readStorage, writeStorage } from '@/shared/lib/storage'

const USER_STORAGE_KEY = 'plos-user-profile'

function createDefaultProfile(): UserProfile {
  return { ...demoProfile }
}

export const useUserStore = defineStore('user', () => {
  const profile = ref<UserProfile>(createDefaultProfile())
  const hydrated = ref(false)

  function bootstrap() {
    profile.value = readStorage<UserProfile>(USER_STORAGE_KEY, createDefaultProfile())
    hydrated.value = true
  }

  function setAuthProvider(provider: AuthProvider) {
    profile.value = {
      ...profile.value,
      provider,
      fullName: profile.value.fullName || (provider === 'telegram' ? 'Telegram user' : 'Yandex user'),
      handle: profile.value.handle || (provider === 'telegram' ? '@telegram' : '@yandex'),
    }
  }

  function completeOnboarding(payload: {
    fullName: string
    currency: string
    timezone: string
  }) {
    profile.value = {
      ...profile.value,
      fullName: payload.fullName.trim(),
      currency: payload.currency,
      timezone: payload.timezone,
      onboardingCompleted: true,
    }
  }

  function reset() {
    profile.value = createDefaultProfile()
  }

  watch(
    profile,
    (nextProfile) => {
      if (!hydrated.value) {
        return
      }

      writeStorage(USER_STORAGE_KEY, nextProfile)
    },
    { deep: true },
  )

  return {
    bootstrap,
    completeOnboarding,
    profile,
    reset,
    setAuthProvider,
  }
})
