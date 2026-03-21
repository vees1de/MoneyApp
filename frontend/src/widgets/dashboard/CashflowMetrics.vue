<script setup lang="ts">
import StatCard from '@/shared/ui/StatCard.vue'

defineProps<{
  inflow: string
  outflow: string
  safeToSpend: string
}>()
</script>

<template>
  <div class="metrics-grid">
    <StatCard label="Income" :value="inflow" hint="Captured for the active period" />
    <StatCard label="Expenses" :value="outflow" hint="Recorded outflow across all accounts" />
    <StatCard label="Safe to spend" :value="safeToSpend" hint="Budget left after reserve buffer" />
  </div>
</template>
