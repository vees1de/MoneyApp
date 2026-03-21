<script setup lang="ts">
import { onMounted } from 'vue'
import { useRouter } from 'vue-router'

import { useFinanceStore } from '@/app/stores/finance'
import QuickTransactionComposer from '@/features/upsert-transaction/QuickTransactionComposer.vue'
import PageContainer from '@/shared/ui/PageContainer.vue'

const router = useRouter()
const financeStore = useFinanceStore()

async function handleSubmitted() {
  await router.push('/transactions')
}

onMounted(async () => {
  await Promise.all([financeStore.fetchAccounts(), financeStore.fetchCategories()])
})
</script>

<template>
  <PageContainer>
    <div class="page-header">
      <h1>Add transaction</h1>
      <p class="muted">Dedicated full-screen entry flow for one-hand capture and review jump-ins.</p>
    </div>

    <QuickTransactionComposer @submitted="handleSubmitted" />
  </PageContainer>
</template>
