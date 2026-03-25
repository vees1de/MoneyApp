const fallbackApiBaseUrl = "/api/v1";

function trimEnv(value: string | undefined) {
  const nextValue = value?.trim();
  return nextValue ? nextValue : undefined;
}

export const env = {
  appName: "Personal Life OS",
  productName: "Finance Core",
  apiBaseUrl: import.meta.env.VITE_API_BASE_URL ?? fallbackApiBaseUrl,
  telegramClientId: trimEnv(import.meta.env.VITE_TELEGRAM_CLIENT_ID),
  yandexClientId: trimEnv(import.meta.env.VITE_YANDEX_CLIENT_ID),
  yandexRedirectUri: trimEnv(import.meta.env.VITE_YANDEX_REDIRECT_URI),
};
