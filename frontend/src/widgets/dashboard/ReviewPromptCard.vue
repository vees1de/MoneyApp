<script setup lang="ts">
defineProps<{
  actionLabel: string
  body: string
  tone: 'success' | 'warning'
}>()
</script>

<template>
  <section class="surface-card stack">
    <div class="split">
      <div>
        <h3>Weekly review</h3>
        <p class="muted">{{ body }}</p>
      </div>
      <span class="pill" :class="tone === 'success' ? 'pill--success' : 'pill--warning'">
        {{ tone === 'success' ? 'Healthy' : 'Attention needed' }}
      </span>
    </div>

    <RouterLink class="button button--primary" to="/review">
      {{ actionLabel }}
    </RouterLink>
  </section>
</template>
