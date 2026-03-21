<script setup lang="ts">
import { computed, reactive, watch } from 'vue'
import { useRoute } from 'vue-router'

import { useAppUiStore } from '@/app/stores/app-ui'
import { useFinanceStore } from '@/app/stores/finance'
import { useI18n } from '@/shared/i18n'
import { parseAmountToMinor } from '@/shared/lib/money'
import { toIsoDate } from '@/shared/lib/date'

const emit = defineEmits<{
  submitted: []
}>()

const financeStore = useFinanceStore()
const appUiStore = useAppUiStore()
const route = useRoute()
const { t } = useI18n()

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

watch(
  () => financeStore.accounts,
  (accounts) => {
    if (!form.accountId && accounts[0]) {
      form.accountId = accounts[0].id
    }
  },
  { deep: true, immediate: true },
)

async function submit() {
  if (!form.accountId || !form.categoryId) {
    appUiStore.pushToast(t('transactionForm.accountCategoryRequired'), 'warning')
    return
  }

  const amountMinor = parseAmountToMinor(form.amount)
  if (amountMinor <= 0) {
    appUiStore.pushToast(t('transactionForm.amountRequired'), 'warning')
    return
  }

  try {
    await financeStore.addTransaction({
      accountId: form.accountId,
      amountMinor,
      categoryId: form.categoryId,
      kind: form.kind,
      note: form.note,
      occurredAt: form.occurredAt,
    })

    appUiStore.pushToast(t('transactionForm.saved'), 'success')
    form.amount = ''
    form.note = ''
    form.occurredAt = toIsoDate(new Date())
    emit('submitted')
  } catch (error) {
    const message = error instanceof Error ? error.message : t('transactionForm.saveFailed')
    appUiStore.pushToast(message, 'warning')
  }
}
</script>

<template>
  <form class="composer" @submit.prevent="submit">
    <!-- Kind toggle -->
    <div class="composer__kind-row">
      <button
        class="kind-btn"
        :class="{ 'kind-btn--active kind-btn--expense': form.kind === 'expense' }"
        type="button"
        @click="form.kind = 'expense'"
      >
        <span class="kind-btn__icon">−</span>
        {{ t('transactionForm.expense') }}
      </button>
      <button
        class="kind-btn"
        :class="{ 'kind-btn--active kind-btn--income': form.kind === 'income' }"
        type="button"
        @click="form.kind = 'income'"
      >
        <span class="kind-btn__icon">+</span>
        {{ t('transactionForm.income') }}
      </button>
    </div>

    <!-- Amount (hero input) -->
    <div class="composer__amount-wrap">
      <span class="composer__currency">₽</span>
      <input
        id="amount"
        v-model="form.amount"
        class="composer__amount-input"
        inputmode="decimal"
        :placeholder="t('transactionForm.placeholderAmount')"
      />
    </div>

    <!-- Secondary fields -->
    <div class="grid grid--two" style="gap:10px">
      <div class="field">
        <label for="account">{{ t('common.account') }}</label>
        <select id="account" v-model="form.accountId">
          <option v-for="account in financeStore.accounts" :key="account.id" :value="account.id">
            {{ account.name }}
          </option>
        </select>
      </div>

      <div class="field">
        <label for="category">{{ t('common.categoryLabel') }}</label>
        <select id="category" v-model="form.categoryId">
          <option v-for="category in availableCategories" :key="category.id" :value="category.id">
            {{ category.name }}
          </option>
        </select>
      </div>
    </div>

    <div class="grid grid--two" style="gap:10px">
      <div class="field">
        <label for="occurredAt">{{ t('common.date') }}</label>
        <input id="occurredAt" v-model="form.occurredAt" type="date" />
      </div>

      <div class="field">
        <label for="note">{{ t('common.note') }}</label>
        <input id="note" v-model="form.note" maxlength="120" :placeholder="t('common.optional')" />
      </div>
    </div>

    <button
      class="button button--block"
      :class="form.kind === 'expense' ? 'composer__save--expense' : 'composer__save--income'"
      type="submit"
    >
      {{ form.kind === 'expense' ? t('transactionForm.saveExpense') : t('transactionForm.saveIncome') }}
    </button>
  </form>
</template>

<style scoped>
.composer {
  background: var(--surface);
  border-radius: var(--radius-xl);
  border: 1px solid var(--border);
  box-shadow: var(--shadow);
  padding: 20px;
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.composer__kind-row {
  display: flex;
  gap: 6px;
}

.kind-btn {
  flex: 1;
  height: 40px;
  border: 1px solid var(--border-strong);
  border-radius: var(--radius-pill);
  background: transparent;
  font-size: 0.9rem;
  font-weight: 600;
  color: var(--text-secondary);
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  transition: background var(--duration-fast) ease, color var(--duration-fast) ease, border-color var(--duration-fast) ease;
}

.kind-btn__icon {
  font-size: 1rem;
  font-weight: 700;
  line-height: 1;
}

.kind-btn--active {
  border-color: transparent;
}

.kind-btn--expense {
  background: var(--expense);
  color: #fff;
}

.kind-btn--income {
  background: var(--income);
  color: #fff;
}

.composer__amount-wrap {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 0 16px;
  background: var(--surface-secondary);
  border-radius: var(--radius-lg);
  min-height: 56px;
}

.composer__currency {
  font-size: 1.5rem;
  font-weight: 300;
  color: var(--text-muted);
  flex-shrink: 0;
}

.composer__amount-input {
  flex: 1;
  border: none;
  background: transparent;
  font-size: 1.75rem;
  font-weight: 700;
  letter-spacing: -0.04em;
  color: var(--text-primary);
  min-width: 0;
}

.composer__amount-input::placeholder {
  color: var(--text-quaternary);
}

.composer__amount-input:focus {
  outline: none;
  box-shadow: none;
}

.composer__save--expense {
  background: var(--expense);
  color: #fff;
}

.composer__save--income {
  background: var(--income);
  color: #fff;
}
</style>
