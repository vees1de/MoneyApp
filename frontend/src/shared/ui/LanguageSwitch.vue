<script setup lang="ts">
import { useI18n } from '@/shared/i18n'

const { locale, setLocale, t } = useI18n()
</script>

<template>
  <div class="language-switch" role="group" :aria-label="t('settings.languageTitle')">
    <button
      class="language-switch__btn"
      :class="{ 'language-switch__btn--active': locale === 'ru' }"
      type="button"
      @click="setLocale('ru')"
    >
      {{ t('locale.ru') }}
    </button>
    <button
      class="language-switch__btn"
      :class="{ 'language-switch__btn--active': locale === 'en' }"
      type="button"
      @click="setLocale('en')"
    >
      {{ t('locale.en') }}
    </button>
  </div>
</template>

<style scoped>
.language-switch {
  display: inline-flex;
  padding: 4px;
  border-radius: var(--radius-pill);
  background: var(--surface-secondary);
  border: 1px solid var(--border);
}

.language-switch__btn {
  min-width: 92px;
  height: 36px;
  padding: 0 14px;
  border: 0;
  border-radius: var(--radius-pill);
  background: transparent;
  color: var(--text-secondary);
  font-size: 0.875rem;
  font-weight: 600;
  transition: background var(--duration-fast) ease, color var(--duration-fast) ease;
}

.language-switch__btn--active {
  background: var(--surface);
  color: var(--text-primary);
  box-shadow: var(--shadow-xs);
}
</style>
