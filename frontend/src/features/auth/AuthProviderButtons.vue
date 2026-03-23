<script setup lang="ts">
import { Eye, EyeOff, Mail } from 'lucide-vue-next'
import { ref } from 'vue'

import type { AuthProvider } from '@/entities/user/model/types'
import { useI18n } from '@/shared/i18n'

defineProps<{
  loading?: boolean
}>()

const emit = defineEmits<{
  select: [provider: Exclude<AuthProvider, 'email'>]
  loginEmail: [email: string, password: string]
}>()

const showEmailForm = ref(false)
const email = ref('')
const password = ref('')
const showPassword = ref(false)
const { t } = useI18n()

function submitEmail() {
  if (!email.value || !password.value) return
  emit('loginEmail', email.value, password.value)
}
</script>

<template>
  <div class="auth-providers">
    <!-- OAuth buttons -->
    <button
      class="auth-btn auth-btn--telegram"
      :disabled="loading"
      @click="emit('select', 'telegram')"
    >
      <!-- Telegram brand icon -->
      <svg viewBox="0 0 24 24" width="20" height="20" fill="none" xmlns="http://www.w3.org/2000/svg">
        <circle cx="12" cy="12" r="12" fill="white" fill-opacity="0.25"/>
        <path d="M5.02 11.64l12.3-4.74c.57-.21 1.07.14.89.99l-2.09 9.84c-.15.71-.57.88-1.15.55l-3.2-2.36-1.54 1.49c-.17.17-.32.31-.65.31l.23-3.27 5.93-5.36c.26-.23-.06-.36-.4-.13L7.08 13.52 3.92 12.54c-.69-.22-.7-.69.13-.98z" fill="white"/>
      </svg>
      {{ t('auth.continueTelegram') }}
    </button>

    <button
      class="auth-btn auth-btn--yandex"
      :disabled="loading"
      @click="emit('select', 'yandex')"
    >
      <!-- Yandex brand icon — explicit fill, no currentColor dependency -->
      <svg viewBox="0 0 24 24" width="20" height="20" xmlns="http://www.w3.org/2000/svg">
        <circle cx="12" cy="12" r="10" fill="white" fill-opacity="0.25"/>
        <path
          d="M13.32 7.666h-.924c-1.694 0-2.585.858-2.585 2.123 0 1.43.616 2.1 1.881 2.959l1.045.704-3.003 4.548H7.919l2.739-4.108c-1.55-1.111-2.43-2.18-2.43-4.02 0-2.289 1.595-3.872 4.092-3.872h2.376v12h-1.376V7.666z"
          fill="white"
        />
      </svg>
      {{ t('auth.continueYandex') }}
    </button>

    <div class="auth-divider">
      <span>{{ t('auth.or') }}</span>
    </div>

    <!-- Email form -->
    <template v-if="showEmailForm">
      <form class="email-form" @submit.prevent="submitEmail">
        <div class="email-form__field">
          <Mail :size="16" class="email-form__icon" />
          <input
            v-model="email"
            type="email"
            class="email-form__input"
            :placeholder="t('auth.emailPlaceholder')"
            autocomplete="email"
            required
          />
        </div>
        <div class="email-form__field">
          <component :is="showPassword ? EyeOff : Eye" :size="16" class="email-form__icon email-form__icon--password" @click="showPassword = !showPassword" />
          <input
            v-model="password"
            :type="showPassword ? 'text' : 'password'"
            class="email-form__input"
            :placeholder="t('auth.passwordPlaceholder')"
            autocomplete="current-password"
            required
          />
        </div>
        <button
          type="submit"
          class="auth-btn auth-btn--email"
          :disabled="loading || !email || !password"
        >
          {{ t('auth.signInEmail') }}
        </button>
      </form>
    </template>

    <button
      v-else
      class="auth-btn auth-btn--ghost"
      :disabled="loading"
      @click="showEmailForm = true"
    >
      <Mail :size="18" />
      {{ t('auth.continueEmail') }}
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
  cursor: pointer;
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

.auth-btn--email {
  background: var(--brand);
  color: #fff;
}

.auth-btn--ghost {
  background: var(--surface-secondary);
  color: var(--text-primary);
  border: 1px solid var(--separator);
}

.auth-divider {
  display: flex;
  align-items: center;
  gap: 12px;
  color: var(--text-muted);
  font-size: 0.8125rem;
  font-weight: 500;
}

.auth-divider::before,
.auth-divider::after {
  content: '';
  flex: 1;
  height: 1px;
  background: var(--separator);
}

.email-form {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.email-form__field {
  position: relative;
  display: flex;
  align-items: center;
}

.email-form__icon {
  position: absolute;
  left: 14px;
  color: var(--text-muted);
  pointer-events: none;
  flex-shrink: 0;
}

.email-form__icon--password {
  pointer-events: auto;
  cursor: pointer;
  left: auto;
  right: 14px;
}

.email-form__input {
  width: 100%;
  height: 48px;
  padding: 0 44px 0 40px;
  border: 1px solid var(--separator);
  border-radius: var(--radius-md);
  background: var(--surface-secondary);
  font-size: 0.9375rem;
  color: var(--text-primary);
  outline: none;
  transition: border-color var(--duration-fast) ease;
}

.email-form__input:focus {
  border-color: var(--brand);
  background: var(--surface);
}

.email-form__input::placeholder {
  color: var(--text-muted);
}
</style>
