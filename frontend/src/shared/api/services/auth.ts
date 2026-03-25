import type { ApiAuthResponse, ApiMeResponse } from "@/shared/api/contracts";
import { apiRequest } from "@/shared/api/http/client";

export interface TelegramAuthPayload {
  provider_user_id: string;
  username?: string;
  first_name?: string;
  last_name?: string;
  photo_url?: string;
  auth_date?: number;
  hash?: string;
}

export interface YandexAuthPayload {
  code?: string;
  id_token?: string;
  provider_user_id?: string;
  email?: string;
  display_name?: string;
  avatar_url?: string;
}

export function loginWithTelegram(payload: TelegramAuthPayload) {
  return apiRequest<ApiAuthResponse>("/auth/telegram", {
    method: "POST",
    auth: false,
    body: JSON.stringify(payload),
  });
}

export function loginWithYandex(payload: YandexAuthPayload) {
  return apiRequest<ApiAuthResponse>("/auth/yandex", {
    method: "POST",
    auth: false,
    body: JSON.stringify(payload),
  });
}

export function loginWithEmail(email: string, _password: string) {
  return apiRequest<ApiAuthResponse>("/auth/email", {
    method: "POST",
    auth: false,
    body: JSON.stringify({
      email,
      display_name: email.split("@")[0],
    }),
  });
}

export function fetchMe() {
  return apiRequest<ApiMeResponse>("/auth/me");
}

export function logout(refreshToken: string) {
  return apiRequest<void>("/auth/logout", {
    method: "POST",
    auth: false,
    retryOnUnauthorized: false,
    body: JSON.stringify({
      refresh_token: refreshToken,
    }),
  });
}
