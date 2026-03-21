import { removeStorage, readStorage, writeStorage } from '@/shared/lib/storage'
import type { AuthProvider } from '@/entities/user/model/types'

const SESSION_STORAGE_KEY = 'plos-session'

export interface StoredSession {
  accessToken: string
  expiresAt: string
  provider: AuthProvider | null
  refreshToken: string
}

export function readSession() {
  return readStorage<StoredSession | null>(SESSION_STORAGE_KEY, null)
}

export function writeSession(session: StoredSession) {
  writeStorage(SESSION_STORAGE_KEY, session)
}

export function clearSession() {
  removeStorage(SESSION_STORAGE_KEY)
}
