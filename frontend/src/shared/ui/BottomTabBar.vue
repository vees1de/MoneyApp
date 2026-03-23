<script setup lang="ts">
import { ArrowLeftRight, ClipboardCheck, CreditCard, House, PiggyBank } from 'lucide-vue-next'

import { useI18n } from '@/shared/i18n'

const { t } = useI18n()

const items = [
  { label: () => t('nav.dashboard'), to: '/dashboard', icon: House },
  { label: () => t('nav.accounts'), to: '/accounts', icon: CreditCard },
  { label: () => t('nav.transactions'), to: '/transactions', icon: ArrowLeftRight },
  { label: () => t('nav.savings'), to: '/savings', icon: PiggyBank },
  { label: () => t('nav.review'), to: '/review', icon: ClipboardCheck },
]
</script>

<template>
  <nav :aria-label="t('nav.primary')" class="bottom-nav">
    <RouterLink
      v-for="item in items"
      :key="item.to"
      :to="item.to"
      class="bottom-nav__link"
    >
      <component :is="item.icon" :size="24" :stroke-width="1.75" />
      <span class="bottom-nav__label">{{ item.label() }}</span>
    </RouterLink>
  </nav>
</template>
