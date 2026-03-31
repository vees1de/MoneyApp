export type UUID = string;

export type QueryValue = string | number | boolean;

export interface ListQuery {
  limit?: number;
  offset?: number;
  [key: string]: QueryValue | QueryValue[] | undefined;
}

export interface ListResponse<T> {
  items: T[];
}

export interface ApiErrorPayload {
  code?: string;
  message?: string;
  details?: Record<string, unknown>;
}
