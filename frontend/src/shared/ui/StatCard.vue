<script setup lang="ts">
defineProps<{
  hint?: string
  label: string
  value: string
}>()
</script>

<template>
  <article class="stat-card">
    <span class="stat-card__label">{{ label }}</span>
    <strong class="stat-card__value">{{ value }}</strong>
    <span v-if="hint" class="stat-card__hint">{{ hint }}</span>
  </article>
</template>
