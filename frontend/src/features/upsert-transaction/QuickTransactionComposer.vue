<script setup lang="ts">
import { computed, reactive, watch } from 'vue'
import { useRoute } from 'vue-router'

import { useAppUiStore } from '@/app/stores/app-ui'
import { useFinanceStore } from '@/app/stores/finance'
import { parseAmountToMinor } from '@/shared/lib/money'
import { toIsoDate } from '@/shared/lib/date'

const emit = defineEmits<{
  submitted: []
}>()

const financeStore = useFinanceStore()
const appUiStore = useAppUiStore()
const route = useRoute()

const form = reactive<{
  accountId: string
  amount: string
  categoryId: string
  kind: 'income' | 'expense'
  note: string
  occurredAt: string
}>({
  kind: route.query.kind === 'income' ? 'income' : 'expense',
  amount: '',
  accountId: financeStore.filters.selectedAccountId ?? financeStore.accounts[0]?.id ?? '',
  categoryId: '',
  occurredAt: toIsoDate(new Date()),
  note: typeof route.query.note === 'string' ? route.query.note : '',
})

const availableCategories = computed(() =>
  financeStore.categories.filter((category) => category.kind === form.kind),
)

watch(
  availableCategories,
  (categories) => {
    const currentCategoryExists = categories.some((category) => category.id === form.categoryId)

    if (!currentCategoryExists) {
      form.categoryId = categories[0]?.id ?? ''
    }
  },
  { immediate: true },
)

function submit() {
  if (!form.accountId || !form.categoryId) {
    appUiStore.pushToast('Choose an account and category first.', 'warning')
    return
  }

  const amountMinor = parseAmountToMinor(form.amount)

  if (amountMinor <= 0) {
    appUiStore.pushToast('Enter a valid amount.', 'warning')
    return
  }

  financeStore.addTransaction({
    accountId: form.accountId,
    amountMinor,
    categoryId: form.categoryId,
    kind: form.kind,
    note: form.note,
    occurredAt: form.occurredAt,
  })

  appUiStore.pushToast('Transaction saved.', 'success')

  form.amount = ''
  form.note = ''
  form.occurredAt = toIsoDate(new Date())

  emit('submitted')
}
</script>

<template>
  <form class="surface-card stack" @submit.prevent="submit">
    <div class="split">
      <div>
        <h3>Quick add</h3>
        <p class="muted">Fast mobile-first capture with account and category defaults.</p>
      </div>
      <div class="chip-list">
        <button
          class="button"
          :class="form.kind === 'expense' ? 'button--primary' : 'button--secondary'"
          type="button"
          @click="form.kind = 'expense'"
        >
          Expense
        </button>
        <button
          class="button"
          :class="form.kind === 'income' ? 'button--primary' : 'button--secondary'"
          type="button"
          @click="form.kind = 'income'"
        >
          Income
        </button>
      </div>
    </div>

    <div class="grid grid--two">
      <div class="field">
        <label for="amount">Amount</label>
        <input id="amount" v-model="form.amount" inputmode="decimal" placeholder="0.00" />
      </div>

      <div class="field">
        <label for="account">Account</label>
        <select id="account" v-model="form.accountId">
          <option v-for="account in financeStore.accounts" :key="account.id" :value="account.id">
            {{ account.name }}
          </option>
        </select>
      </div>
    </div>

    <div class="grid grid--two">
      <div class="field">
        <label for="category">Category</label>
        <select id="category" v-model="form.categoryId">
          <option v-for="category in availableCategories" :key="category.id" :value="category.id">
            {{ category.name }}
          </option>
        </select>
      </div>

      <div class="field">
        <label for="occurredAt">Date</label>
        <input id="occurredAt" v-model="form.occurredAt" type="date" />
      </div>
    </div>

    <div class="field">
      <label for="note">Note</label>
      <input id="note" v-model="form.note" maxlength="120" placeholder="Optional note" />
    </div>

    <button class="button button--primary button--block" type="submit">Save transaction</button>
  </form>
</template>
