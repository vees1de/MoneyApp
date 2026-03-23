<script setup lang="ts">
import { onMounted } from 'vue'
import { useRouter } from 'vue-router'

import { useFinanceStore } from '@/app/stores/finance'
import TransferComposer from '@/features/transfer/TransferComposer.vue'
import { useI18n } from '@/shared/i18n'
import PageContainer from '@/shared/ui/PageContainer.vue'

const router = useRouter()
const financeStore = useFinanceStore()
const { t } = useI18n()

async function handleSubmitted() {
  await router.push('/transactions')
}

onMounted(async () => {
  await financeStore.fetchAccounts()
})
</script>

<template>
  <PageContainer>
    <div class="page-header">
      <h1>{{ t('transfer.title') }}</h1>
      <p class="muted">{{ t('transfer.subtitle') }}</p>
    </div>

    <TransferComposer @submitted="handleSubmitted" />
  </PageContainer>
</template>
