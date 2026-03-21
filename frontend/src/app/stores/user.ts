import { ref } from 'vue'
import { defineStore } from 'pinia'

import type { UserProfile } from '@/entities/user/model/types'
import { mapApiUser } from '@/shared/api/mappers'
import { fetchMe } from '@/shared/api/services/auth'
import { updateUserPreferences } from '@/shared/api/services/users'

function createEmptyProfile(): UserProfile {
  return {
    id: '',
    fullName: '',
    handle: '',
    currency: 'RUB',
    timezone: 'Asia/Yakutsk',
    provider: null,
    onboardingCompleted: false,
  }
}

export const useUserStore = defineStore('user', () => {
  const profile = ref<UserProfile>(createEmptyProfile())
  const loading = ref(false)

  function bootstrap() {
    profile.value = createEmptyProfile()
  }

  function setProfile(nextProfile: UserProfile) {
    profile.value = nextProfile
  }

  async function fetchProfile() {
    loading.value = true
    try {
      const response = await fetchMe()
      profile.value = {
        ...mapApiUser(response.user),
        provider: profile.value.provider,
      }
      return profile.value
    } finally {
      loading.value = false
    }
  }

  async function updatePreferences(input: {
    baseCurrency?: string
    onboardingCompleted?: boolean
    timezone?: string
    weeklyReviewHour?: number
    weeklyReviewWeekday?: number
  }) {
    const response = await updateUserPreferences(input)
    profile.value = {
      ...mapApiUser(response.user),
      provider: profile.value.provider,
    }
    return profile.value
  }

  async function completeOnboarding(payload: {
    currency: string
    fullName: string
    timezone: string
  }) {
    const nextProfile = await updatePreferences({
      baseCurrency: payload.currency,
      onboardingCompleted: true,
      timezone: payload.timezone,
    })

    profile.value = {
      ...nextProfile,
      fullName: payload.fullName.trim() || nextProfile.fullName,
    }

    return profile.value
  }

  function reset() {
    profile.value = createEmptyProfile()
  }

  return {
    bootstrap,
    completeOnboarding,
    fetchProfile,
    loading,
    profile,
    reset,
    setProfile,
    updatePreferences,
  }
})
