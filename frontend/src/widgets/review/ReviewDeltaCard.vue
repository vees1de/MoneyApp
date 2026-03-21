<script setup lang="ts">
import { computed } from 'vue'

import { formatMoney } from '@/shared/lib/money'

const props = defineProps<{
  actualBalanceMinor: number | null
  deltaHint: string
  deltaMinor: number | null
  expectedBalanceMinor: number
}>()

const deltaTone = computed(() => {
  if (props.deltaMinor === null) {
    return ''
  }

  if (props.deltaMinor === 0) {
    return 'pill--success'
  }

  return 'pill--warning'
})
</script>

<template>
  <section class="surface-card stack">
    <div class="section-card__header">
      <div>
        <h3>Balance check</h3>
        <p class="muted">Expected vs actual balance for the active review window.</p>
      </div>
      <span v-if="deltaMinor !== null" class="pill" :class="deltaTone">
        Delta: {{ formatMoney(deltaMinor) }}
      </span>
    </div>

    <div class="grid grid--two">
      <div class="stat-card">
        <span class="stat-card__label">Expected balance</span>
        <strong class="stat-card__value">{{ formatMoney(expectedBalanceMinor) }}</strong>
      </div>
      <div class="stat-card">
        <span class="stat-card__label">Actual balance</span>
        <strong class="stat-card__value">
          {{ actualBalanceMinor === null ? 'Not submitted' : formatMoney(actualBalanceMinor) }}
        </strong>
      </div>
    </div>

    <p class="muted">{{ deltaHint }}</p>
  </section>
</template>
