<script setup lang="ts">
import AppHeader from '@/shared/ui/AppHeader.vue'
import BottomTabBar from '@/shared/ui/BottomTabBar.vue'
import OfflineBanner from '@/shared/ui/OfflineBanner.vue'
import ToastStack from '@/shared/ui/ToastStack.vue'
</script>

<template>
  <div class="app-shell">
    <div class="app-shell__inner app-layout">
      <AppHeader />
      <OfflineBanner />

      <main class="app-layout__content">
        <RouterView />
      </main>

      <BottomTabBar />
      <ToastStack />
    </div>
  </div>
</template>
