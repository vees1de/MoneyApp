<script setup lang="ts">
import type { SavingsGoal } from '@/entities/savings-goal/model/types'
import { useI18n } from '@/shared/i18n'
import { formatMoney, formatPercent } from '@/shared/lib/money'

defineProps<{
  goals: SavingsGoal[]
  progressRatio: number
}>()

const { t } = useI18n()

function goalRatio(goal: SavingsGoal): number {
  if (!goal.targetMinor) return 0
  return Math.min(goal.savedMinor / goal.targetMinor, 1)
}
</script>

<template>
  <section class="section-card">
    <div class="sg-header">
      <h3>{{ t('dashboard.savingsTitle') }}</h3>
      <span class="sg-badge">{{ formatPercent(progressRatio) }}</span>
    </div>

    <div v-if="goals.length" class="sg-list">
      <div v-for="goal in goals" :key="goal.id" class="sg-row">
        <div class="sg-row-top">
          <span class="sg-name">{{ goal.name }}</span>
          <span class="sg-amount">{{ formatMoney(goal.savedMinor) }}</span>
        </div>
        <div class="sg-track">
          <div class="sg-fill" :style="{ width: `${goalRatio(goal) * 100}%` }" />
        </div>
        <div class="sg-row-sub">
          <span class="sg-pct">{{ t('dashboard.funded', { value: formatPercent(goalRatio(goal)) }) }}</span>
          <span class="sg-target">{{ t('common.of') }} {{ formatMoney(goal.targetMinor) }}</span>
        </div>
      </div>
    </div>
    <div v-else class="sg-empty">
      <RouterLink class="button button--secondary button--sm" to="/savings">{{ t('dashboard.createFirstGoal') }}</RouterLink>
    </div>
  </section>
</template>

<style scoped>
.sg-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 14px;
}

.sg-header h3 {
  margin: 0;
  font-size: 1.0625rem;
  font-weight: 600;
  letter-spacing: -0.01em;
}

.sg-badge {
  font-size: 0.8125rem;
  font-weight: 600;
  padding: 4px 10px;
  border-radius: var(--radius-pill);
  background: var(--brand-soft);
  color: var(--brand);
}

.sg-list {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.sg-row {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.sg-row-top {
  display: flex;
  justify-content: space-between;
  align-items: baseline;
}

.sg-name {
  font-size: 0.9375rem;
  font-weight: 600;
  color: var(--text-primary);
}

.sg-amount {
  font-size: 0.9375rem;
  font-weight: 700;
  letter-spacing: -0.02em;
  font-variant-numeric: tabular-nums;
  color: var(--brand);
}

.sg-track {
  height: 6px;
  background: var(--surface-secondary);
  border-radius: var(--radius-pill);
  overflow: hidden;
}

.sg-fill {
  height: 100%;
  border-radius: var(--radius-pill);
  background: var(--brand);
  transition: width var(--duration-gentle, 500ms) var(--ease-out);
}

.sg-row-sub {
  display: flex;
  justify-content: space-between;
}

.sg-pct {
  font-size: 0.75rem;
  font-weight: 500;
  font-variant-numeric: tabular-nums;
  color: var(--brand);
}

.sg-target {
  font-size: 0.75rem;
  font-variant-numeric: tabular-nums;
  color: var(--text-muted);
}

.sg-empty {
  padding: 8px 0;
}
</style>
