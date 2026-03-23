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
      class="section-card"
      :class="`toast-stack__item toast-stack__item--${toast.tone}`"
    >
      <div class="toast-stack__copy">
        <strong>{{ toast.text }}</strong>
        <button
          v-if="toast.actionLabel"
          class="toast-stack__action"
          type="button"
          @click="appUiStore.runToastAction(toast.id)"
        >
          {{ toast.actionLabel }}
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.toast-stack {
  position: fixed;
  right: 16px;
  left: 16px;
  bottom: 104px;
  display: grid;
  gap: 10px;
  z-index: 30;
}

.toast-stack__item {
  min-width: 0;
  padding: 14px 16px;
}

.toast-stack__item--success {
  border-color: rgba(52, 199, 89, 0.16);
}

.toast-stack__item--warning {
  border-color: rgba(255, 149, 0, 0.16);
}

.toast-stack__copy {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.toast-stack__action {
  border: none;
  background: transparent;
  color: var(--brand);
  font-size: 0.875rem;
  font-weight: 700;
  padding: 0;
}

@media (min-width: 960px) {
  .toast-stack {
    left: auto;
    width: min(360px, calc(100vw - 32px));
  }
}
</style>
