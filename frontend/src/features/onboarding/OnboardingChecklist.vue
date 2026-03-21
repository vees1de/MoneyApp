<script setup lang="ts">
import { useI18n } from '@/shared/i18n'

const { t } = useI18n()
</script>

<template>
  <div class="stack">
    <div class="row">
      <span class="pill pill--success">1</span>
      <div>
        <strong>{{ t('onboarding.checklist.defaultsTitle') }}</strong>
        <div class="tiny">{{ t('onboarding.checklist.defaultsBody') }}</div>
      </div>
    </div>
    <div class="row">
      <span class="pill pill--success">2</span>
      <div>
        <strong>{{ t('onboarding.checklist.accountTitle') }}</strong>
        <div class="tiny">{{ t('onboarding.checklist.accountBody') }}</div>
      </div>
    </div>
    <div class="row">
      <span class="pill pill--success">3</span>
      <div>
        <strong>{{ t('onboarding.checklist.dashboardTitle') }}</strong>
        <div class="tiny">{{ t('onboarding.checklist.dashboardBody') }}</div>
      </div>
    </div>
  </div>
</template>
