<script setup lang="ts">
import { computed } from 'vue'

import { useI18n } from '@/shared/i18n'
import { formatMoney } from '@/shared/lib/money'

const props = defineProps<{
  actualBalanceMinor: number | null
  deltaHint: string
  deltaMinor: number | null
  expectedBalanceMinor: number
}>()

const deltaColor = computed(() => {
  if (props.deltaMinor === null) return 'var(--text-muted)'
  if (props.deltaMinor === 0) return 'var(--income)'
  return 'var(--expense)'
})

const deltaBg = computed(() => {
  if (props.deltaMinor === null) return 'var(--surface-secondary)'
  if (props.deltaMinor === 0) return 'var(--income-soft)'
  return 'var(--danger-soft)'
})

const { t } = useI18n()
</script>

<template>
  <section class="surface-card delta-card">
    <div class="delta-header">
      <div>
        <h3 style="margin:0;font-size:1.0625rem;font-weight:600;letter-spacing:-0.01em">{{ t('review.balanceCheck') }}</h3>
        <p class="muted" style="margin:4px 0 0;font-size:0.875rem">{{ t('review.balanceCheckBody') }}</p>
      </div>
      <div
        v-if="deltaMinor !== null"
        class="delta-badge"
        :style="{ background: deltaBg, color: deltaColor }"
      >
        {{ deltaMinor >= 0 ? '+' : '−' }}{{ formatMoney(Math.abs(deltaMinor)) }}
      </div>
    </div>

    <div class="delta-grid">
      <div class="delta-metric">
        <span class="delta-metric__label">{{ t('common.expected') }}</span>
        <strong class="delta-metric__value">{{ formatMoney(expectedBalanceMinor) }}</strong>
      </div>
      <div class="delta-separator">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" width="18" height="18" style="color:var(--text-muted)">
          <path stroke-linecap="round" stroke-linejoin="round" d="M8.25 4.5l7.5 7.5-7.5 7.5"/>
        </svg>
      </div>
      <div class="delta-metric delta-metric--right">
        <span class="delta-metric__label">{{ t('common.actual') }}</span>
        <strong class="delta-metric__value" :style="{ color: actualBalanceMinor !== null ? 'var(--text-primary)' : 'var(--text-muted)' }">
          {{ actualBalanceMinor === null ? t('common.notYet') : formatMoney(actualBalanceMinor) }}
        </strong>
      </div>
    </div>

    <p v-if="deltaHint" class="delta-hint">{{ deltaHint }}</p>
  </section>
</template>

<style scoped>
.delta-card {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.delta-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
}

.delta-badge {
  flex-shrink: 0;
  font-size: 0.875rem;
  font-weight: 700;
  letter-spacing: -0.02em;
  padding: 6px 12px;
  border-radius: var(--radius-pill);
}

.delta-grid {
  display: flex;
  align-items: center;
  gap: 12px;
  background: var(--surface-secondary);
  border-radius: var(--radius-lg);
  padding: 16px;
}

.delta-metric {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.delta-metric--right {
  text-align: right;
  align-items: flex-end;
}

.delta-metric__label {
  font-size: 0.6875rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--text-muted);
}

.delta-metric__value {
  font-size: 1.25rem;
  font-weight: 700;
  letter-spacing: -0.03em;
  color: var(--text-primary);
}

.delta-separator {
  display: flex;
  align-items: center;
  justify-content: center;
}

.delta-hint {
  margin: 0;
  font-size: 0.875rem;
  color: var(--text-secondary);
  line-height: 1.45;
  padding: 12px 14px;
  background: var(--surface-secondary);
  border-radius: var(--radius-md);
}
</style>
