<script setup lang="ts">
defineProps<{
  canResolve: boolean
  deltaMinor: number | null
}>()

const emit = defineEmits<{
  resolve: []
  skip: []
}>()
</script>

<template>
  <div class="surface-card stack">
    <div class="page-header">
      <h3>Resolution actions</h3>
      <p class="muted">Jump straight into the most likely fix instead of hunting around the app.</p>
    </div>

    <div class="cta-list">
      <RouterLink class="button button--secondary" :to="`/transactions/new?kind=expense&note=Review adjustment`">
        Add expense
      </RouterLink>
      <RouterLink class="button button--secondary" :to="`/transactions/new?kind=income&note=Review adjustment`">
        Add income
      </RouterLink>
      <button class="button button--secondary" type="button" @click="emit('skip')">Skip this week</button>
    </div>

    <div class="divider" />

    <div class="stack">
      <div class="tiny" v-if="deltaMinor === null">
        Enter the actual balance first. The app will explain whether the gap points to missing income or missing spending.
      </div>
      <div class="tiny" v-else-if="deltaMinor < 0">
        Negative delta: most likely an expense was not captured.
      </div>
      <div class="tiny" v-else-if="deltaMinor > 0">
        Positive delta: most likely a missed income or a cancelled expense.
      </div>
      <div class="tiny" v-else>
        Delta is zero. You can confirm the week now.
      </div>

      <button class="button button--primary button--block" :disabled="!canResolve" @click="emit('resolve')">
        Mark review complete
      </button>
    </div>
  </div>
</template>
