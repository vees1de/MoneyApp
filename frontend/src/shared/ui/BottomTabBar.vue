<script setup lang="ts">
import { computed } from 'vue'

import { useI18n } from '@/shared/i18n'

const { t } = useI18n()

const items = computed(() => [
  {
    label: t('nav.dashboard'),
    to: '/dashboard',
    icon: `<svg viewBox="0 0 24 24" fill="currentColor" width="24" height="24"><path d="M10.55 2.533a2.25 2.25 0 0 1 2.9 0l6.75 5.695c.508.427.8 1.056.8 1.72V19.75A2.25 2.25 0 0 1 18.75 22H15.5a2.25 2.25 0 0 1-2.25-2.25v-3.5a.75.75 0 0 0-.75-.75h-1a.75.75 0 0 0-.75.75v3.5A2.25 2.25 0 0 1 8.5 22H5.25A2.25 2.25 0 0 1 3 19.75V9.948c0-.664.292-1.293.8-1.72l6.75-5.695Z"/></svg>`,
  },
  {
    label: t('nav.accounts'),
    to: '/accounts',
    icon: `<svg viewBox="0 0 24 24" fill="currentColor" width="24" height="24"><path d="M4.5 3.75a3 3 0 0 0-3 3v.75h21v-.75a3 3 0 0 0-3-3h-15ZM21.75 11.25H2.25v7.5a3 3 0 0 0 3 3h13.5a3 3 0 0 0 3-3v-7.5ZM7.5 16.5h-1.5a.75.75 0 0 1 0-1.5h1.5a.75.75 0 0 1 0 1.5Zm4.5 0H10.5a.75.75 0 0 1 0-1.5H12a.75.75 0 0 1 0 1.5Z"/></svg>`,
  },
  {
    label: t('nav.transactions'),
    to: '/transactions',
    icon: `<svg viewBox="0 0 24 24" fill="currentColor" width="24" height="24"><path fill-rule="evenodd" d="M15.97 2.47a.75.75 0 0 1 1.06 0l4.5 4.5a.75.75 0 0 1 0 1.06l-4.5 4.5a.75.75 0 1 1-1.06-1.06l3.22-3.22H7.5a.75.75 0 0 1 0-1.5h11.69l-3.22-3.22a.75.75 0 0 1 0-1.06Zm-7.94 9a.75.75 0 0 1 0 1.06l-3.22 3.22H16.5a.75.75 0 0 1 0 1.5H4.81l3.22 3.22a.75.75 0 1 1-1.06 1.06l-4.5-4.5a.75.75 0 0 1 0-1.06l4.5-4.5a.75.75 0 0 1 1.06 0Z" clip-rule="evenodd"/></svg>`,
  },
  {
    label: t('nav.savings'),
    to: '/savings',
    icon: `<svg viewBox="0 0 24 24" fill="currentColor" width="24" height="24"><path fill-rule="evenodd" d="M12 2.25c-5.385 0-9.75 4.365-9.75 9.75s4.365 9.75 9.75 9.75 9.75-4.365 9.75-9.75S17.385 2.25 12 2.25Zm0 8.625a1.125 1.125 0 1 0 0 2.25 1.125 1.125 0 0 0 0-2.25ZM15.375 12a3.375 3.375 0 1 1-6.75 0 3.375 3.375 0 0 1 6.75 0ZM12 5.25a.75.75 0 0 1 .75.75v1.013a6.758 6.758 0 0 1 5.237 5.237H19a.75.75 0 0 1 0 1.5h-1.013A6.758 6.758 0 0 1 12.75 19v1a.75.75 0 0 1-1.5 0v-1.013A6.758 6.758 0 0 1 5.013 12.75H4a.75.75 0 0 1 0-1.5h1.013A6.758 6.758 0 0 1 11.25 6.013V6a.75.75 0 0 1 .75-.75Z" clip-rule="evenodd"/></svg>`,
  },
  {
    label: t('nav.review'),
    to: '/review',
    icon: `<svg viewBox="0 0 24 24" fill="currentColor" width="24" height="24"><path fill-rule="evenodd" d="M2.25 12c0-5.385 4.365-9.75 9.75-9.75s9.75 4.365 9.75 9.75-4.365 9.75-9.75 9.75S2.25 17.385 2.25 12Zm13.36-1.814a.75.75 0 1 0-1.22-.872l-3.236 4.53L9.53 12.22a.75.75 0 0 0-1.06 1.06l2.25 2.25a.75.75 0 0 0 1.14-.094l3.75-5.25Z" clip-rule="evenodd"/></svg>`,
  },
])
</script>

<template>
  <nav :aria-label="t('nav.dashboard')" class="bottom-nav">
    <RouterLink
      v-for="item in items"
      :key="item.to"
      :to="item.to"
      class="bottom-nav__link"
    >
      <!-- eslint-disable-next-line vue/no-v-html -->
      <span style="display:flex" v-html="item.icon" />
      <span class="bottom-nav__label">{{ item.label }}</span>
    </RouterLink>
  </nav>
</template>
