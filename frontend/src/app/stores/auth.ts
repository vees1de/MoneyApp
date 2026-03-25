import { computed, ref } from "vue";
import { defineStore } from "pinia";

import type { AuthProvider } from "@/entities/user/model/types";
import { useDashboardStore } from "@/app/stores/dashboard";
import { useFinanceStore } from "@/app/stores/finance";
import { useReviewStore } from "@/app/stores/review";
import { useSavingsStore } from "@/app/stores/savings";
import { useUserStore } from "@/app/stores/user";
import { mapApiUser } from "@/shared/api/mappers";
import { clearSession, readSession, writeSession } from "@/shared/api/session";
import {
  fetchMe,
  loginWithEmail,
  loginWithTelegram,
  loginWithYandex,
  logout as apiLogout,
} from "@/shared/api/services/auth";
import { getYandexAuthPayload } from "@/shared/lib/auth-provider";
import { requestTelegramAuthPayload } from "@/shared/lib/telegram-oidc";

type AuthStatus = "unknown" | "guest" | "authenticated";

export const useAuthStore = defineStore("auth", () => {
  const status = ref<AuthStatus>("unknown");
  const provider = ref<AuthProvider | null>(null);
  const accessToken = ref<string | null>(null);
  const expiresAt = ref<string | null>(null);
  const bootstrapped = ref(false);

  function setSession(input: {
    accessToken: string;
    expiresAt: string;
    provider: AuthProvider | null;
    refreshToken: string;
  }) {
    status.value = "authenticated";
    provider.value = input.provider;
    accessToken.value = input.accessToken;
    expiresAt.value = input.expiresAt;

    writeSession({
      accessToken: input.accessToken,
      expiresAt: input.expiresAt,
      provider: input.provider,
      refreshToken: input.refreshToken,
    });
  }

  async function bootstrap() {
    const session = readSession();

    if (!session) {
      status.value = "guest";
      accessToken.value = null;
      expiresAt.value = null;
      provider.value = null;
      bootstrapped.value = true;
      return;
    }

    accessToken.value = session.accessToken;
    expiresAt.value = session.expiresAt;
    provider.value = session.provider;
    status.value = "authenticated";

    try {
      const response = await fetchMe();
      useUserStore().setProfile({
        ...mapApiUser(response.user),
        provider: session.provider,
      });
    } catch {
      clearSession();
      status.value = "guest";
      accessToken.value = null;
      expiresAt.value = null;
      provider.value = null;
    } finally {
      bootstrapped.value = true;
    }
  }

  async function login(nextProvider: Exclude<AuthProvider, "email">) {
    const response = await (async () => {
      if (nextProvider === "telegram") {
        return loginWithTelegram(await requestTelegramAuthPayload());
      }

      const payload = getYandexAuthPayload();
      if (!payload) {
        throw new Error(
          "Yandex auth payload is unavailable in the current runtime.",
        );
      }

      return loginWithYandex(payload);
    })();

    const nextExpiresAt = new Date(
      Date.now() + response.tokens.expires_in * 1000,
    ).toISOString();

    setSession({
      accessToken: response.tokens.access_token,
      expiresAt: nextExpiresAt,
      provider: nextProvider,
      refreshToken: response.tokens.refresh_token,
    });

    useUserStore().setProfile({
      ...mapApiUser(response.user),
      provider: nextProvider,
    });
  }

  async function loginWithEmailCredentials(email: string, password: string) {
    const response = await loginWithEmail(email, password);
    const nextExpiresAt = new Date(
      Date.now() + response.tokens.expires_in * 1000,
    ).toISOString();

    setSession({
      accessToken: response.tokens.access_token,
      expiresAt: nextExpiresAt,
      provider: "email",
      refreshToken: response.tokens.refresh_token,
    });

    useUserStore().setProfile({
      ...mapApiUser(response.user),
      provider: "email",
    });
  }

  async function logout() {
    const session = readSession();

    if (session?.refreshToken) {
      try {
        await apiLogout(session.refreshToken);
      } catch {
        // Ignore logout failures and clear local session regardless.
      }
    }

    status.value = "guest";
    provider.value = null;
    accessToken.value = null;
    expiresAt.value = null;
    clearSession();
    useFinanceStore().reset();
    useSavingsStore().reset();
    useReviewStore().reset();
    useDashboardStore().reset();
    useUserStore().reset();
  }

  const isAuthenticated = computed(() => status.value === "authenticated");

  return {
    accessToken,
    bootstrapped,
    expiresAt,
    isAuthenticated,
    login,
    loginWithEmailCredentials,
    logout,
    provider,
    status,
    bootstrap,
  };
});
