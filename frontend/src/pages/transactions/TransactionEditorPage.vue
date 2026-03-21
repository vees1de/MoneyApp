<script setup lang="ts">
import { useRouter } from 'vue-router'

import QuickTransactionComposer from '@/features/upsert-transaction/QuickTransactionComposer.vue'
import PageContainer from '@/shared/ui/PageContainer.vue'

const router = useRouter()

async function handleSubmitted() {
  await router.push('/transactions')
}
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
