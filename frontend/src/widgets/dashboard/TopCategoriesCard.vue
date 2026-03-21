<script setup lang="ts">
import type { TopCategorySummary } from '@/entities/dashboard/model/types'
import { formatMoney } from '@/shared/lib/money'

defineProps<{
  categories: TopCategorySummary[]
}>()
</script>

<template>
  <section class="surface-card stack">
    <div class="section-card__header">
      <div>
        <h3>Top categories</h3>
        <p class="muted">Largest expense buckets for the current period.</p>
      </div>
    </div>

    <div v-if="categories.length" class="list">
      <div v-for="category in categories" :key="category.categoryId" class="list-item">
        <span>{{ category.label }}</span>
        <strong class="amount--expense">{{ formatMoney(category.amountMinor) }}</strong>
      </div>
    </div>
    <p v-else class="tiny">No expense categories yet. Add transactions to unlock spending insights.</p>
  </section>
</template>
