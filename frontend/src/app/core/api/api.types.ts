export type UUID = string;

export interface ListQuery {
  limit?: number;
  offset?: number;
  [key: string]: string | number | boolean | undefined;
}

export interface ApiErrorPayload {
  code?: string;
  message?: string;
  details?: Record<string, unknown>;
}