<script setup lang="ts">
import type { AuthProvider } from '@/entities/user/model/types'

defineProps<{
  loading?: boolean
}>()

const emit = defineEmits<{
  select: [provider: AuthProvider]
}>()
</script>

<template>
  <div class="stack">
    <button class="button button--primary button--block" :disabled="loading" @click="emit('select', 'telegram')">
      Continue with Telegram
    </button>
    <button
      class="button button--secondary button--block"
      :disabled="loading"
      @click="emit('select', 'yandex')"
    >
      Continue with Yandex
    </button>
  </div>
</template>
