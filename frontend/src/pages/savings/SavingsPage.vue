<script setup lang="ts">
import { computed, reactive } from 'vue'

import { useAppUiStore } from '@/app/stores/app-ui'
import { useSavingsStore } from '@/app/stores/savings'
import { formatMoney, formatPercent, parseAmountToMinor } from '@/shared/lib/money'
import EmptyState from '@/shared/ui/EmptyState.vue'
import PageContainer from '@/shared/ui/PageContainer.vue'
import SectionCard from '@/shared/ui/SectionCard.vue'

const savingsStore = useSavingsStore()
const appUiStore = useAppUiStore()

const form = reactive({
  name: '',
  target: '',
  targetDate: '',
})

const summaryText = computed(() => formatPercent(savingsStore.progressRatio))

function submit() {
  if (!form.name.trim()) {
    appUiStore.pushToast('Goal name is required.', 'warning')
    return
  }

  const targetMinor = parseAmountToMinor(form.target)

  if (targetMinor <= 0) {
    appUiStore.pushToast('Target amount should be greater than zero.', 'warning')
    return
  }

  savingsStore.addGoal({
    name: form.name,
    targetMinor,
    targetDate: form.targetDate || null,
  })

  appUiStore.pushToast('Savings goal created.', 'success')

  form.name = ''
  form.target = ''
  form.targetDate = ''
}
</script>

<template>
  <PageContainer>
    <div class="page-header">
      <h1>Savings goals</h1>
      <p class="muted">Savings is not a side-note. It should sit next to cash flow and review in the MVP.</p>
    </div>

    <div class="grid grid--two">
      <form class="surface-card stack" @submit.prevent="submit">
        <div class="page-header">
          <h2>Create goal</h2>
          <p class="muted">Minimal goal model now, richer planning later.</p>
        </div>

        <div class="field">
          <label for="goalName">Goal</label>
          <input id="goalName" v-model="form.name" placeholder="Emergency fund" />
        </div>

        <div class="grid grid--two">
          <div class="field">
            <label for="target">Target amount</label>
            <input id="target" v-model="form.target" inputmode="decimal" placeholder="0.00" />
          </div>

          <div class="field">
            <label for="targetDate">Target date</label>
            <input id="targetDate" v-model="form.targetDate" type="date" />
          </div>
        </div>

        <button class="button button--primary button--block" type="submit">Create goal</button>
      </form>

      <SectionCard>
        <div class="page-header">
          <h2>Portfolio summary</h2>
          <p class="muted">Current goals are {{ summaryText }} funded in aggregate.</p>
        </div>

        <div class="cta-list">
          <button class="button button--secondary" type="button" @click="savingsStore.toggleShowCompleted">
            {{ savingsStore.showCompleted ? 'Hide completed' : 'Show completed' }}
          </button>
          <RouterLink class="button button--secondary" to="/dashboard">Back to dashboard</RouterLink>
        </div>
      </SectionCard>
    </div>

    <SectionCard v-if="savingsStore.visibleGoals.length">
      <div class="list">
        <div v-for="goal in savingsStore.visibleGoals" :key="goal.id" class="list-item">
          <div class="stack" style="gap: 4px;">
            <strong>{{ goal.name }}</strong>
          <span class="tiny">
              {{ formatMoney(goal.savedMinor, goal.currency) }} of {{ formatMoney(goal.targetMinor, goal.currency) }}
          </span>
        </div>

          <span class="pill">{{ formatPercent(goal.targetMinor ? goal.savedMinor / goal.targetMinor : 0) }}</span>
        </div>
      </div>
    </SectionCard>

    <EmptyState
      v-else
      title="No visible goals"
      description="Create the first goal to connect finance tracking with actual intent."
    />
  </PageContainer>
</template>
