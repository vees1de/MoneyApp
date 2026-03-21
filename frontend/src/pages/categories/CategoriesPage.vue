<script setup lang="ts">
import { onMounted, reactive } from 'vue'

import { useAppUiStore } from '@/app/stores/app-ui'
import { useFinanceStore } from '@/app/stores/finance'
import EmptyState from '@/shared/ui/EmptyState.vue'
import PageContainer from '@/shared/ui/PageContainer.vue'

const financeStore = useFinanceStore()
const appUiStore = useAppUiStore()

const form = reactive({
  color: '#007AFF',
  kind: 'expense' as 'income' | 'expense',
  name: '',
})

function submit() {
  if (!form.name.trim()) {
    appUiStore.pushToast('Category name is required.', 'warning')
    return
  }

  void (async () => {
    try {
      await financeStore.addCategory({ color: form.color, kind: form.kind, name: form.name })
      appUiStore.pushToast('Category created.', 'success')
      form.name = ''
    } catch (error) {
      const message = error instanceof Error ? error.message : 'Category creation failed.'
      appUiStore.pushToast(message, 'warning')
    }
  })()
}

function remove(categoryId: string) {
  void (async () => {
    try {
      await financeStore.removeCategory(categoryId)
      appUiStore.pushToast('Category deleted.', 'success')
    } catch (error) {
      const message = error instanceof Error ? error.message : 'Category still has transactions.'
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
      <h1>Categories</h1>
      <p class="muted">Organize your transactions for cleaner insights and review.</p>
    </div>

    <!-- Category list -->
    <section v-if="financeStore.categories.length" class="section-card">
      <div class="cat-list">
        <div v-for="category in financeStore.categories" :key="category.id" class="cat-row">
          <div class="cat-swatch" :style="{ background: `${category.color}20`, borderColor: `${category.color}30` }">
            <div class="cat-dot" :style="{ background: category.color }" />
          </div>
          <div class="cat-info">
            <span class="cat-name">{{ category.name }}</span>
            <span class="cat-meta">{{ category.kind }} · {{ category.scope }}</span>
          </div>
          <button
            v-if="category.scope === 'custom'"
            class="cat-delete"
            type="button"
            @click="remove(category.id)"
          >
            Delete
          </button>
        </div>
      </div>
    </section>

    <EmptyState
      v-else
      icon="🏷"
      title="No categories yet"
      description="Create custom categories to organize your transactions."
    />

    <!-- New category form -->
    <section class="section-card">
      <h2 style="margin:0 0 16px;font-size:1.0625rem;font-weight:600;letter-spacing:-0.01em">New category</h2>
      <form class="stack" style="gap:14px" @submit.prevent="submit">
        <div class="field">
          <label for="categoryName">Name</label>
          <input id="categoryName" v-model="form.name" placeholder="Subscriptions" />
        </div>

        <div class="grid grid--two">
          <div class="field">
            <label for="categoryKind">Type</label>
            <select id="categoryKind" v-model="form.kind">
              <option value="expense">Expense</option>
              <option value="income">Income</option>
            </select>
          </div>

          <div class="field">
            <label for="categoryColor">Color</label>
            <div class="color-field">
              <input id="categoryColor" v-model="form.color" type="color" class="color-input" />
              <span class="color-preview" :style="{ background: form.color }" />
              <span class="color-value">{{ form.color }}</span>
            </div>
          </div>
        </div>

        <button class="button button--primary button--block" type="submit">Create category</button>
      </form>
    </section>

    <section class="section-card">
      <p class="muted" style="margin:0 0 12px;font-size:0.875rem">Used in</p>
      <div class="cta-list">
        <RouterLink class="button button--secondary" to="/transactions">Transactions</RouterLink>
        <RouterLink class="button button--secondary" to="/review">Weekly review</RouterLink>
      </div>
    </section>
  </PageContainer>
</template>

<style scoped>
.cat-list {
  display: flex;
  flex-direction: column;
}

.cat-row {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 0;
  border-bottom: 1px solid var(--separator);
}

.cat-row:first-child { padding-top: 0; }
.cat-row:last-child { border-bottom: none; padding-bottom: 0; }

.cat-swatch {
  width: 36px;
  height: 36px;
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
  gap: 1px;
}

.cat-name {
  font-size: 0.9375rem;
  font-weight: 600;
  color: var(--text-primary);
}

.cat-meta {
  font-size: 0.75rem;
  color: var(--text-muted);
  text-transform: capitalize;
}

.cat-delete {
  height: 32px;
  padding: 0 12px;
  border: 1px solid var(--border-strong);
  border-radius: var(--radius-pill);
  background: transparent;
  font-size: 0.8125rem;
  font-weight: 500;
  color: var(--expense);
  transition: background var(--duration-fast) ease;
  flex-shrink: 0;
}

.cat-delete:active {
  background: var(--danger-soft);
}

/* Color picker field */
.color-field {
  display: flex;
  align-items: center;
  gap: 10px;
  min-height: 46px;
  padding: 0 14px;
  border-radius: var(--radius-md);
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
  border-radius: 4px;
  flex-shrink: 0;
}

.color-preview {
  width: 20px;
  height: 20px;
  border-radius: 50%;
  flex-shrink: 0;
  border: 2px solid rgba(0,0,0,0.06);
  display: none;
}

.color-value {
  font-size: 0.875rem;
  color: var(--text-muted);
  font-family: ui-monospace, monospace;
}
</style>
