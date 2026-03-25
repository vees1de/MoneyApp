import type { TelegramAuthPayload } from "@/shared/api/services/auth";

import { env } from "@/shared/config/env";
import { getLocale, translate } from "@/shared/i18n";

const TELEGRAM_LOGIN_SCRIPT_URL =
  "https://oauth.telegram.org/js/telegram-login.js?3";

interface TelegramLoginInitOptions {
  client_id: number;
  lang?: string;
  nonce?: string;
}

interface TelegramLoginCallbackPayload {
  error?: string;
  id_token?: string;
}

interface TelegramLoginApi {
  auth(
    options: TelegramLoginInitOptions,
    callback: (payload: TelegramLoginCallbackPayload) => void,
  ): void;
}

declare global {
  interface Window {
    Telegram?: {
      Login?: TelegramLoginApi;
    };
  }
}

let telegramLoginPromise: Promise<TelegramLoginApi> | null = null;

function trimToUndefined(value: string | undefined | null) {
  const nextValue = value?.trim();
  return nextValue ? nextValue : undefined;
}

function getTelegramLoginApi() {
  const loginApi = window.Telegram?.Login;
  if (!loginApi) {
    throw new Error(translate("auth.telegramScriptFailed"));
  }

  return loginApi;
}

function loadTelegramLoginSdk() {
  if (typeof window === "undefined" || typeof document === "undefined") {
    return Promise.reject(new Error(translate("auth.telegramScriptFailed")));
  }

  if (telegramLoginPromise) {
    return telegramLoginPromise;
  }

  telegramLoginPromise = new Promise<TelegramLoginApi>((resolve, reject) => {
    try {
      const existingScript = document.querySelector<HTMLScriptElement>(
        `script[src="${TELEGRAM_LOGIN_SCRIPT_URL}"]`,
      );

      if (existingScript) {
        existingScript.addEventListener(
          "load",
          () => resolve(getTelegramLoginApi()),
          { once: true },
        );
        existingScript.addEventListener(
          "error",
          () => reject(new Error(translate("auth.telegramScriptFailed"))),
          { once: true },
        );

        const loginApi = window.Telegram?.Login;
        if (loginApi) {
          resolve(loginApi);
        }
        return;
      }

      const script = document.createElement("script");
      script.src = TELEGRAM_LOGIN_SCRIPT_URL;
      script.async = true;
      script.onload = () => {
        try {
          resolve(getTelegramLoginApi());
        } catch (error) {
          reject(error);
        }
      };
      script.onerror = () =>
        reject(new Error(translate("auth.telegramScriptFailed")));
      document.head.append(script);
    } catch (error) {
      reject(error);
    }
  }).catch((error) => {
    telegramLoginPromise = null;
    throw error;
  });

  return telegramLoginPromise;
}

export function preloadTelegramLoginSdk() {
  void loadTelegramLoginSdk().catch(() => {
    // The actual login attempt will surface the error to the user.
  });
}

function createNonce() {
  const bytes = new Uint8Array(16);
  crypto.getRandomValues(bytes);
  return Array.from(bytes, (value) => value.toString(16).padStart(2, "0")).join(
    "",
  );
}

export async function requestTelegramAuthPayload(): Promise<TelegramAuthPayload> {
  const clientId = trimToUndefined(env.telegramClientId);
  if (!clientId) {
    throw new Error(translate("auth.telegramUnavailable"));
  }

  const numericClientId = Number(clientId);
  if (!Number.isSafeInteger(numericClientId) || numericClientId <= 0) {
    throw new Error(translate("auth.telegramUnavailable"));
  }

  const loginApi = await loadTelegramLoginSdk();
  const nonce = createNonce();

  return new Promise<TelegramAuthPayload>((resolve, reject) => {
    loginApi.auth(
      {
        client_id: numericClientId,
        lang: getLocale(),
        nonce,
      },
      (payload) => {
        const errorMessage = trimToUndefined(payload.error);
        if (errorMessage) {
          reject(new Error(errorMessage));
          return;
        }

        const idToken = trimToUndefined(payload.id_token);
        if (!idToken) {
          reject(new Error(translate("auth.telegramUnexpectedResponse")));
          return;
        }

        resolve({
          id_token: idToken,
          nonce,
        });
      },
    );
  });
}
