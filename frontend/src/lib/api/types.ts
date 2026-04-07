export interface ApiErrorBody {
  code: string;
  message: string;
  dependency?: string;
  fieldErrors?: Array<{ field: string; message: string }>;
}

export interface ApiEnvelope<T> {
  success: boolean;
  data?: T;
  error?: ApiErrorBody;
  meta?: {
    page?: number;
    limit?: number;
    total?: number;
  };
}

export function successEnvelope<T>(data: T): ApiEnvelope<T> {
  return { success: true, data };
}

export function errorEnvelope(code: string, message: string, dependency?: string): ApiEnvelope<never> {
  return {
    success: false,
    error: {
      code,
      message,
      dependency,
    },
  };
}

export type BackendService = "gateway" | "auth" | "catalog" | "jam" | "playback" | "bff" | "realtime";
