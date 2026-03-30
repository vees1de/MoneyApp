<script setup lang="ts">
import { ArrowLeftRight, ClipboardCheck, CreditCard, House, PiggyBank, Settings } from 'lucide-vue-next'
import { computed } from 'vue'
import { useRoute } from 'vue-router'

import { env } from '@/shared/config/env'
import { useUserStore } from '@/app/stores/user'
import { useI18n } from '@/shared/i18n'

const route = useRoute()
const userStore = useUserStore()
const { t } = useI18n()

const firstName = computed(() => userStore.profile.fullName?.split(' ')[0] ?? '')

const navItems = [
  { label: () => t('nav.dashboard'), to: '/dashboard', icon: House },
  { label: () => t('nav.accounts'), to: '/accounts', icon: CreditCard },
  { label: () => t('nav.transactions'), to: '/transactions', icon: ArrowLeftRight },
  { label: () => t('nav.savings'), to: '/savings', icon: PiggyBank },
  { label: () => t('nav.review'), to: '/review', icon: ClipboardCheck },
]

function isActive(to: string): boolean {
  return route.path === to || route.path.startsWith(to + '/')
}
</script>

<template>
  <aside class="sidebar">
    <div class="sidebar__logo">
      <div class="sidebar__logo-icon">
        <svg viewBox="0 0 64 64" width="20" height="20">
          <rect width="64" height="64" rx="14" fill="#333"/>
          <circle cx="32" cy="32" r="22" fill="#F5F0E1"/>
          <path d="M22 11.5 C18 14, 14 19, 12.5 24 L17 24 C18 20, 20 17, 23 14.5Z" fill="#333"/>
          <path d="M36 11 C40 12, 43 14, 45 17 L38 17 C37 15, 36.5 13, 36 11Z" fill="#333"/>
          <path d="M22 14 C20 16, 19.5 18, 19.5 20 L19.5 46 C19.5 48, 21 50, 23 50 L25 49 L25 14 C24 13.5, 23 13.5, 22 14Z" fill="#333"/>
          <path d="M25 30 C25 26, 29 23, 34 23 C39 23, 43 26, 44 30 L44 42 C44 47, 40 51, 34 51 C28 51, 25 47, 25 42Z" fill="#333"/>
          <path d="M25 32 C25 27.5, 29 24.5, 34 24.5 C39 24.5, 42.5 27.5, 42.5 32 L42.5 41 C42.5 45.5, 39 48.5, 34 48.5 C29 48.5, 25 45.5, 25 41Z" fill="#F5F0E1"/>
          <ellipse cx="33.5" cy="37" rx="6" ry="7" fill="#333"/>
        </svg>
      </div>
      <span class="sidebar__logo-text">{{ env.appName }}</span>
    </div>

    <div v-if="firstName" class="sidebar__user">
      <div class="sidebar__avatar">{{ firstName.charAt(0).toUpperCase() }}</div>
      <div class="sidebar__user-meta">
        <span class="sidebar__user-name">{{ firstName }}</span>
        <span class="sidebar__user-handle">{{ t('header.greeting', { name: firstName }) }}</span>
      </div>
    </div>

    <nav class="sidebar__nav">
      <RouterLink
        v-for="item in navItems"
        :key="item.to"
        :to="item.to"
        class="sidebar__link"
        :class="{ 'router-link-active': isActive(item.to) }"
      >
        <component :is="item.icon" :size="18" :stroke-width="isActive(item.to) ? 2.2 : 1.7" />
        <span>{{ item.label() }}</span>
      </RouterLink>
    </nav>

    <div class="sidebar__footer">
      <RouterLink
        to="/settings"
        class="sidebar__link"
        :class="{ 'router-link-active': isActive('/settings') }"
      >
        <Settings :size="18" :stroke-width="isActive('/settings') ? 2.2 : 1.7" />
        <span>{{ t('route.settings') }}</span>
      </RouterLink>
    </div>
  </aside>
</template>

<style scoped>
.sidebar {
  width: var(--sidebar-width);
  flex-shrink: 0;
  height: 100vh;
  height: 100dvh;
  position: sticky;
  top: 0;
  display: flex;
  flex-direction: column;
  background: var(--surface-glass);
  backdrop-filter: blur(40px) saturate(200%);
  -webkit-backdrop-filter: blur(40px) saturate(200%);
  border-right: 0.5px solid var(--border);
  padding: 20px 14px 24px;
  overflow-y: auto;
}

.sidebar__logo {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 10px 20px;
}

.sidebar__logo-icon {
  flex-shrink: 0;
  display: flex;
}

.sidebar__logo-text {
  font-size: 1.0625rem;
  font-weight: 700;
  color: var(--text-primary);
  letter-spacing: -0.025em;
}

.sidebar__user {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px 10px;
  margin-bottom: 12px;
  border-radius: var(--radius-sm);
  background: var(--surface-secondary);
}

.sidebar__avatar {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  background: var(--brand-soft);
  color: var(--brand);
  font-size: 0.875rem;
  font-weight: 700;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.sidebar__user-meta {
  display: flex;
  flex-direction: column;
  gap: 1px;
  min-width: 0;
}

.sidebar__user-name {
  font-size: 0.875rem;
  font-weight: 600;
  color: var(--text-primary);
  letter-spacing: -0.01em;
}

.sidebar__user-handle {
  font-size: 0.75rem;
  color: var(--text-muted);
}

.sidebar__nav {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.sidebar__link {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 12px;
  border-radius: var(--radius-xs);
  font-size: 0.9375rem;
  font-weight: 500;
  color: var(--text-secondary);
  letter-spacing: -0.01em;
  text-decoration: none;
  transition: all var(--duration-base) var(--ease-out);
}

.sidebar__link:hover {
  background: var(--surface-secondary);
  color: var(--text-primary);
}

.sidebar__link.router-link-active {
  background: var(--brand-soft);
  color: var(--brand);
  font-weight: 600;
}

.sidebar__footer {
  padding-top: 14px;
  border-top: 0.5px solid var(--separator);
  margin-top: 8px;
}
</style>
