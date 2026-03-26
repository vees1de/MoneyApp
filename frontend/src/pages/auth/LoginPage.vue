<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { useRouter } from "vue-router";

import type { AuthProvider } from "@/entities/user/model/types";
import { useAppUiStore } from "@/app/stores/app-ui";
import { useAuthStore } from "@/app/stores/auth";
import { useUserStore } from "@/app/stores/user";
import AuthProviderButtons from "@/features/auth/AuthProviderButtons.vue";
import { env } from "@/shared/config/env";
import { translateProvider, useI18n } from "@/shared/i18n";
import {
  buildYandexAuthorizeUrl,
  clearTelegramAuthPayloadFromUrl,
  clearYandexAuthPayloadFromUrl,
  getAvailableAuthProviders,
  hasTelegramAuthPayload,
  hasYandexAuthPayload,
  validateYandexState,
} from "@/shared/lib/auth-provider";
import LanguageSwitch from "@/shared/ui/LanguageSwitch.vue";

const authStore = useAuthStore();
const userStore = useUserStore();
const appUiStore = useAppUiStore();
const router = useRouter();
const { t } = useI18n();

const loading = ref(false);
const availableProviders = computed(() => getAvailableAuthProviders());

function getErrorMessage(error: unknown) {
  if (error instanceof Error) {
    return error.message;
  }

  if (
    typeof error === "object" &&
    error !== null &&
    "message" in error &&
    typeof error.message === "string"
  ) {
    return error.message;
  }

  return t("auth.signInFailed");
}

async function login(provider: Exclude<AuthProvider, "email">) {
  if (provider === "yandex" && !hasYandexAuthPayload()) {
    const authorizeUrl = buildYandexAuthorizeUrl();
    if (!authorizeUrl) {
      throw new Error(t("auth.yandexUnavailable"));
    }

    window.location.assign(authorizeUrl);
    return;
  }

  try {
    loading.value = true;
    await authStore.login(provider);
    if (provider === "telegram") {
      clearTelegramAuthPayloadFromUrl();
    }
    if (provider === "yandex") {
      clearYandexAuthPayloadFromUrl();
    }
    appUiStore.pushToast(
      t("auth.signedInWith", { provider: translateProvider(provider) }),
      "success",
    );
    await router.push(
      userStore.profile.onboardingCompleted ? "/dashboard" : "/onboarding",
    );
  } catch (error) {
    if (provider === "telegram" && hasTelegramAuthPayload()) {
      clearTelegramAuthPayloadFromUrl();
    }
    appUiStore.pushToast(getErrorMessage(error), "warning");
  } finally {
    loading.value = false;
  }
}

onMounted(async () => {
  if (hasTelegramAuthPayload() && !loading.value) {
    await login("telegram");
    return;
  }

  if (!hasYandexAuthPayload() || loading.value) {
    return;
  }

  if (!validateYandexState()) {
    appUiStore.pushToast(t("auth.yandexStateInvalid"), "warning");
    clearYandexAuthPayloadFromUrl();
    return;
  }

  await login("yandex");
});
</script>

<template>
  <div class="login-wrapper">
    <div class="login-card">
      <div class="login-toolbar">
        <LanguageSwitch />
      </div>

      <div class="login-brand">
        <div class="login-brand__icon">
          <svg viewBox="0 0 28 28" fill="none" width="32" height="32">
            <rect width="28" height="28" rx="8" fill="var(--brand)" />
            <path
              d="M8 20V10.5L14 7l6 3.5V20"
              stroke="#fff"
              stroke-width="1.8"
              stroke-linecap="round"
              stroke-linejoin="round"
            />
            <path
              d="M11 20v-5h6v5"
              stroke="#fff"
              stroke-width="1.8"
              stroke-linecap="round"
              stroke-linejoin="round"
            />
          </svg>
        </div>
        <div class="login-brand__label">
          <span class="login-brand__name">{{ env.appName }}</span>
          <span class="login-brand__sub">{{ env.productName }}</span>
        </div>
      </div>

      <div class="login-headline">
        <h1 class="login-headline__title">{{ t("auth.loginTitle") }}</h1>
        <p class="login-headline__sub">{{ t("auth.loginSubtitle") }}</p>
      </div>

      <div class="login-features">
        <div class="login-feature">
          <span class="login-feature__icon">📊</span>
          <span class="login-feature__text">{{ t("auth.featureBalance") }}</span>
        </div>
        <div class="login-feature">
          <span class="login-feature__icon">🔁</span>
          <span class="login-feature__text">{{ t("auth.featureReview") }}</span>
        </div>
        <div class="login-feature">
          <span class="login-feature__icon">🎯</span>
          <span class="login-feature__text">{{ t("auth.featureSavings") }}</span>
        </div>
      </div>

      <AuthProviderButtons
        :loading="loading"
        :available-providers="availableProviders"
        @select="login"
      />

      <p class="login-legal">
        {{ t("auth.legal") }}
      </p>
    </div>
  </div>
</template>

<style scoped>
.login-toolbar {
  display: flex;
  justify-content: flex-end;
}

.login-wrapper {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 100vh;
  min-height: 100dvh;
  padding: 24px 0;
}

.login-card {
  width: 100%;
  max-width: 420px;
  background: var(--surface);
  border-radius: var(--radius-2xl);
  border: 1px solid var(--border);
  box-shadow: var(--shadow-lg);
  padding: 36px 32px;
  display: flex;
  flex-direction: column;
  gap: 24px;
  animation: loginCardIn 0.5s var(--ease-out) both;
}

@keyframes loginCardIn {
  from {
    opacity: 0;
    transform: translateY(16px) scale(0.98);
  }
  to {
    opacity: 1;
    transform: translateY(0) scale(1);
  }
}

.login-brand {
  display: flex;
  align-items: center;
  gap: 12px;
}

.login-brand__icon {
  flex-shrink: 0;
  border-radius: 10px;
  overflow: hidden;
  display: flex;
}

.login-brand__label {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.login-brand__name {
  font-size: 1rem;
  font-weight: 700;
  color: var(--text-primary);
  letter-spacing: -0.02em;
}

.login-brand__sub {
  font-size: 0.8125rem;
  color: var(--text-muted);
}

.login-headline__title {
  margin: 0 0 10px;
  font-size: 2rem;
  font-weight: 700;
  letter-spacing: -0.03em;
  line-height: 1.08;
  color: var(--text-primary);
}

.login-headline__sub {
  margin: 0;
  font-size: 1rem;
  color: var(--text-secondary);
  line-height: 1.5;
}

.login-features {
  display: flex;
  gap: 10px;
}

.login-feature {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  padding: 16px 8px;
  background: var(--surface-secondary);
  border-radius: var(--radius-md);
  text-align: center;
  transition: background-color var(--duration-base) var(--ease-out);
}

.login-feature:hover {
  background: var(--surface-fill);
}

.login-feature__icon {
  font-size: 1.375rem;
}

.login-feature__text {
  font-size: 0.75rem;
  font-weight: 600;
  color: var(--text-secondary);
  line-height: 1.25;
}

.login-legal {
  margin: 0;
  font-size: 0.75rem;
  color: var(--text-muted);
  text-align: center;
  line-height: 1.45;
}
</style>
