<script setup lang="ts">
defineProps<{
  balance: string
  reviewStatus: string
  safeToSpend: string
}>()
</script>

<template>
  <section class="surface-card hero-card stack">
    <span class="pill">Personal Life OS / Finance Core</span>
    <div class="stack">
      <span class="tiny">Current balance</span>
      <strong class="hero-card__amount">{{ balance }}</strong>
      <p class="muted">
        A single place to see what is available now, what can be spent safely, and what needs review.
      </p>
    </div>

    <div class="hero-card__meta">
      <span class="pill">Safe to spend: {{ safeToSpend }}</span>
      <span class="pill">Weekly review: {{ reviewStatus }}</span>
    </div>
  </section>
</template>
