<script setup lang="ts">
import { ArrowLeftRight, Minus, Plus, X } from 'lucide-vue-next'
import { computed, ref, watch } from 'vue'
import { useRoute } from 'vue-router'

import { useI18n } from '@/shared/i18n'

const route = useRoute()
const { t } = useI18n()
const expanded = ref(false)

const hidden = computed(() =>
  ['transaction-new', 'transaction-edit', 'transaction-transfer', 'login', 'onboarding'].includes(
    String(route.name ?? ''),
  ),
)

watch(
  () => route.fullPath,
  () => {
    expanded.value = false
  },
)

const actions = computed(() => [
  {
    icon: Minus,
    label: t('dashboard.quickExpense'),
    to: '/transactions/new?kind=expense',
    tone: 'expense',
  },
  {
    icon: Plus,
    label: t('dashboard.quickIncome'),
    to: '/transactions/new?kind=income',
    tone: 'income',
  },
  {
    icon: ArrowLeftRight,
    label: t('common.transfer'),
    to: '/transactions/transfer',
    tone: 'neutral',
  },
])
</script>

<template>
  <div v-if="!hidden" class="fab-wrap">
    <template v-if="expanded">
      <RouterLink
        v-for="(action, i) in actions"
        :key="action.to"
        class="fab-action"
        :class="`fab-action--${action.tone}`"
        :to="action.to"
        :style="{ animationDelay: `${i * 50}ms` }"
      >
        <component :is="action.icon" :size="17" :stroke-width="2" />
        <span>{{ action.label }}</span>
      </RouterLink>
    </template>

    <button
      class="fab-trigger"
      :class="{ 'is-open': expanded }"
      type="button"
      @click="expanded = !expanded"
    >
      <X v-if="expanded" :size="20" :stroke-width="2" />
      <Plus v-else :size="22" :stroke-width="2.2" />
    </button>
  </div>
</template>

<style scoped>
.fab-action--expense {
  color: var(--expense);
}

.fab-action--income {
  color: var(--income);
}

.fab-action--neutral {
  color: var(--text-primary);
}
</style>
