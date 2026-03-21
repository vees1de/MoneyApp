<script setup lang="ts">
import type { AuthProvider } from '@/entities/user/model/types'
import { useI18n } from '@/shared/i18n'

defineProps<{
  loading?: boolean
}>()

const emit = defineEmits<{
  select: [provider: AuthProvider]
}>()

const { t } = useI18n()
</script>

<template>
  <div class="auth-providers">
    <button
      class="auth-btn auth-btn--telegram"
      :disabled="loading"
      @click="emit('select', 'telegram')"
    >
      <svg viewBox="0 0 24 24" fill="currentColor" width="20" height="20">
        <path d="M11.944 0A12 12 0 0 0 0 12a12 12 0 0 0 12 12 12 12 0 0 0 12-12A12 12 0 0 0 12 0a12 12 0 0 0-.056 0zm4.962 7.224c.1-.002.321.023.465.14a.506.506 0 0 1 .171.325c.016.093.036.306.02.472-.18 1.898-.962 6.502-1.36 8.627-.168.9-.499 1.201-.82 1.23-.696.065-1.225-.46-1.9-.902-1.056-.693-1.653-1.124-2.678-1.8-1.185-.78-.417-1.21.258-1.91.177-.184 3.247-2.977 3.307-3.23.007-.032.014-.15-.056-.212s-.174-.041-.249-.024c-.106.024-1.793 1.14-5.061 3.345-.48.33-.913.49-1.302.48-.428-.008-1.252-.241-1.865-.44-.752-.245-1.349-.374-1.297-.789.027-.216.325-.437.893-.663 3.498-1.524 5.83-2.529 6.998-3.014 3.332-1.386 4.025-1.627 4.476-1.635z"/>
      </svg>
      {{ t('auth.continueTelegram') }}
    </button>

    <button
      class="auth-btn auth-btn--yandex"
      :disabled="loading"
      @click="emit('select', 'yandex')"
    >
      <svg viewBox="0 0 24 24" fill="currentColor" width="20" height="20">
        <path d="M2.04 12c0-5.523 4.476-10 10-10 5.523 0 10 4.477 10 10s-4.477 10-10 10c-5.524 0-10-4.477-10-10z"/>
        <path fill="#fff" d="M13.32 7.666h-.924c-1.694 0-2.585.858-2.585 2.123 0 1.43.616 2.1 1.881 2.959l1.045.704-3.003 4.548H7.919l2.739-4.108c-1.55-1.111-2.43-2.18-2.43-4.02 0-2.289 1.595-3.872 4.092-3.872h2.376v12h-1.376V7.666z"/>
      </svg>
      {{ t('auth.continueYandex') }}
    </button>
  </div>
</template>

<style scoped>
.auth-providers {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.auth-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  width: 100%;
  height: 52px;
  border: 0;
  border-radius: var(--radius-pill);
  font-size: 1rem;
  font-weight: 600;
  letter-spacing: -0.01em;
  transition: opacity var(--duration-fast) ease, transform var(--duration-fast) ease;
}

.auth-btn:active {
  transform: scale(0.97);
  opacity: 0.85;
}

.auth-btn:disabled {
  opacity: 0.45;
  cursor: not-allowed;
}

.auth-btn--telegram {
  background: #2AABEE;
  color: #fff;
}

.auth-btn--yandex {
  background: #FC3F1D;
  color: #fff;
}
</style>
