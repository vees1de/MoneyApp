<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import { useFinanceStore } from '@/app/stores/finance'
import type { Transaction } from '@/entities/transaction/model/types'
import TransferComposer from '@/features/transfer/TransferComposer.vue'
import QuickTransactionComposer from '@/features/upsert-transaction/QuickTransactionComposer.vue'
import { useI18n } from '@/shared/i18n'
import { formatDateTime } from '@/shared/lib/date'
import PageContainer from '@/shared/ui/PageContainer.vue'

const route = useRoute()
const router = useRouter()
const financeStore = useFinanceStore()
const { t } = useI18n()

const transaction = ref<Transaction | null>(null)
const pageLoading = ref(false)

const transactionId = computed(() =>
  typeof route.params.id === 'string' ? route.params.id : null,
)

const isEditing = computed(() => Boolean(transactionId.value))
const isTransfer = computed(() => transaction.value?.type === 'transfer')

const pageTitle = computed(() => {
  if (!isEditing.value) {
    return t('transactions.addTitle')
  }

  return isTransfer.value ? t('transactions.transferEditTitle') : t('transactions.editTitle')
})

const pageSubtitle = computed(() => {
  if (!isEditing.value) {
    return t('transactions.addSubtitle')
  }

  return isTransfer.value ? t('transactions.transferEditSubtitle') : t('transactions.editSubtitle')
})

const transactionFormValue = computed(() =>
  transaction.value && !isTransfer.value
    ? {
        accountId: transaction.value.accountId,
        amountMinor: transaction.value.amountMinor,
        categoryId: transaction.value.categoryId,
        kind: transaction.value.type === 'income' ? ('income' as const) : ('expense' as const),
        note: transaction.value.note,
        occurredAt: transaction.value.occurredAt,
        title: transaction.value.title,
      }
    : null,
)

const transferFormValue = computed(() =>
  transaction.value && isTransfer.value
    ? {
        amountMinor: transaction.value.amountMinor,
        fromAccountId: transaction.value.accountId,
        note: transaction.value.note,
        occurredAt: transaction.value.occurredAt,
        title: transaction.value.title,
        toAccountId: transaction.value.transferAccountId ?? '',
      }
    : null,
)

async function loadScreen() {
  pageLoading.value = true

  try {
    await Promise.all([financeStore.fetchAccounts(), financeStore.fetchCategories()])

    if (transactionId.value) {
      transaction.value = await financeStore.fetchTransaction(transactionId.value)
    } else {
      transaction.value = null
    }
  } finally {
    pageLoading.value = false
  }
}

async function handleSubmitted() {
  await router.push('/transactions')
}

async function handleDelete() {
  if (!transactionId.value) {
    return
  }

  financeStore.queueTransactionDeletion(transactionId.value)
  await router.push('/transactions')
}

watch(transactionId, () => {
  void loadScreen()
})

onMounted(async () => {
  await loadScreen()
})
</script>

<template>
  <PageContainer>
    <div class="page-header">
      <h1>{{ pageTitle }}</h1>
      <p class="muted">{{ pageSubtitle }}</p>
    </div>

    <section v-if="pageLoading" class="section-card">
      <p class="muted" style="margin: 0">{{ t('common.loading') }}</p>
    </section>

    <template v-else>
      <section v-if="isEditing && transaction" class="section-card editor-meta">
        <div>
          <strong class="editor-meta__title">
            {{ transaction.title || transaction.note || t('transactions.untitled') }}
          </strong>
          <p class="tiny" style="margin: 4px 0 0">
            {{ t('common.lastUpdated', { date: formatDateTime(transaction.updatedAt) }) }}
          </p>
        </div>

        <button class="button button--secondary editor-meta__delete" type="button" @click="handleDelete">
          {{ t('transactions.delete') }}
        </button>
      </section>

      <TransferComposer
        v-if="isTransfer"
        :initial-value="transferFormValue"
        :mode="isEditing ? 'edit' : 'create'"
        :submit-label="isEditing ? t('transfer.update') : t('transfer.save')"
        :transaction-id="transactionId"
        @submitted="handleSubmitted"
      />

      <QuickTransactionComposer
        v-else
        :initial-value="transactionFormValue"
        :mode="isEditing ? 'edit' : 'create'"
        :transaction-id="transactionId"
        @submitted="handleSubmitted"
      />
    </template>
  </PageContainer>
</template>

<style scoped>
.editor-meta {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
}

.editor-meta__title {
  display: block;
  font-size: 0.9375rem;
  font-weight: 700;
  letter-spacing: -0.01em;
}

.editor-meta__delete {
  color: var(--expense);
}
</style>
