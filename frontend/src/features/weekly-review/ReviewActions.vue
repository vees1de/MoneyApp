<script setup lang="ts">
import { computed } from 'vue'

import { useI18n } from '@/shared/i18n'

const emit = defineEmits<{
  resolve: []
  skip: []
}>()

const props = defineProps<{
  canResolve: boolean
  deltaMinor: number | null
}>()

const { t } = useI18n()

const expenseReviewRoute = computed(() => `/transactions/new?kind=expense&note=${encodeURIComponent(t('review.adjustmentNote'))}`)
const incomeReviewRoute = computed(() => `/transactions/new?kind=income&note=${encodeURIComponent(t('review.adjustmentNote'))}`)
</script>

<template>
  <div class="surface-card stack">
    <div class="page-header">
      <h3>{{ t('review.actionsTitle') }}</h3>
      <p class="muted">{{ t('review.actionsBody') }}</p>
    </div>

    <div class="cta-list">
      <RouterLink class="button button--secondary" :to="expenseReviewRoute">
        {{ t('review.addExpense') }}
      </RouterLink>
      <RouterLink class="button button--secondary" :to="incomeReviewRoute">
        {{ t('review.addIncome') }}
      </RouterLink>
      <button class="button button--secondary" type="button" @click="emit('skip')">{{ t('review.skipWeek') }}</button>
    </div>

    <div class="divider" />

    <div class="stack">
      <div class="tiny" v-if="props.deltaMinor === null">
        {{ t('review.deltaMissing') }}
      </div>
      <div class="tiny" v-else-if="props.deltaMinor < 0">
        {{ t('review.deltaNegative') }}
      </div>
      <div class="tiny" v-else-if="props.deltaMinor > 0">
        {{ t('review.deltaPositive') }}
      </div>
      <div class="tiny" v-else>
        {{ t('review.deltaZero') }}
      </div>

      <button class="button button--primary button--block" :disabled="!props.canResolve" @click="emit('resolve')">
        {{ t('review.resolve') }}
      </button>
    </div>
  </div>
</template>
