<script setup lang="ts">
import AppHeader from '@/shared/ui/AppHeader.vue'
import BottomTabBar from '@/shared/ui/BottomTabBar.vue'
import OfflineBanner from '@/shared/ui/OfflineBanner.vue'
import QuickActionFab from '@/shared/ui/QuickActionFab.vue'
import SidebarNav from '@/shared/ui/SidebarNav.vue'
import ToastStack from '@/shared/ui/ToastStack.vue'
</script>

<template>
  <div class="app-shell">
    <!-- Desktop sidebar — shown via CSS at ≥960px -->
    <div class="desktop-sidebar-wrap">
      <SidebarNav />
    </div>

    <div class="app-shell__inner app-layout">
      <!-- Mobile-only header -->
      <div class="mobile-chrome">
        <AppHeader />
      </div>
      <OfflineBanner />

      <main class="app-layout__content">
        <RouterView />
      </main>

      <!-- Mobile-only bottom nav -->
      <div class="mobile-chrome">
        <BottomTabBar />
      </div>
      <QuickActionFab />
      <ToastStack />
    </div>
  </div>
</template>
