<script setup lang="ts">
import { useI18n } from '@/shared/i18n'

const { t } = useI18n()
</script>

<template>
  <div class="app-shell">
    <div class="app-shell__inner auth-layout">
      <section class="surface-card auth-hero stack">
        <div class="page-header">
          <h1>{{ t('notFound.title') }}</h1>
          <p class="muted">{{ t('notFound.body') }}</p>
        </div>

        <RouterLink class="button button--primary" to="/dashboard">{{ t('notFound.cta') }}</RouterLink>
      </section>
    </div>
  </div>
</template>
