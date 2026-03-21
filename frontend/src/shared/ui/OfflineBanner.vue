<script setup lang="ts">
import { storeToRefs } from 'pinia'

import { useAppUiStore } from '@/app/stores/app-ui'
import { useI18n } from '@/shared/i18n'

const appUiStore = useAppUiStore()
const { isOnline } = storeToRefs(appUiStore)
const { t } = useI18n()
</script>

<template>
  <div v-if="!isOnline" class="offline-banner">
    {{ t('offline.banner') }}
  </div>
</template>
