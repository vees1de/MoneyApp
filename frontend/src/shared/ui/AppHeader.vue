<script setup lang="ts">
import { computed } from 'vue'
import { RouterLink, useRoute } from 'vue-router'

import { env } from '@/shared/config/env'
import { useUserStore } from '@/app/stores/user'
import { useI18n } from '@/shared/i18n'

const route = useRoute()
const userStore = useUserStore()
const { t } = useI18n()

const greeting = computed(() => {
  const name = userStore.profile.fullName?.split(' ')[0]
  return name ? t('header.greeting', { name }) : env.appName
})
</script>

<template>
  <header class="header">
    <div class="header__title-block">
      <span class="header__greeting">{{ greeting }}</span>
    </div>

    <RouterLink
      v-if="route.name !== 'settings'"
      class="header-avatar"
      to="/settings"
      :aria-label="t('header.settingsAria')"
    >
      <svg viewBox="0 0 24 24" fill="none" width="20" height="20" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">
        <circle cx="12" cy="12" r="3" />
        <path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1-2.83 2.83l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-4 0v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83-2.83l.06-.06A1.65 1.65 0 0 0 4.68 15a1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1 0-4h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 2.83-2.83l.06.06A1.65 1.65 0 0 0 9 4.68a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 4 0v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 2.83l-.06.06A1.65 1.65 0 0 0 19.4 9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 0 4h-.09a1.65 1.65 0 0 0-1.51 1z" />
      </svg>
    </RouterLink>
  </header>
</template>

<style scoped>
.header-avatar {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 38px;
  height: 38px;
  border-radius: 50%;
  background: var(--surface-secondary);
  color: var(--text-secondary);
  flex-shrink: 0;
  transition: background-color var(--duration-base) var(--ease-out),
              transform var(--duration-fast) var(--ease-out),
              color var(--duration-base) var(--ease-out);
}

.header-avatar:hover {
  background: var(--surface-fill);
  color: var(--text-primary);
}

.header-avatar:active {
  transform: scale(0.92);
  background: var(--surface-fill);
}
</style>
