import { computed, ref, watch } from 'vue'
import { defineStore } from 'pinia'

import type { AuthProvider } from '@/entities/user/model/types'
import { removeStorage, readStorage, writeStorage } from '@/shared/lib/storage'
import { useUserStore } from './user'

type AuthStatus = 'unknown' | 'guest' | 'authenticated'

interface StoredSession {
  status: AuthStatus
  provider: AuthProvider | null
  accessToken: string | null
  expiresAt: string | null
}

const AUTH_STORAGE_KEY = 'plos-auth-session'

const defaultSession: StoredSession = {
  status: 'guest',
  provider: null,
  accessToken: null,
  expiresAt: null,
}

export const useAuthStore = defineStore('auth', () => {
  const status = ref<AuthStatus>('unknown')
  const provider = ref<AuthProvider | null>(null)
  const accessToken = ref<string | null>(null)
  const expiresAt = ref<string | null>(null)
  const bootstrapped = ref(false)

  function bootstrap() {
    const saved = readStorage<StoredSession>(AUTH_STORAGE_KEY, defaultSession)

    status.value = saved.status
    provider.value = saved.provider
    accessToken.value = saved.accessToken
    expiresAt.value = saved.expiresAt

    if (saved.status === 'authenticated' && saved.provider) {
      useUserStore().setAuthProvider(saved.provider)
    }

    bootstrapped.value = true
  }

  function login(nextProvider: AuthProvider) {
    const userStore = useUserStore()

    status.value = 'authenticated'
    provider.value = nextProvider
    accessToken.value = `demo-${nextProvider}-token`
    expiresAt.value = new Date(Date.now() + 1000 * 60 * 45).toISOString()

    userStore.setAuthProvider(nextProvider)
  }

  function logout() {
    status.value = 'guest'
    provider.value = null
    accessToken.value = null
    expiresAt.value = null

    removeStorage(AUTH_STORAGE_KEY)
  }

  const isAuthenticated = computed(() => status.value === 'authenticated')

  watch(
    [status, provider, accessToken, expiresAt],
    () => {
      if (status.value === 'guest') {
        removeStorage(AUTH_STORAGE_KEY)
        return
      }

      if (status.value === 'unknown') {
        return
      }

      writeStorage(AUTH_STORAGE_KEY, {
        status: status.value,
        provider: provider.value,
        accessToken: accessToken.value,
        expiresAt: expiresAt.value,
      } satisfies StoredSession)
    },
  )

  return {
    accessToken,
    bootstrapped,
    expiresAt,
    isAuthenticated,
    login,
    logout,
    provider,
    status,
    bootstrap,
  }
})
