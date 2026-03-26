<script setup lang="ts">
import { computed } from 'vue'

import type { TopCategorySummary } from '@/entities/dashboard/model/types'
import { useI18n } from '@/shared/i18n'
import { formatMoney } from '@/shared/lib/money'

const props = defineProps<{
  categories: TopCategorySummary[]
}>()

const colors = ['#FF3B30', '#FF9500', '#AF52DE', '#5AC8FA']
const { t } = useI18n()

const total = computed(() =>
  props.categories.reduce((sum, c) => sum + c.amountMinor, 0),
)
</script>

<template>
  <section class="section-card">
    <div class="cat-header">
      <h3>{{ t('dashboard.topSpending') }}</h3>
      <RouterLink class="cat-link" to="/categories">{{ t('common.seeAll') }}</RouterLink>
    </div>

    <div v-if="categories.length" class="cat-list">
      <div v-for="(category, i) in categories" :key="category.categoryId" class="cat-row">
        <div class="cat-dot" :style="{ background: colors[i % colors.length] }" />
        <div class="cat-info">
          <div class="cat-name-row">
            <span class="cat-name">{{ category.label }}</span>
            <span class="cat-amount">{{ formatMoney(category.amountMinor) }}</span>
          </div>
          <div class="cat-track">
            <div
              class="cat-fill"
              :style="{
                width: total > 0 ? `${(category.amountMinor / total) * 100}%` : '0%',
                background: colors[i % colors.length],
              }"
            />
          </div>
        </div>
      </div>
    </div>
    <p v-else class="tiny" style="margin: 0">{{ t('dashboard.topSpendingEmpty') }}</p>
  </section>
</template>

<style scoped>
.cat-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
}

.cat-header h3 {
  margin: 0;
  font-size: 1.0625rem;
  font-weight: 600;
  letter-spacing: -0.015em;
}

.cat-link {
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--brand);
  transition: opacity var(--duration-fast) var(--ease-out);
}

.cat-link:hover {
  opacity: 0.7;
}

.cat-list {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.cat-row {
  display: flex;
  align-items: flex-start;
  gap: 10px;
}

.cat-dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  flex-shrink: 0;
  margin-top: 4px;
}

.cat-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.cat-name-row {
  display: flex;
  justify-content: space-between;
  align-items: baseline;
  gap: 8px;
}

.cat-name {
  font-size: 0.9375rem;
  font-weight: 500;
  color: var(--text-primary);
  letter-spacing: -0.01em;
}

.cat-amount {
  font-size: 0.875rem;
  font-weight: 600;
  color: var(--expense);
  white-space: nowrap;
  font-variant-numeric: tabular-nums;
}

.cat-track {
  height: 4px;
  background: var(--surface-fill);
  border-radius: var(--radius-pill);
  overflow: hidden;
}

.cat-fill {
  height: 100%;
  border-radius: var(--radius-pill);
  transition: width var(--duration-gentle) var(--ease-out);
  opacity: 0.70;
}
</style>
