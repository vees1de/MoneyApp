<script setup lang="ts">
import { storeToRefs } from 'pinia'

import { useAppUiStore } from '@/app/stores/app-ui'

const appUiStore = useAppUiStore()
const { toasts } = storeToRefs(appUiStore)
</script>

<template>
  <div
    v-if="toasts.length"
    class="toast-stack"
  >
    <div
      v-for="toast in toasts"
      :key="toast.id"
      class="toast"
      :class="`toast--${toast.tone}`"
    >
      <span class="toast__text">{{ toast.text }}</span>
      <button
        v-if="toast.actionLabel"
        class="toast__action"
        type="button"
        @click="appUiStore.runToastAction(toast.id)"
      >
        {{ toast.actionLabel }}
      </button>
    </div>
  </div>
</template>
