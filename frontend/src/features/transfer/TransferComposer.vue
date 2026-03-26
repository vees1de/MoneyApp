<script setup lang="ts">
import { computed, reactive, watch } from 'vue'

import { useAppUiStore } from '@/app/stores/app-ui'
import { useFinanceStore } from '@/app/stores/finance'
import { useI18n } from '@/shared/i18n'
import { toIsoDate } from '@/shared/lib/date'
import { parseAmountToMinor } from '@/shared/lib/money'

interface TransferComposerValue {
  amountMinor: number
  fromAccountId: string
  note: string
  occurredAt: string
  title: string
  toAccountId: string
}

const props = withDefaults(
  defineProps<{
    initialValue?: Partial<TransferComposerValue> | null
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
const { t } = useI18n()

const form = reactive({
  amount: '',
  fromAccountId: '',
  note: '',
  occurredAt: toIsoDate(new Date()),
  title: '',
  toAccountId: '',
})

function amountMinorToInput(amountMinor: number | undefined) {
  if (!amountMinor) {
    return ''
  }

  return (amountMinor / 100).toFixed(2).replace(/\.00$/, '')
}

function applyInitialValue() {
  form.amount = amountMinorToInput(props.initialValue?.amountMinor)
  form.fromAccountId =
    props.initialValue?.fromAccountId ??
    financeStore.filters.selectedAccountId ??
    financeStore.accounts[0]?.id ??
    ''
  form.toAccountId = props.initialValue?.toAccountId ?? financeStore.accounts[1]?.id ?? ''
  form.occurredAt = props.initialValue?.occurredAt
    ? toIsoDate(props.initialValue.occurredAt)
    : toIsoDate(new Date())
  form.note = props.initialValue?.note ?? ''
  form.title = props.initialValue?.title ?? ''
}

watch(
  () => [
    props.mode,
    props.transactionId,
    props.initialValue?.fromAccountId,
    props.initialValue?.toAccountId,
    financeStore.accounts[0]?.id,
    financeStore.accounts[1]?.id,
  ],
  applyInitialValue,
  { immediate: true },
)

const fromAccount = computed(() => financeStore.getAccountById(form.fromAccountId))

const availableDestinationAccounts = computed(() =>
  financeStore.accounts.filter((account) => account.id !== form.fromAccountId),
)

const saveLabel = computed(() => {
  if (props.submitLabel) {
    return props.submitLabel
  }

  return props.mode === 'edit' ? t('transfer.update') : t('transfer.save')
})

function swapAccounts() {
  const nextFrom = form.toAccountId
  form.toAccountId = form.fromAccountId
  form.fromAccountId = nextFrom
}

async function submit() {
  if (!form.fromAccountId || !form.toAccountId) {
    appUiStore.pushToast(t('transfer.accountRequired'), 'warning')
    return
  }

  if (form.fromAccountId === form.toAccountId) {
    appUiStore.pushToast(t('transfer.sameAccount'), 'warning')
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
            accountId: form.fromAccountId,
            amountMinor,
            categoryId: null,
            note: form.note,
            occurredAt: form.occurredAt,
            title: form.title,
            transferAccountId: form.toAccountId,
            type: 'transfer',
          })
        : await financeStore.addTransfer({
            amountMinor,
            fromAccountId: form.fromAccountId,
            note: form.note,
            occurredAt: form.occurredAt,
            title: form.title,
            toAccountId: form.toAccountId,
          })

    appUiStore.pushToast(props.mode === 'edit' ? t('transfer.updated') : t('transfer.saved'), 'success')

    if (props.mode === 'create') {
      form.amount = ''
      form.note = ''
      form.title = ''
      form.occurredAt = toIsoDate(new Date())
    }

    emit('submitted', transaction.id)
  } catch (error) {
    const message = error instanceof Error ? error.message : t('transfer.failed')
    appUiStore.pushToast(message, 'warning')
  }
}
</script>

<template>
  <form class="transfer-form" @submit.prevent="submit">
    <div class="transfer-form__hero">
      <span class="transfer-form__currency">{{ fromAccount?.currency ?? 'RUB' }}</span>
      <input
        v-model="form.amount"
        class="transfer-form__amount"
        inputmode="decimal"
        :placeholder="t('transactionForm.placeholderAmount')"
      />
    </div>

    <div class="field">
      <label for="fromAccount">{{ t('transfer.fromAccount') }}</label>
      <select id="fromAccount" v-model="form.fromAccountId">
        <option v-for="account in financeStore.accounts" :key="account.id" :value="account.id">
          {{ account.name }}
        </option>
      </select>
    </div>

    <div class="transfer-form__swap">
      <button class="button button--secondary" type="button" @click="swapAccounts">
        {{ t('transfer.swap') }}
      </button>
    </div>

    <div class="field">
      <label for="toAccount">{{ t('transfer.toAccount') }}</label>
      <select id="toAccount" v-model="form.toAccountId">
        <option v-for="account in availableDestinationAccounts" :key="account.id" :value="account.id">
          {{ account.name }}
        </option>
      </select>
    </div>

    <div class="grid grid--two" style="gap: 10px">
      <div class="field">
        <label for="title">{{ t('common.title') }}</label>
        <input
          id="title"
          v-model="form.title"
          maxlength="80"
          :placeholder="t('transfer.titlePlaceholder')"
        />
      </div>

      <div class="field">
        <label for="occurredAt">{{ t('common.date') }}</label>
        <input id="occurredAt" v-model="form.occurredAt" type="date" />
      </div>
    </div>

    <div class="field">
      <label for="note">{{ t('common.note') }}</label>
      <input id="note" v-model="form.note" maxlength="120" :placeholder="t('common.optional')" />
    </div>

    <button class="button button--primary button--block" type="submit">
      {{ saveLabel }}
    </button>
  </form>
</template>

<style scoped>
.transfer-form {
  display: grid;
  gap: 16px;
  background: var(--surface);
  border-radius: var(--radius-xl);
  border: 1px solid var(--border);
  box-shadow: var(--shadow);
  padding: 22px;
}

.transfer-form__hero {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 0 18px;
  background: var(--surface-secondary);
  border-radius: var(--radius-lg);
  min-height: 60px;
}

.transfer-form__currency {
  font-size: 0.875rem;
  font-weight: 700;
  color: var(--text-muted);
  letter-spacing: 0.06em;
}

.transfer-form__amount {
  flex: 1;
  border: none;
  background: transparent;
  font-size: 1.75rem;
  font-weight: 700;
  letter-spacing: -0.04em;
  font-variant-numeric: tabular-nums;
}

.transfer-form__amount:focus {
  outline: none;
  box-shadow: none;
}

.transfer-form__swap {
  display: flex;
  justify-content: center;
}
</style>
