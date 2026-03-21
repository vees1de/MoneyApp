<script setup lang="ts">
defineProps<{
  description: string
  title: string
}>()
</script>

<template>
  <div class="empty-state surface-card">
    <h3>{{ title }}</h3>
    <p class="muted">{{ description }}</p>
  </div>
</template>
