<script setup lang="ts">
import { computed, onMounted, reactive } from 'vue'

import { useAppUiStore } from '@/app/stores/app-ui'
import { useSavingsStore } from '@/app/stores/savings'
import { useI18n } from '@/shared/i18n'
import { formatMoney, formatPercent, parseAmountToMinor } from '@/shared/lib/money'
import EmptyState from '@/shared/ui/EmptyState.vue'
import PageContainer from '@/shared/ui/PageContainer.vue'

const savingsStore = useSavingsStore()
const appUiStore = useAppUiStore()
const { t } = useI18n()

const form = reactive({
  name: '',
  target: '',
  targetDate: '',
})

const overallPercent = computed(() => formatPercent(savingsStore.progressRatio))

function goalRatio(goal: { savedMinor: number; targetMinor: number }) {
  if (!goal.targetMinor) return 0
  return Math.min(goal.savedMinor / goal.targetMinor, 1)
}

function submit() {
  if (!form.name.trim()) {
    appUiStore.pushToast(t('savings.goalNameRequired'), 'warning')
    return
  }

  const targetMinor = parseAmountToMinor(form.target)
  if (targetMinor <= 0) {
    appUiStore.pushToast(t('savings.targetRequired'), 'warning')
    return
  }

  void (async () => {
    try {
      await savingsStore.addGoal({
        name: form.name,
        targetMinor,
        targetDate: form.targetDate || null,
      })
      appUiStore.pushToast(t('savings.created'), 'success')
      form.name = ''
      form.target = ''
      form.targetDate = ''
    } catch (error) {
      const message = error instanceof Error ? error.message : t('savings.createFailed')
      appUiStore.pushToast(message, 'warning')
    }
  })()
}

onMounted(async () => {
  await savingsStore.hydrateSavings()
})
</script>

<template>
  <PageContainer>
    <div class="page-header">
      <h1>{{ t('savings.title') }}</h1>
      <p class="muted">{{ t('savings.subtitle') }}</p>
    </div>

    <div v-if="savingsStore.visibleGoals.length" class="savings-summary">
      <div class="savings-summary__label">{{ t('savings.overallFunded') }}</div>
      <div class="savings-summary__pct">{{ overallPercent }}</div>
      <div class="savings-summary__track">
        <div
          class="savings-summary__fill"
          :style="{ width: `${savingsStore.progressRatio * 100}%` }"
        />
      </div>
      <button
        class="button button--secondary button--sm"
        type="button"
        @click="savingsStore.toggleShowCompleted"
      >
        {{ savingsStore.showCompleted ? t('savings.hideCompleted') : t('savings.showCompleted') }}
      </button>
    </div>

    <section v-if="savingsStore.visibleGoals.length" class="section-card">
      <div class="goals-list">
        <div v-for="goal in savingsStore.visibleGoals" :key="goal.id" class="goal-row">
          <div class="goal-row__top">
            <span class="goal-name">{{ goal.name }}</span>
            <span class="goal-pct-badge" :class="{ 'goal-pct-badge--complete': goalRatio(goal) >= 1 }">
              {{ formatPercent(goalRatio(goal)) }}
            </span>
          </div>

          <div class="goal-track">
            <div
              class="goal-fill"
              :class="{ 'goal-fill--complete': goalRatio(goal) >= 1 }"
              :style="{ width: `${goalRatio(goal) * 100}%` }"
            />
          </div>

          <div class="goal-row__bottom">
            <span class="goal-saved">{{ formatMoney(goal.savedMinor, goal.currency) }} {{ t('common.saved') }}</span>
            <span class="goal-target">{{ t('common.of') }} {{ formatMoney(goal.targetMinor, goal.currency) }}</span>
          </div>
        </div>
      </div>
    </section>

    <EmptyState
      v-else
      icon="🎯"
      :title="t('savings.emptyTitle')"
      :description="t('savings.emptyDescription')"
    />

    <section class="section-card">
      <h2 class="section-title">{{ t('savings.newGoal') }}</h2>
      <form class="stack" @submit.prevent="submit">
        <div class="field">
          <label for="goalName">{{ t('savings.goalName') }}</label>
          <input id="goalName" v-model="form.name" :placeholder="t('savings.goalPlaceholder')" />
        </div>

        <div class="grid grid--two">
          <div class="field">
            <label for="target">{{ t('common.targetAmount') }}</label>
            <input id="target" v-model="form.target" inputmode="decimal" :placeholder="t('transactionForm.placeholderAmount')" />
          </div>

          <div class="field">
            <label for="targetDate">{{ t('common.targetDate') }}</label>
            <input id="targetDate" v-model="form.targetDate" type="date" />
          </div>
        </div>

        <button class="button button--primary button--block" type="submit">{{ t('savings.create') }}</button>
      </form>
    </section>
  </PageContainer>
</template>

<style scoped>
.section-title {
  margin: 0 0 18px;
  font-size: 1.0625rem;
  font-weight: 600;
  letter-spacing: -0.015em;
}

.savings-summary {
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: var(--radius-xl);
  box-shadow: var(--shadow-sm);
  padding: 22px;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.savings-summary__label {
  font-size: 0.6875rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.06em;
  color: var(--text-muted);
}

.savings-summary__pct {
  font-size: 2.25rem;
  font-weight: 700;
  letter-spacing: -0.04em;
  color: var(--brand);
  line-height: 1;
  font-variant-numeric: tabular-nums;
}

.savings-summary__track {
  height: 8px;
  background: var(--surface-fill);
  border-radius: var(--radius-pill);
  overflow: hidden;
}

.savings-summary__fill {
  height: 100%;
  background: var(--brand);
  border-radius: var(--radius-pill);
  transition: width var(--duration-gentle) var(--ease-out);
}

.goals-list {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.goal-row {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.goal-row__top {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.goal-name {
  font-size: 0.9375rem;
  font-weight: 600;
  color: var(--text-primary);
  letter-spacing: -0.01em;
}

.goal-pct-badge {
  font-size: 0.8125rem;
  font-weight: 700;
  padding: 3px 10px;
  border-radius: var(--radius-pill);
  background: var(--brand-soft);
  color: var(--brand);
  font-variant-numeric: tabular-nums;
}

.goal-pct-badge--complete {
  background: var(--income-soft);
  color: var(--income);
}

.goal-track {
  height: 6px;
  background: var(--surface-fill);
  border-radius: var(--radius-pill);
  overflow: hidden;
}

.goal-fill {
  height: 100%;
  border-radius: var(--radius-pill);
  background: var(--brand);
  transition: width var(--duration-gentle) var(--ease-out);
}

.goal-fill--complete {
  background: var(--income);
}

.goal-row__bottom {
  display: flex;
  justify-content: space-between;
}

.goal-saved {
  font-size: 0.8125rem;
  font-weight: 500;
  color: var(--text-secondary);
}

.goal-target {
  font-size: 0.8125rem;
  color: var(--text-muted);
}
</style>
