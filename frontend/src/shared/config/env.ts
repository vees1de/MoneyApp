const fallbackApiBaseUrl = "/api/v1";

function trimEnv(value: string | undefined) {
  const nextValue = value?.trim();
  return nextValue ? nextValue : undefined;
}

export const env = {
  appName: "Personal Life OS",
  productName: "Finance Core",
  apiBaseUrl: import.meta.env.VITE_API_BASE_URL ?? fallbackApiBaseUrl,
  telegramBotUsername: trimEnv(import.meta.env.VITE_TELEGRAM_BOT_USERNAME),
  yandexClientId: trimEnv(import.meta.env.VITE_YANDEX_CLIENT_ID),
  yandexRedirectUri: trimEnv(import.meta.env.VITE_YANDEX_REDIRECT_URI),
};
