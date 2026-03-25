<script setup lang="ts">
import { onMounted, ref, watch } from "vue";

import { env } from "@/shared/config/env";

const props = defineProps<{
  disabled?: boolean;
}>();

const container = ref<HTMLDivElement | null>(null);

function buildAuthUrl() {
  const url = new URL(window.location.href);
  url.search = "";
  url.hash = "";
  return url.toString();
}

function mountWidget() {
  if (!container.value || !env.telegramBotUsername) {
    return;
  }

  container.value.replaceChildren();

  const script = document.createElement("script");
  script.async = true;
  script.src = "https://telegram.org/js/telegram-widget.js?23";
  script.setAttribute("data-telegram-login", env.telegramBotUsername);
  script.setAttribute("data-size", "large");
  script.setAttribute("data-radius", "26");
  script.setAttribute("data-auth-url", buildAuthUrl());
  script.setAttribute("data-request-access", "write");

  container.value.append(script);
}

onMounted(() => {
  mountWidget();
});

watch(
  () => env.telegramBotUsername,
  () => {
    mountWidget();
  },
);
</script>

<template>
  <div
    class="telegram-widget"
    :class="{ 'telegram-widget--disabled': disabled }"
    :aria-disabled="disabled ? 'true' : 'false'"
  >
    <div ref="container" class="telegram-widget__mount" />
    <div v-if="disabled" class="telegram-widget__overlay" />
  </div>
</template>

<style scoped>
.telegram-widget {
  position: relative;
  display: flex;
  justify-content: center;
  width: 100%;
}

.telegram-widget__mount {
  display: flex;
  justify-content: center;
  width: 100%;
}

.telegram-widget__overlay {
  position: absolute;
  inset: 0;
  cursor: not-allowed;
}

.telegram-widget--disabled {
  opacity: 0.45;
}
</style>
