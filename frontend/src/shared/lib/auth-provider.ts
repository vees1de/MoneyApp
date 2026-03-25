import type {
  TelegramAuthPayload,
  YandexAuthPayload,
} from "@/shared/api/services/auth";

import { env } from "@/shared/config/env";

const YANDEX_AUTHORIZE_URL = "https://oauth.yandex.ru/authorize";
const YANDEX_STATE_STORAGE_KEY = "moneyapp.yandex_oauth_state";
const TELEGRAM_PARAM_KEYS = [
  "id",
  "first_name",
  "last_name",
  "username",
  "photo_url",
  "auth_date",
  "hash",
] as const;
const YANDEX_PARAM_KEYS = [
  "code",
  "id_token",
  "provider_user_id",
  "email",
  "display_name",
  "avatar_url",
  "username",
  "error",
  "error_description",
  "state",
] as const;

function trimToUndefined(value: string | null) {
  const nextValue = value?.trim();
  return nextValue ? nextValue : undefined;
}

function readSearchAndHashParams() {
  if (typeof window === "undefined") {
    return new URLSearchParams();
  }

  const params = new URLSearchParams(window.location.search);
  const hash = window.location.hash.startsWith("#")
    ? window.location.hash.slice(1)
    : window.location.hash;
  const hashParams = new URLSearchParams(hash);

  for (const [key, value] of hashParams.entries()) {
    if (!params.has(key)) {
      params.set(key, value);
    }
  }

  return params;
}

function parseNumber(value: string | undefined) {
  if (!value) {
    return undefined;
  }

  const parsed = Number(value);
  return Number.isFinite(parsed) ? parsed : undefined;
}

export function getTelegramAuthPayload(): TelegramAuthPayload | null {
  const params = readSearchAndHashParams();
  const providerUserId = trimToUndefined(params.get("id"));
  const hash = trimToUndefined(params.get("hash"));
  const authDate = parseNumber(trimToUndefined(params.get("auth_date")));

  if (!providerUserId || !hash || authDate == null) {
    return null;
  }

  return {
    provider_user_id: providerUserId,
    username: trimToUndefined(params.get("username")),
    first_name: trimToUndefined(params.get("first_name")),
    last_name: trimToUndefined(params.get("last_name")),
    photo_url: trimToUndefined(params.get("photo_url")),
    auth_date: authDate,
    hash,
  };
}

export function hasTelegramAuthPayload() {
  return getTelegramAuthPayload() !== null;
}

export function isTelegramAuthConfigured() {
  return Boolean(env.telegramBotUsername);
}

export function getYandexAuthPayload(): YandexAuthPayload | null {
  const params = readSearchAndHashParams();
  const payload: YandexAuthPayload = {
    code: trimToUndefined(params.get("code")),
    id_token: trimToUndefined(params.get("id_token")),
    provider_user_id: trimToUndefined(params.get("provider_user_id")),
    email: trimToUndefined(params.get("email")),
    display_name:
      trimToUndefined(params.get("display_name")) ??
      trimToUndefined(params.get("username")),
    avatar_url: trimToUndefined(params.get("avatar_url")),
  };

  if (
    !payload.code &&
    !payload.id_token &&
    !payload.provider_user_id &&
    !payload.email &&
    !payload.display_name &&
    !payload.avatar_url
  ) {
    return null;
  }

  return payload;
}

export function hasYandexAuthPayload() {
  return getYandexAuthPayload() !== null;
}

export function isYandexAuthConfigured() {
  return Boolean(env.yandexClientId && env.yandexRedirectUri);
}

function canUseSessionStorage() {
  return (
    typeof window !== "undefined" &&
    typeof window.sessionStorage !== "undefined"
  );
}

function createYandexState() {
  if (
    typeof crypto !== "undefined" &&
    typeof crypto.randomUUID === "function"
  ) {
    return crypto.randomUUID();
  }

  return `${Date.now()}-${Math.random().toString(36).slice(2)}`;
}

export function buildYandexAuthorizeUrl() {
  if (!env.yandexClientId || !env.yandexRedirectUri) {
    return null;
  }

  const state = createYandexState();
  if (canUseSessionStorage()) {
    window.sessionStorage.setItem(YANDEX_STATE_STORAGE_KEY, state);
  }

  const url = new URL(YANDEX_AUTHORIZE_URL);
  url.searchParams.set("response_type", "code");
  url.searchParams.set("client_id", env.yandexClientId);
  url.searchParams.set("redirect_uri", env.yandexRedirectUri);
  url.searchParams.set("state", state);

  return url.toString();
}

export function validateYandexState() {
  const params = readSearchAndHashParams();
  const returnedState = trimToUndefined(params.get("state"));
  if (!returnedState) {
    return true;
  }

  if (!canUseSessionStorage()) {
    return true;
  }

  const expectedState = trimToUndefined(
    window.sessionStorage.getItem(YANDEX_STATE_STORAGE_KEY),
  );
  window.sessionStorage.removeItem(YANDEX_STATE_STORAGE_KEY);

  return Boolean(expectedState && expectedState === returnedState);
}

export function clearYandexAuthPayloadFromUrl() {
  if (typeof window === "undefined") {
    return;
  }

  const url = new URL(window.location.href);

  for (const key of YANDEX_PARAM_KEYS) {
    url.searchParams.delete(key);
  }

  const hashParams = new URLSearchParams(
    url.hash.startsWith("#") ? url.hash.slice(1) : url.hash,
  );
  let hashChanged = false;

  for (const key of YANDEX_PARAM_KEYS) {
    if (hashParams.has(key)) {
      hashParams.delete(key);
      hashChanged = true;
    }
  }

  const nextHash = hashParams.toString();
  url.hash =
    hashChanged && nextHash ? `#${nextHash}` : hashChanged ? "" : url.hash;

  window.history.replaceState(
    {},
    "",
    `${url.pathname}${url.search}${url.hash}`,
  );
}

export function clearTelegramAuthPayloadFromUrl() {
  if (typeof window === "undefined") {
    return;
  }

  const url = new URL(window.location.href);

  for (const key of TELEGRAM_PARAM_KEYS) {
    url.searchParams.delete(key);
  }

  const hashParams = new URLSearchParams(
    url.hash.startsWith("#") ? url.hash.slice(1) : url.hash,
  );
  let hashChanged = false;

  for (const key of TELEGRAM_PARAM_KEYS) {
    if (hashParams.has(key)) {
      hashParams.delete(key);
      hashChanged = true;
    }
  }

  const nextHash = hashParams.toString();
  url.hash =
    hashChanged && nextHash ? `#${nextHash}` : hashChanged ? "" : url.hash;

  window.history.replaceState(
    {},
    "",
    `${url.pathname}${url.search}${url.hash}`,
  );
}

export function getAvailableAuthProviders() {
  return {
    telegram: isTelegramAuthConfigured() || hasTelegramAuthPayload(),
    yandex: isYandexAuthConfigured() || hasYandexAuthPayload(),
  };
}
