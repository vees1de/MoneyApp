<script setup lang="ts">
import type { Account } from '@/entities/account/model/types'
import { formatMoney } from '@/shared/lib/money'

defineProps<{
  accounts: Account[]
}>()
</script>

<template>
  <section class="surface-card stack">
    <div class="section-card__header">
      <div>
        <h3>Accounts</h3>
        <p class="muted">Balances across cards, cash, and savings buckets.</p>
      </div>
      <RouterLink class="button button--secondary" to="/accounts">Manage</RouterLink>
    </div>

    <div class="list">
      <div v-for="account in accounts" :key="account.id" class="list-item">
        <div class="stack" style="gap: 4px;">
          <strong>{{ account.name }}</strong>
          <span class="tiny">{{ account.type }}</span>
        </div>
        <strong>{{ formatMoney(account.balanceMinor, account.currency) }}</strong>
      </div>
    </div>
  </section>
</template>
