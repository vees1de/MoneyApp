<script setup lang="ts">
import { storeToRefs } from 'pinia'

import { useAppUiStore } from '@/app/stores/app-ui'

const appUiStore = useAppUiStore()
const { toasts } = storeToRefs(appUiStore)
</script>

<template>
  <div
    v-if="toasts.length"
    style="position: fixed; right: 20px; bottom: 110px; display: grid; gap: 10px; z-index: 20;"
  >
    <div
      v-for="toast in toasts"
      :key="toast.id"
      class="section-card"
      style="min-width: 240px; padding: 14px 16px;"
    >
      <strong>{{ toast.text }}</strong>
    </div>
  </div>
</template>
