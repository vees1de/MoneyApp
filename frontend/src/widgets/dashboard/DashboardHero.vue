<script setup lang="ts">
import { computed } from 'vue'

import { useI18n } from '@/shared/i18n'

const props = defineProps<{
  balance: string
  reviewStatus: string
  safeToSpend: string
}>()

const { t } = useI18n()

const isReviewResolved = computed(() => props.reviewStatus === 'matched' || props.reviewStatus === 'resolved')
</script>

<template>
  <section class="surface-card hero-card">
    <div class="hero-card__inner">
      <div class="hero-card__eyebrow">
        <span class="hero-eyebrow-label">{{ t('dashboard.totalBalance') }}</span>
        <span class="hero-badge">Finance Core</span>
      </div>

      <strong class="hero-card__amount">{{ props.balance }}</strong>

      <p class="hero-card__sub muted">{{ t('dashboard.heroSubtitle') }}</p>
    </div>

    <div class="hero-card__meta">
      <div class="hero-metric">
        <span class="hero-metric__label">{{ t('dashboard.safeToSpend') }}</span>
        <span class="hero-metric__value">{{ props.safeToSpend }}</span>
      </div>
      <div class="hero-metric__divider" />
      <div class="hero-metric">
        <span class="hero-metric__label">{{ t('dashboard.review') }}</span>
        <span
          class="hero-metric__value"
          :class="isReviewResolved ? 'hero-metric__value--green' : 'hero-metric__value--amber'"
        >{{ isReviewResolved ? t('dashboard.statusDone') : t('dashboard.statusPending') }}</span>
      </div>
    </div>
  </section>
</template>

<style scoped>
.hero-card__inner {
  display: flex;
  flex-direction: column;
  gap: 10px;
  position: relative;
  z-index: 1;
}

.hero-card__eyebrow {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}

.hero-eyebrow-label {
  font-size: 0.8125rem;
  font-weight: 500;
  color: rgba(255,255,255,0.55);
  letter-spacing: 0.02em;
}

.hero-badge {
  font-size: 0.75rem;
  font-weight: 600;
  padding: 3px 9px;
  border-radius: 999px;
  background: rgba(255,255,255,0.10);
  color: rgba(255,255,255,0.70);
  border: 1px solid rgba(255,255,255,0.08);
  letter-spacing: 0.02em;
}

.hero-card__sub {
  margin: 0;
  font-size: 0.8125rem;
}

.hero-card__meta {
  position: relative;
  z-index: 1;
  display: flex;
  align-items: center;
  gap: 16px;
  margin-top: 20px;
  padding-top: 18px;
  border-top: 1px solid rgba(255,255,255,0.10);
}

.hero-metric {
  display: flex;
  flex-direction: column;
  gap: 3px;
}

.hero-metric__label {
  font-size: 0.6875rem;
  font-weight: 500;
  color: rgba(255,255,255,0.45);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.hero-metric__value {
  font-size: 1rem;
  font-weight: 600;
  color: rgba(255,255,255,0.90);
  letter-spacing: -0.01em;
}

.hero-metric__value--green {
  color: #5CDB8F;
}

.hero-metric__value--amber {
  color: #FFBE5C;
}

.hero-metric__divider {
  width: 1px;
  height: 32px;
  background: rgba(255,255,255,0.12);
  flex-shrink: 0;
}
</style>
