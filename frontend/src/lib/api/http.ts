import { getBackendBaseUrl } from "@/lib/api/config";
import { normalizeUpstreamError } from "@/lib/api/errors";
import type { ApiErrorBody, BackendService } from "@/lib/api/types";

interface BackendJsonResult<T> {
  ok: boolean;
  status: number;
  data?: T;
  error?: ApiErrorBody;
}

interface BackendJsonOptions {
  service: BackendService;
  path: string;
  method?: "GET" | "POST" | "PATCH" | "PUT" | "DELETE";
  body?: unknown;
  timeoutMs?: number;
  authHeader?: string;
  cookieHeader?: string;
}

async function parseJson(response: Response): Promise<unknown> {
  const text = await response.text();
  if (!text) {
    return {};
  }

  try {
    return JSON.parse(text) as unknown;
  } catch {
    return {};
  }
}

export async function backendJson<T>(options: BackendJsonOptions): Promise<BackendJsonResult<T>> {
  const timeoutMs = options.timeoutMs ?? 5000;
  const controller = new AbortController();
  const timeout = setTimeout(() => controller.abort(), timeoutMs);

  try {
    const headers = new Headers();
    headers.set("Content-Type", "application/json");
    if (options.authHeader) {
      headers.set("Authorization", options.authHeader);
    }
    if (options.cookieHeader) {
      headers.set("Cookie", options.cookieHeader);
    }

    const response = await fetch(`${getBackendBaseUrl(options.service)}${options.path}`, {
      method: options.method ?? "GET",
      headers,
      body: options.body === undefined ? undefined : JSON.stringify(options.body),
      cache: "no-store",
      signal: controller.signal,
    });

    const payload = await parseJson(response);
    if (!response.ok) {
      return {
        ok: false,
        status: response.status,
        error: normalizeUpstreamError(response.status, payload, "upstream request failed"),
      };
    }

    return {
      ok: true,
      status: response.status,
      data: payload as T,
    };
  } catch {
    return {
      ok: false,
      status: 503,
      error: {
        code: "dependency_timeout",
        message: "upstream request timed out or unavailable",
      },
    };
  } finally {
    clearTimeout(timeout);
  }
}
