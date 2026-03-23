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
</script>

<template>
  <aside class="sidebar">
    <!-- Logo -->
    <div class="sidebar__brand">
      <div class="sidebar__logo">
        <svg viewBox="0 0 28 28" fill="none" width="22" height="22">
          <rect width="28" height="28" rx="8" fill="var(--brand)"/>
          <path d="M8 20V10.5L14 7l6 3.5V20" stroke="#fff" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"/>
          <path d="M11 20v-5h6v5" stroke="#fff" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"/>
        </svg>
      </div>
      <div class="sidebar__brand-text">
        <span class="sidebar__app-name">{{ env.appName }}</span>
        <span class="sidebar__product">{{ env.productName }}</span>
      </div>
    </div>

    <!-- User greeting -->
    <div v-if="firstName" class="sidebar__user">
      <div class="sidebar__avatar">{{ firstName.charAt(0).toUpperCase() }}</div>
      <div class="sidebar__user-info">
        <span class="sidebar__user-name">{{ firstName }}</span>
        <span class="sidebar__user-sub">{{ t('header.greeting', { name: firstName }) }}</span>
      </div>
    </div>

    <!-- Nav -->
    <nav class="sidebar__nav">
      <RouterLink
        v-for="item in navItems"
        :key="item.to"
        :to="item.to"
        class="sidebar__link"
        :class="{ 'sidebar__link--active': route.path === item.to || route.path.startsWith(item.to + '/') }"
      >
        <component :is="item.icon" :size="18" :stroke-width="route.path === item.to ? 2.25 : 1.75" />
        <span>{{ item.label() }}</span>
      </RouterLink>
    </nav>

    <!-- Settings at bottom -->
    <div class="sidebar__footer">
      <RouterLink to="/settings" class="sidebar__link" :class="{ 'sidebar__link--active': route.path === '/settings' }">
        <Settings :size="18" :stroke-width="route.path === '/settings' ? 2.25 : 1.75" />
        <span>{{ t('route.settings') }}</span>
      </RouterLink>
    </div>
  </aside>
</template>

<style scoped>
.sidebar {
  width: 220px;
  flex-shrink: 0;
  height: 100vh;
  position: sticky;
  top: 0;
  display: flex;
  flex-direction: column;
  background: var(--surface);
  border-right: 1px solid var(--separator);
  padding: 20px 12px 24px;
  overflow-y: auto;
}

.sidebar__brand {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 4px 8px 20px;
}

.sidebar__logo {
  flex-shrink: 0;
  border-radius: 8px;
  overflow: hidden;
  display: flex;
}

.sidebar__brand-text {
  display: flex;
  flex-direction: column;
  gap: 1px;
  min-width: 0;
}

.sidebar__app-name {
  font-size: 0.9375rem;
  font-weight: 700;
  color: var(--text-primary);
  letter-spacing: -0.02em;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.sidebar__product {
  font-size: 0.75rem;
  color: var(--text-muted);
}

.sidebar__user {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 8px 16px;
  border-bottom: 1px solid var(--separator);
  margin-bottom: 8px;
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

.sidebar__user-info {
  display: flex;
  flex-direction: column;
  gap: 1px;
  min-width: 0;
}

.sidebar__user-name {
  font-size: 0.875rem;
  font-weight: 600;
  color: var(--text-primary);
}

.sidebar__user-sub {
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
  padding: 9px 10px;
  border-radius: var(--radius-md);
  font-size: 0.9375rem;
  font-weight: 500;
  color: var(--text-secondary);
  transition: background var(--duration-fast) ease, color var(--duration-fast) ease;
  text-decoration: none;
}

.sidebar__link:hover {
  background: var(--surface-secondary);
  color: var(--text-primary);
}

.sidebar__link--active {
  background: var(--brand-soft);
  color: var(--brand);
  font-weight: 600;
}

.sidebar__footer {
  padding-top: 12px;
  border-top: 1px solid var(--separator);
}
</style>
