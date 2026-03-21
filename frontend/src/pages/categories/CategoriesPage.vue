<script setup lang="ts">
import { reactive } from 'vue'

import { useAppUiStore } from '@/app/stores/app-ui'
import { useFinanceStore } from '@/app/stores/finance'
import EmptyState from '@/shared/ui/EmptyState.vue'
import PageContainer from '@/shared/ui/PageContainer.vue'
import SectionCard from '@/shared/ui/SectionCard.vue'

const financeStore = useFinanceStore()
const appUiStore = useAppUiStore()

const form = reactive({
  color: '#185d43',
  kind: 'expense' as 'income' | 'expense',
  name: '',
})

function submit() {
  if (!form.name.trim()) {
    appUiStore.pushToast('Category name is required.', 'warning')
    return
  }

  financeStore.addCategory({
    color: form.color,
    kind: form.kind,
    name: form.name,
  })

  appUiStore.pushToast('Category created.', 'success')
  form.name = ''
}

function remove(categoryId: string) {
  const removed = financeStore.deleteCategory(categoryId)
  appUiStore.pushToast(
    removed ? 'Category deleted.' : 'Category still has transactions and cannot be removed.',
    removed ? 'success' : 'warning',
  )
}
</script>

<template>
  <PageContainer>
    <div class="page-header">
      <h1>Categories</h1>
      <p class="muted">System and custom categories stay explicit because review quality depends on them.</p>
    </div>

    <div class="grid grid--two">
      <form class="surface-card stack" @submit.prevent="submit">
        <div class="page-header">
          <h2>New custom category</h2>
          <p class="muted">Keep creation quick and leave advanced taxonomy for later growth.</p>
        </div>

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
            <input id="categoryColor" v-model="form.color" type="color" />
          </div>
        </div>

        <button class="button button--primary button--block" type="submit">Create category</button>
      </form>

      <SectionCard>
        <div class="page-header">
          <h2>Why categories matter</h2>
          <p class="muted">
            Spending patterns, dashboard insights, and weekly discrepancy hints all depend on coherent categories.
          </p>
        </div>

        <div class="cta-list">
          <RouterLink class="button button--secondary" to="/transactions">Back to transactions</RouterLink>
          <RouterLink class="button button--secondary" to="/review">Use in review</RouterLink>
        </div>
      </SectionCard>
    </div>

    <SectionCard v-if="financeStore.categories.length">
      <div class="list">
        <div v-for="category in financeStore.categories" :key="category.id" class="list-item">
          <div class="row">
            <span
              :style="{
                width: '14px',
                height: '14px',
                borderRadius: '999px',
                background: category.color,
                display: 'inline-block',
              }"
            />
            <div class="stack" style="gap: 4px;">
              <strong>{{ category.name }}</strong>
              <span class="tiny">{{ category.kind }} · {{ category.scope }}</span>
            </div>
          </div>

          <button
            v-if="category.scope === 'custom'"
            class="button button--secondary"
            type="button"
            @click="remove(category.id)"
          >
            Delete
          </button>
        </div>
      </div>
    </SectionCard>

    <EmptyState
      v-else
      title="No categories yet"
      description="Create custom categories to make the dashboard and weekly review more useful."
    />
  </PageContainer>
</template>
