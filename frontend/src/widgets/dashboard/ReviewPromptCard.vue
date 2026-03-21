<script setup lang="ts">
import { useI18n } from '@/shared/i18n'

defineProps<{
  actionLabel: string
  body: string
  tone: 'success' | 'warning'
}>()

const { t } = useI18n()
</script>

<template>
  <section class="review-prompt" :class="`review-prompt--${tone}`">
    <div class="review-prompt__header">
      <div class="review-prompt__icon">
        <svg v-if="tone === 'success'" viewBox="0 0 24 24" fill="currentColor" width="18" height="18">
          <path fill-rule="evenodd" d="M2.25 12c0-5.385 4.365-9.75 9.75-9.75s9.75 4.365 9.75 9.75-4.365 9.75-9.75 9.75S2.25 17.385 2.25 12Zm13.36-1.814a.75.75 0 1 0-1.22-.872l-3.236 4.53L9.53 12.22a.75.75 0 0 0-1.06 1.06l2.25 2.25a.75.75 0 0 0 1.14-.094l3.75-5.25Z" clip-rule="evenodd"/>
        </svg>
        <svg v-else viewBox="0 0 24 24" fill="currentColor" width="18" height="18">
          <path fill-rule="evenodd" d="M9.401 3.003c1.155-2 4.043-2 5.197 0l7.355 12.748c1.154 2-.29 4.5-2.599 4.5H4.645c-2.309 0-3.752-2.5-2.598-4.5L9.4 3.003ZM12 8.25a.75.75 0 0 1 .75.75v3.75a.75.75 0 0 1-1.5 0V9a.75.75 0 0 1 .75-.75Zm0 8.25a.75.75 0 1 0 0-1.5.75.75 0 0 0 0 1.5Z" clip-rule="evenodd"/>
        </svg>
      </div>
      <div class="review-prompt__label">{{ t('dashboard.reviewPromptTitle') }}</div>
      <span class="review-prompt__badge">
        {{ tone === 'success' ? t('dashboard.reviewHealthyBadge') : t('dashboard.reviewWarningBadge') }}
      </span>
    </div>

    <p class="review-prompt__body">{{ body }}</p>

    <RouterLink class="review-prompt__cta" to="/review">
      {{ actionLabel }}
    </RouterLink>
  </section>
</template>

<style scoped>
.review-prompt {
  padding: 18px 20px;
  border-radius: var(--radius-xl);
  border: 1px solid;
  display: flex;
  flex-direction: column;
  gap: 10px;
  box-shadow: var(--shadow-xs);
}

.review-prompt--warning {
  background: rgba(255, 149, 0, 0.06);
  border-color: rgba(255, 149, 0, 0.20);
}

.review-prompt--success {
  background: rgba(52, 199, 89, 0.06);
  border-color: rgba(52, 199, 89, 0.18);
}

.review-prompt__header {
  display: flex;
  align-items: center;
  gap: 8px;
}

.review-prompt__icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 30px;
  height: 30px;
  border-radius: 50%;
  flex-shrink: 0;
}

.review-prompt--warning .review-prompt__icon {
  background: var(--warning-soft);
  color: var(--warning);
}

.review-prompt--success .review-prompt__icon {
  background: var(--income-soft);
  color: var(--income);
}

.review-prompt__label {
  font-size: 0.9375rem;
  font-weight: 600;
  color: var(--text-primary);
  flex: 1;
}

.review-prompt__badge {
  font-size: 0.75rem;
  font-weight: 600;
  padding: 3px 9px;
  border-radius: var(--radius-pill);
}

.review-prompt--warning .review-prompt__badge {
  background: var(--warning-soft);
  color: var(--warning);
}

.review-prompt--success .review-prompt__badge {
  background: var(--income-soft);
  color: var(--income);
}

.review-prompt__body {
  margin: 0;
  font-size: 0.875rem;
  color: var(--text-secondary);
  line-height: 1.45;
}

.review-prompt__cta {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  height: 40px;
  padding: 0 16px;
  border-radius: var(--radius-pill);
  font-size: 0.9rem;
  font-weight: 600;
  align-self: flex-start;
  transition: opacity var(--duration-fast) ease;
}

.review-prompt--warning .review-prompt__cta {
  background: var(--warning);
  color: #fff;
}

.review-prompt--success .review-prompt__cta {
  background: var(--income-soft);
  color: var(--income);
}

.review-prompt__cta:active {
  opacity: 0.80;
}
</style>
