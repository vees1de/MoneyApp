<script setup lang="ts">
import { useI18n } from '@/shared/i18n'

defineProps<{
  inflow: string
  outflow: string
  safeToSpend: string
}>()

const { t } = useI18n()
</script>

<template>
  <div class="cashflow-metrics">
    <div class="cashflow-card cashflow-card--income">
      <span class="cashflow-card__label">{{ t('dashboard.cashflowIncome') }}</span>
      <strong class="cashflow-card__value">{{ inflow }}</strong>
      <span class="cashflow-card__sub">{{ t('common.thisPeriod') }}</span>
    </div>

    <div class="cashflow-card cashflow-card--expense">
      <span class="cashflow-card__label">{{ t('dashboard.cashflowExpenses') }}</span>
      <strong class="cashflow-card__value">{{ outflow }}</strong>
      <span class="cashflow-card__sub">{{ t('common.thisPeriod') }}</span>
    </div>

    <div class="cashflow-card cashflow-card--safe">
      <span class="cashflow-card__label">{{ t('dashboard.cashflowSafe') }}</span>
      <strong class="cashflow-card__value">{{ safeToSpend }}</strong>
      <span class="cashflow-card__sub">{{ t('dashboard.afterReserve') }}</span>
    </div>
  </div>
</template>

<style scoped>
.cashflow-metrics {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 10px;
}

.cashflow-card {
  display: flex;
  flex-direction: column;
  gap: 3px;
  padding: 14px 14px 12px;
  border-radius: var(--radius-lg);
  border: 1px solid var(--border);
  background: var(--surface);
  box-shadow: var(--shadow-xs);
}

.cashflow-card__label {
  font-size: 0.6875rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--text-muted);
}

.cashflow-card__value {
  font-size: 1.0625rem;
  font-weight: 700;
  letter-spacing: -0.025em;
  line-height: 1.1;
  margin-top: 2px;
}

.cashflow-card__sub {
  font-size: 0.6875rem;
  color: var(--text-muted);
  margin-top: 2px;
}

.cashflow-card--income .cashflow-card__value {
  color: var(--income);
}

.cashflow-card--expense .cashflow-card__value {
  color: var(--expense);
}

.cashflow-card--safe .cashflow-card__value {
  color: var(--brand);
}
</style>
