<script setup lang="ts">
import type { SavingsGoal } from '@/entities/savings-goal/model/types'
import { formatMoney, formatPercent } from '@/shared/lib/money'

defineProps<{
  goals: SavingsGoal[]
  progressRatio: number
}>()
</script>

<template>
  <section class="surface-card stack">
    <div class="section-card__header">
      <div>
        <h3>Savings goals</h3>
        <p class="muted">Keep long-term intent visible, not buried under daily transactions.</p>
      </div>
      <span class="pill">{{ formatPercent(progressRatio) }} funded</span>
    </div>

    <div v-if="goals.length" class="list">
      <div v-for="goal in goals" :key="goal.id" class="list-item">
        <div class="stack" style="gap: 4px;">
          <strong>{{ goal.name }}</strong>
          <span class="tiny">{{ formatMoney(goal.savedMinor) }} of {{ formatMoney(goal.targetMinor) }}</span>
        </div>
        <RouterLink class="button button--secondary" to="/savings">Open</RouterLink>
      </div>
    </div>
    <p v-else class="tiny">No active goals yet. Create one to connect spending with intent.</p>
  </section>
</template>
