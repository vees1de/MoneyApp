<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import { useRoute } from 'vue-router'

import { useAppUiStore } from '@/app/stores/app-ui'
import { useFinanceStore } from '@/app/stores/finance'
import { useI18n } from '@/shared/i18n'
import { toIsoDate } from '@/shared/lib/date'
import { parseAmountToMinor } from '@/shared/lib/money'

interface TransactionComposerValue {
  accountId: string
  amountMinor: number
  categoryId: string | null
  kind: 'income' | 'expense'
  note: string
  occurredAt: string
  title: string
}

const props = withDefaults(
  defineProps<{
    initialValue?: Partial<TransactionComposerValue> | null
    mode?: 'create' | 'edit'
    submitLabel?: string
    transactionId?: string | null
  }>(),
  {
    initialValue: null,
    mode: 'create',
    submitLabel: '',
    transactionId: null,
  },
)

const emit = defineEmits<{
  submitted: [transactionId: string]
}>()

const financeStore = useFinanceStore()
const appUiStore = useAppUiStore()
const route = useRoute()
const { t } = useI18n()

const presetAmounts = [500, 1000, 2500, 5000]
const detailsExpanded = ref(props.mode === 'edit')

const form = reactive<{
  accountId: string
  amount: string
  categoryId: string
  kind: 'income' | 'expense'
  note: string
  occurredAt: string
  title: string
}>({
  kind: 'expense',
  amount: '',
  accountId: '',
  categoryId: '',
  occurredAt: toIsoDate(new Date()),
  note: '',
  title: '',
})

function amountMinorToInput(amountMinor: number | undefined) {
  if (!amountMinor) {
    return ''
  }

  return (amountMinor / 100).toFixed(2).replace(/\.00$/, '')
}

function applyInitialValue() {
  const nextKind =
    props.initialValue?.kind ??
    (route.query.kind === 'income' ? 'income' : 'expense')

  form.kind = nextKind
  form.amount = amountMinorToInput(props.initialValue?.amountMinor)
  form.accountId =
    props.initialValue?.accountId ??
    financeStore.filters.selectedAccountId ??
    financeStore.accounts[0]?.id ??
    ''
  form.categoryId = props.initialValue?.categoryId ?? ''
  form.occurredAt = props.initialValue?.occurredAt
    ? toIsoDate(props.initialValue.occurredAt)
    : toIsoDate(new Date())
  form.note = props.initialValue?.note ?? (typeof route.query.note === 'string' ? route.query.note : '')
  form.title = props.initialValue?.title ?? ''
}

const initKey = computed(() =>
  [
    props.mode,
    props.transactionId ?? 'new',
    props.initialValue?.accountId ?? financeStore.filters.selectedAccountId ?? financeStore.accounts[0]?.id ?? '',
    props.initialValue?.categoryId ?? '',
    props.initialValue?.kind ?? route.query.kind ?? 'expense',
  ].join(':'),
)

const availableCategories = computed(() =>
  financeStore.categories.filter((category) => category.kind === form.kind),
)

const recentCategories = computed(() => financeStore.getRecentCategories(form.kind, 5))

const selectedAccount = computed(() => financeStore.getAccountById(form.accountId))

const saveLabel = computed(() => {
  if (props.submitLabel) {
    return props.submitLabel
  }

  if (props.mode === 'edit') {
    return form.kind === 'expense' ? t('transactionForm.updateExpense') : t('transactionForm.updateIncome')
  }

  return form.kind === 'expense' ? t('transactionForm.saveExpense') : t('transactionForm.saveIncome')
})

watch(initKey, applyInitialValue, { immediate: true })

watch(
  availableCategories,
  (categories) => {
    if (!categories.length) {
      form.categoryId = ''
      return
    }

    const currentCategoryExists = categories.some((category) => category.id === form.categoryId)
    if (!currentCategoryExists) {
      form.categoryId = categories[0]?.id ?? ''
    }
  },
  { immediate: true },
)

function applyPresetAmount(amount: number) {
  form.amount = amount.toString()
}

function selectRecentCategory(categoryId: string) {
  form.categoryId = categoryId
}

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
    const transaction =
      props.mode === 'edit' && props.transactionId
        ? await financeStore.updateTransactionEntry({
            id: props.transactionId,
            accountId: form.accountId,
            amountMinor,
            categoryId: form.categoryId,
            note: form.note,
            occurredAt: form.occurredAt,
            title: form.title,
            type: form.kind,
          })
        : await financeStore.addTransaction({
            accountId: form.accountId,
            amountMinor,
            categoryId: form.categoryId,
            kind: form.kind,
            note: form.note,
            occurredAt: form.occurredAt,
            title: form.title,
          })

    appUiStore.pushToast(
      props.mode === 'edit' ? t('transactionForm.updated') : t('transactionForm.saved'),
      'success',
    )

    if (props.mode === 'create') {
      form.amount = ''
      form.note = ''
      form.title = ''
      form.occurredAt = toIsoDate(new Date())
    }

    emit('submitted', transaction.id)
  } catch (error) {
    const message = error instanceof Error ? error.message : t('transactionForm.saveFailed')
    appUiStore.pushToast(message, 'warning')
  }
}
</script>

<template>
  <form class="composer" @submit.prevent="submit">
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

    <div class="composer__amount-wrap">
      <span class="composer__currency">{{ selectedAccount?.currency ?? 'RUB' }}</span>
      <input
        id="amount"
        v-model="form.amount"
        class="composer__amount-input"
        inputmode="decimal"
        :placeholder="t('transactionForm.placeholderAmount')"
      />
    </div>

    <div class="chip-group">
      <span class="tiny chip-group__label">{{ t('transactionForm.presetAmounts') }}</span>
      <div class="chip-group__items">
        <button
          v-for="amount in presetAmounts"
          :key="amount"
          class="chip chip--amount"
          type="button"
          @click="applyPresetAmount(amount)"
        >
          {{ amount }}
        </button>
      </div>
    </div>

    <div v-if="recentCategories.length" class="chip-group">
      <span class="tiny chip-group__label">{{ t('transactionForm.recentCategories') }}</span>
      <div class="chip-group__items">
        <button
          v-for="category in recentCategories"
          :key="category.id"
          class="chip"
          :class="{ 'chip--active': form.categoryId === category.id }"
          type="button"
          @click="selectRecentCategory(category.id)"
        >
          <span class="chip__dot" :style="{ background: category.color }" />
          {{ category.name }}
        </button>
      </div>
    </div>

    <div class="field">
      <label for="category">{{ t('common.categoryLabel') }}</label>
      <select id="category" v-model="form.categoryId">
        <option v-for="category in availableCategories" :key="category.id" :value="category.id">
          {{ category.name }}
        </option>
      </select>
    </div>

    <button
      v-if="props.mode === 'create'"
      class="composer__details-toggle"
      type="button"
      @click="detailsExpanded = !detailsExpanded"
    >
      {{ detailsExpanded ? t('transactionForm.hideDetails') : t('transactionForm.moreDetails') }}
    </button>

    <div v-if="detailsExpanded || props.mode === 'edit'" class="composer__details">
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
          <label for="occurredAt">{{ t('common.date') }}</label>
          <input id="occurredAt" v-model="form.occurredAt" type="date" />
        </div>
      </div>

      <div class="grid grid--two" style="gap:10px">
        <div class="field">
          <label for="title">{{ t('common.title') }}</label>
          <input
            id="title"
            v-model="form.title"
            maxlength="80"
            :placeholder="t('transactionForm.titlePlaceholder')"
          />
        </div>

        <div class="field">
          <label for="note">{{ t('common.note') }}</label>
          <input id="note" v-model="form.note" maxlength="120" :placeholder="t('common.optional')" />
        </div>
      </div>
    </div>

    <button
      class="button button--block"
      :class="form.kind === 'expense' ? 'composer__save--expense' : 'composer__save--income'"
      type="submit"
    >
      {{ saveLabel }}
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
  gap: 10px;
  padding: 0 16px;
  background: var(--surface-secondary);
  border-radius: var(--radius-lg);
  min-height: 56px;
}

.composer__currency {
  font-size: 0.875rem;
  font-weight: 700;
  color: var(--text-muted);
  flex-shrink: 0;
  letter-spacing: 0.06em;
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

.chip-group {
  display: grid;
  gap: 8px;
}

.chip-group__label {
  font-weight: 600;
}

.chip-group__items {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.chip {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  min-height: 34px;
  padding: 0 14px;
  border: 1px solid var(--border);
  border-radius: var(--radius-pill);
  background: var(--surface-secondary);
  color: var(--text-secondary);
  font-size: 0.875rem;
  font-weight: 600;
}

.chip--amount {
  background: var(--surface);
}

.chip--active {
  background: var(--brand-soft);
  border-color: rgba(0, 122, 255, 0.24);
  color: var(--brand);
}

.chip__dot {
  width: 9px;
  height: 9px;
  border-radius: 50%;
}

.composer__details-toggle {
  border: none;
  background: transparent;
  color: var(--brand);
  text-align: left;
  font-size: 0.875rem;
  font-weight: 700;
  padding: 0;
}

.composer__details {
  display: grid;
  gap: 10px;
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
