<script setup lang="ts">
import { storeToRefs } from 'pinia'

import { useAppUiStore } from '@/app/stores/app-ui'

const appUiStore = useAppUiStore()
const { isOnline } = storeToRefs(appUiStore)
</script>

<template>
  <div v-if="!isOnline" class="offline-banner">
    You are offline. The app shows the latest cached state and blocks network-only actions.
  </div>
</template>
