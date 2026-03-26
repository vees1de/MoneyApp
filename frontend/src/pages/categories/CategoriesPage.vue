<script setup lang="ts">
import { onMounted, reactive } from 'vue'

import { useAppUiStore } from '@/app/stores/app-ui'
import { useFinanceStore } from '@/app/stores/finance'
import { translateCategoryKind, translateCategoryScope, useI18n } from '@/shared/i18n'
import EmptyState from '@/shared/ui/EmptyState.vue'
import PageContainer from '@/shared/ui/PageContainer.vue'

const financeStore = useFinanceStore()
const appUiStore = useAppUiStore()
const { t } = useI18n()

const form = reactive({
  color: '#0071E3',
  kind: 'expense' as 'income' | 'expense',
  name: '',
})

function submit() {
  if (!form.name.trim()) {
    appUiStore.pushToast(t('categories.nameRequired'), 'warning')
    return
  }

  void (async () => {
    try {
      await financeStore.addCategory({ color: form.color, kind: form.kind, name: form.name })
      appUiStore.pushToast(t('categories.created'), 'success')
      form.name = ''
    } catch (error) {
      const message = error instanceof Error ? error.message : t('categories.createFailed')
      appUiStore.pushToast(message, 'warning')
    }
  })()
}

function remove(categoryId: string) {
  void (async () => {
    try {
      await financeStore.removeCategory(categoryId)
      appUiStore.pushToast(t('categories.deleted'), 'success')
    } catch (error) {
      const message = error instanceof Error ? error.message : t('categories.deleteFailed')
      appUiStore.pushToast(message, 'warning')
    }
  })()
}

onMounted(async () => {
  await financeStore.fetchCategories()
})
</script>

<template>
  <PageContainer>
    <div class="page-header">
      <h1>{{ t('categories.title') }}</h1>
      <p class="muted">{{ t('categories.subtitle') }}</p>
    </div>

    <section v-if="financeStore.categories.length" class="section-card">
      <div class="cat-list">
        <div v-for="category in financeStore.categories" :key="category.id" class="cat-row">
          <div class="cat-swatch" :style="{ background: `${category.color}10`, borderColor: `${category.color}20` }">
            <div class="cat-dot" :style="{ background: category.color }" />
          </div>
          <div class="cat-info">
            <span class="cat-name">{{ category.name }}</span>
            <span class="cat-meta">{{ translateCategoryKind(category.kind) }} · {{ translateCategoryScope(category.scope) }}</span>
          </div>
          <button
            v-if="category.scope === 'custom'"
            class="cat-delete"
            type="button"
            @click="remove(category.id)"
          >
            {{ t('categories.delete') }}
          </button>
        </div>
      </div>
    </section>

    <EmptyState
      v-else
      icon="🏷"
      :title="t('categories.emptyTitle')"
      :description="t('categories.emptyDescription')"
    />

    <section class="section-card">
      <h2 class="section-title">{{ t('categories.newTitle') }}</h2>
      <form class="stack" @submit.prevent="submit">
        <div class="field">
          <label for="categoryName">{{ t('common.name') }}</label>
          <input id="categoryName" v-model="form.name" :placeholder="t('categories.namePlaceholder')" />
        </div>

        <div class="grid grid--two">
          <div class="field">
            <label for="categoryKind">{{ t('common.type') }}</label>
            <select id="categoryKind" v-model="form.kind">
              <option value="expense">{{ t('categoryKind.expense') }}</option>
              <option value="income">{{ t('categoryKind.income') }}</option>
            </select>
          </div>

          <div class="field">
            <label for="categoryColor">{{ t('common.color') }}</label>
            <div class="color-field">
              <input id="categoryColor" v-model="form.color" type="color" class="color-input" />
              <span class="color-preview" :style="{ background: form.color }" />
              <span class="color-value">{{ form.color }}</span>
            </div>
          </div>
        </div>

        <button class="button button--primary button--block" type="submit">{{ t('categories.create') }}</button>
      </form>
    </section>

    <section class="section-card">
      <p class="muted" style="margin: 0 0 12px; font-size: 0.875rem">{{ t('common.usedIn') }}</p>
      <div class="cta-list">
        <RouterLink class="button button--secondary" to="/transactions">{{ t('route.transactions') }}</RouterLink>
        <RouterLink class="button button--secondary" to="/review">{{ t('route.review') }}</RouterLink>
      </div>
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

.cat-list {
  display: flex;
  flex-direction: column;
}

.cat-row {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 13px 0;
  border-bottom: 0.5px solid var(--separator);
}

.cat-row:first-child { padding-top: 0; }
.cat-row:last-child { border-bottom: none; padding-bottom: 0; }

.cat-swatch {
  width: 38px;
  height: 38px;
  border-radius: var(--radius-sm);
  display: flex;
  align-items: center;
  justify-content: center;
  border: 1px solid transparent;
  flex-shrink: 0;
}

.cat-dot {
  width: 12px;
  height: 12px;
  border-radius: 50%;
}

.cat-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.cat-name {
  font-size: 0.9375rem;
  font-weight: 600;
  color: var(--text-primary);
  letter-spacing: -0.01em;
}

.cat-meta {
  font-size: 0.75rem;
  color: var(--text-muted);
  text-transform: capitalize;
}

.cat-delete {
  height: 34px;
  padding: 0 14px;
  border: 1px solid var(--border-strong);
  border-radius: var(--radius-pill);
  background: transparent;
  font-size: 0.8125rem;
  font-weight: 500;
  color: var(--expense);
  flex-shrink: 0;
  transition: background-color var(--duration-fast) var(--ease-out),
              transform var(--duration-fast) var(--ease-out);
}

.cat-delete:hover {
  background: var(--danger-soft);
}

.cat-delete:active {
  transform: scale(0.96);
}

.color-field {
  display: flex;
  align-items: center;
  gap: 10px;
  min-height: 48px;
  padding: 0 16px;
  border-radius: var(--radius-sm);
  border: 1px solid var(--border-strong);
  background: var(--surface-strong);
}

.color-input {
  width: 28px;
  height: 28px;
  border: none;
  padding: 0;
  background: transparent;
  cursor: pointer;
  border-radius: 6px;
  flex-shrink: 0;
}

.color-preview {
  width: 20px;
  height: 20px;
  border-radius: 50%;
  flex-shrink: 0;
  border: 2px solid var(--border);
  display: none;
}

.color-value {
  font-size: 0.875rem;
  color: var(--text-muted);
  font-family: ui-monospace, 'SF Mono', monospace;
}
</style>
