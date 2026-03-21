<script setup lang="ts">
defineProps<{
  description: string
  title: string
  icon?: string
}>()
</script>

<template>
  <div class="empty-state section-card">
    <div class="empty-state__icon">{{ icon ?? '✦' }}</div>
    <p class="empty-state__title">{{ title }}</p>
    <p class="empty-state__desc">{{ description }}</p>
  </div>
</template>
