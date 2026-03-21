<script setup lang="ts">
const items = [
  { label: 'Home', to: '/dashboard' },
  { label: 'Accounts', to: '/accounts' },
  { label: 'Money', to: '/transactions' },
  { label: 'Goals', to: '/savings' },
  { label: 'Review', to: '/review' },
]
</script>

<template>
  <nav aria-label="Primary navigation" class="bottom-nav">
    <RouterLink
      v-for="item in items"
      :key="item.to"
      :to="item.to"
      class="bottom-nav__link"
    >
      {{ item.label }}
    </RouterLink>
  </nav>
</template>
