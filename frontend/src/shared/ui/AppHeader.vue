<script setup lang="ts">
import { computed } from 'vue'
import { RouterLink, useRoute } from 'vue-router'

import { env } from '@/shared/config/env'
import { useUserStore } from '@/app/stores/user'

const route = useRoute()
const userStore = useUserStore()

const title = computed(() => route.meta.title ?? env.productName)
const subtitle = computed(() => {
  if (route.name === 'dashboard') {
    return 'Track cash flow, review deltas, and keep savings moving.'
  }

  return env.productName
})
</script>

<template>
  <header class="header">
    <div class="header__title">
      <span class="tiny">{{ userStore.profile.fullName || env.appName }}</span>
      <strong>{{ title }}</strong>
      <span class="tiny">{{ subtitle }}</span>
    </div>

    <RouterLink v-if="route.name !== 'settings'" class="button button--secondary" to="/settings">
      Settings
    </RouterLink>
  </header>
</template>
