<script setup lang="ts">
import { ArrowLeftRight, Minus, Plus, WandSparkles } from 'lucide-vue-next'
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
  <div v-if="!hidden" class="quick-fab">
    <div v-if="expanded" class="quick-fab__menu">
      <RouterLink
        v-for="action in actions"
        :key="action.to"
        :class="['quick-fab__item', `quick-fab__item--${action.tone}`]"
        :to="action.to"
      >
        <component :is="action.icon" :size="18" :stroke-width="2.2" />
        <span>{{ action.label }}</span>
      </RouterLink>
    </div>

    <button class="quick-fab__button" type="button" @click="expanded = !expanded">
      <WandSparkles v-if="expanded" :size="20" :stroke-width="2.2" />
      <Plus v-else :size="20" :stroke-width="2.2" />
    </button>
  </div>
</template>

<style scoped>
.quick-fab {
  position: fixed;
  right: max(16px, calc((100vw - var(--content-width)) / 2 + 16px));
  bottom: 92px;
  display: grid;
  gap: 10px;
  z-index: 18;
}

.quick-fab__menu {
  display: grid;
  gap: 10px;
}

.quick-fab__item {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  justify-content: flex-start;
  min-height: 44px;
  padding: 0 16px;
  border-radius: var(--radius-pill);
  box-shadow: var(--shadow);
  background: var(--surface);
  color: var(--text-primary);
  font-size: 0.875rem;
  font-weight: 700;
}

.quick-fab__item--expense {
  color: var(--expense);
}

.quick-fab__item--income {
  color: var(--income);
}

.quick-fab__button {
  justify-self: end;
  width: 56px;
  height: 56px;
  border: none;
  border-radius: 50%;
  background: linear-gradient(135deg, var(--brand), #36a4ff);
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: var(--shadow-lg);
}

@media (min-width: 960px) {
  .quick-fab {
    bottom: 32px;
  }
}
</style>
