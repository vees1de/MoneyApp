<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'

import { useAppUiStore } from '@/app/stores/app-ui'
import { useFinanceStore } from '@/app/stores/finance'
import { useUserStore } from '@/app/stores/user'
import type { Account } from '@/entities/account/model/types'
import { translateAccountType, useI18n } from '@/shared/i18n'
import { formatDate } from '@/shared/lib/date'
import { formatMoney, parseAmountToMinor } from '@/shared/lib/money'
import EmptyState from '@/shared/ui/EmptyState.vue'
import PageContainer from '@/shared/ui/PageContainer.vue'

const financeStore = useFinanceStore()
const appUiStore = useAppUiStore()
const userStore = useUserStore()
const { t } = useI18n()

const form = reactive({
  balance: '',
  name: '',
  type: 'bank' as 'bank' | 'cash' | 'savings',
})

const editAccountId = ref<string | null>(null)
const editForm = reactive({
  name: '',
})

const typeIcon: Record<string, string> = { bank: '🏦', cash: '💵', savings: '🎯' }

function startEditing(account: Account) {
  editAccountId.value = account.id
  editForm.name = account.name
}

function stopEditing() {
  editAccountId.value = null
  editForm.name = ''
}

function submit() {
  if (!form.name.trim()) {
    appUiStore.pushToast(t('accounts.nameRequired'), 'warning')
    return
  }

  void (async () => {
    try {
      await financeStore.addAccount({
        name: form.name,
        type: form.type,
        balanceMinor: parseAmountToMinor(form.balance),
        currency: userStore.profile.currency || 'RUB',
      })
      appUiStore.pushToast(t('accounts.created'), 'success')
      form.balance = ''
      form.name = ''
      form.type = 'bank'
    } catch (error) {
      const message = error instanceof Error ? error.message : t('accounts.createFailed')
      appUiStore.pushToast(message, 'warning')
    }
  })()
}

function saveAccountEdit() {
  if (!editAccountId.value || !editForm.name.trim()) {
    appUiStore.pushToast(t('accounts.nameRequired'), 'warning')
    return
  }

  void (async () => {
    try {
      await financeStore.renameAccount({
        id: editAccountId.value as string,
        name: editForm.name,
      })
      appUiStore.pushToast(t('accounts.updated'), 'success')
      stopEditing()
    } catch (error) {
      const message = error instanceof Error ? error.message : t('accounts.updateFailed')
      appUiStore.pushToast(message, 'warning')
    }
  })()
}

function archiveAccount(accountId: string) {
  void (async () => {
    try {
      await financeStore.archiveAccount(accountId)
      appUiStore.pushToast(t('accounts.archived'), 'success')
      if (editAccountId.value === accountId) {
        stopEditing()
      }
    } catch (error) {
      const message = error instanceof Error ? error.message : t('accounts.archiveFailed')
      appUiStore.pushToast(message, 'warning')
    }
  })()
}

onMounted(async () => {
  await financeStore.fetchAccounts()
})
</script>

<template>
  <PageContainer>
    <div class="page-header">
      <h1>{{ t('accounts.title') }}</h1>
      <p class="muted">{{ t('accounts.subtitle') }}</p>
    </div>

    <section v-if="financeStore.accounts.length" class="section-card">
      <div class="acct-list">
        <div v-for="account in financeStore.accounts" :key="account.id" class="acct-item">
          <div class="acct-row">
            <div class="acct-icon">{{ typeIcon[account.type] ?? '🏦' }}</div>
            <div class="acct-body">
              <span class="acct-name">{{ account.name }}</span>
              <span class="acct-meta">
                {{ t('accounts.meta', { type: translateAccountType(account.type), date: formatDate(account.updatedAt) }) }}
              </span>
            </div>
            <strong class="acct-balance">{{ formatMoney(account.balanceMinor, account.currency) }}</strong>
          </div>

          <div class="acct-actions">
            <button class="acct-action-link" type="button" @click="startEditing(account)">
              {{ t('common.edit') }}
            </button>
            <RouterLink class="acct-action-link" :to="`/transactions?account=${account.id}`">
              {{ t('route.transactions') }}
            </RouterLink>
          </div>

          <form
            v-if="editAccountId === account.id"
            class="acct-edit"
            @submit.prevent="saveAccountEdit"
          >
            <div class="field">
              <label :for="`edit-account-${account.id}`">{{ t('common.name') }}</label>
              <input :id="`edit-account-${account.id}`" v-model="editForm.name" />
            </div>

            <p class="tiny" style="margin:0">{{ t('accounts.typeLockedHint') }}</p>

            <div class="acct-edit__actions">
              <button class="button button--secondary" type="button" @click="stopEditing">
                {{ t('common.cancel') }}
              </button>
              <button class="button button--secondary acct-edit__archive" type="button" @click="archiveAccount(account.id)">
                {{ t('common.archive') }}
              </button>
              <button class="button button--primary" type="submit">{{ t('common.save') }}</button>
            </div>
          </form>
        </div>
      </div>
    </section>

    <EmptyState
      v-else
      icon="🏦"
      :title="t('accounts.emptyTitle')"
      :description="t('accounts.emptyDescription')"
    />

    <section class="section-card">
      <h2 style="margin:0 0 16px;font-size:1.0625rem;font-weight:600;letter-spacing:-0.01em">{{ t('accounts.addTitle') }}</h2>
      <form class="stack" style="gap:14px" @submit.prevent="submit">
        <div class="field">
          <label for="accountName">{{ t('common.name') }}</label>
          <input id="accountName" v-model="form.name" :placeholder="t('accounts.namePlaceholder')" />
        </div>

        <div class="grid grid--two">
          <div class="field">
            <label for="accountType">{{ t('common.type') }}</label>
            <select id="accountType" v-model="form.type">
              <option value="bank">🏦 {{ t('accountType.bank') }}</option>
              <option value="cash">💵 {{ t('accountType.cash') }}</option>
              <option value="savings">🎯 {{ t('accountType.savings') }}</option>
            </select>
          </div>

          <div class="field">
            <label for="balance">{{ t('common.openingBalance') }}</label>
            <input id="balance" v-model="form.balance" inputmode="decimal" :placeholder="t('transactionForm.placeholderAmount')" />
          </div>
        </div>

        <button class="button button--primary button--block" type="submit">{{ t('accounts.create') }}</button>
      </form>
    </section>
  </PageContainer>
</template>

<style scoped>
.acct-list {
  display: flex;
  flex-direction: column;
}

.acct-item {
  padding: 12px 0;
  border-bottom: 1px solid var(--separator);
}

.acct-item:first-child { padding-top: 0; }
.acct-item:last-child { border-bottom: none; padding-bottom: 0; }

.acct-row {
  display: flex;
  align-items: center;
  gap: 12px;
}

.acct-icon {
  width: 38px;
  height: 38px;
  border-radius: var(--radius-sm);
  background: var(--surface-secondary);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.125rem;
  flex-shrink: 0;
}

.acct-body {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 1px;
}

.acct-name {
  font-size: 0.9375rem;
  font-weight: 600;
  color: var(--text-primary);
}

.acct-meta {
  font-size: 0.75rem;
  color: var(--text-muted);
  text-transform: capitalize;
}

.acct-balance {
  font-size: 0.9375rem;
  font-weight: 700;
  letter-spacing: -0.02em;
  white-space: nowrap;
}

.acct-actions {
  display: flex;
  gap: 16px;
  margin: 10px 0 0 50px;
}

.acct-action-link {
  padding: 0;
  border: none;
  background: transparent;
  color: var(--brand);
  font-size: 0.8125rem;
  font-weight: 700;
}

.acct-edit {
  display: grid;
  gap: 12px;
  margin-top: 12px;
  margin-left: 50px;
  padding: 14px;
  border-radius: var(--radius-lg);
  background: var(--surface-secondary);
}

.acct-edit__actions {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.acct-edit__archive {
  color: var(--expense);
}
</style>
