export interface ConflictRetryGuidance {
  currentQueueVersion: number;
  playbackEpoch?: number;
}

export interface ApiErrorBody {
  code: string;
  message: string;
  dependency?: string;
  retry?: ConflictRetryGuidance;
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

export function errorEnvelope(
  code: string,
  message: string,
  dependency?: string,
  retry?: ConflictRetryGuidance,
): ApiEnvelope<never> {
  return {
    success: false,
    error: {
      code,
      message,
      dependency,
      retry,
    },
  };
}

export type BackendService = "gateway" | "auth" | "catalog" | "jam" | "playback" | "bff" | "realtime";
