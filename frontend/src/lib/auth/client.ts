import type { ApiEnvelope } from "@/lib/api/types";
import { loginRequestSchema, type ClaimsPayload, type LoginRequest } from "@/lib/auth/schemas";

export interface AuthSessionPayload {
  claims: ClaimsPayload;
  expiresAt: string;
}

interface ErrorField {
  field: string;
  message: string;
}

function invalidInputEnvelope(message: string, fieldErrors?: ErrorField[]): ApiEnvelope<never> {
  return {
    success: false,
    error: {
      code: "invalid_input",
      message,
      fieldErrors,
    },
  };
}

async function parseEnvelope<T>(response: Response): Promise<ApiEnvelope<T>> {
  const payload = (await response.json().catch(() => null)) as ApiEnvelope<T> | null;
  if (!payload || typeof payload.success !== "boolean") {
    return {
      success: false,
      error: {
        code: "invalid_response",
        message: "invalid response from auth route",
      },
    };
  }
  return payload;
}

function getCookieValue(name: string): string {
  if (typeof document === "undefined") {
    return "";
  }

  const needle = `${name}=`;
  const parts = document.cookie.split(";").map((part) => part.trim());
  for (const part of parts) {
    if (part.startsWith(needle)) {
      return decodeURIComponent(part.slice(needle.length));
    }
  }

  return "";
}

function csrfHeader(): Record<string, string> {
  const csrfToken = getCookieValue("csrf_token");
  if (!csrfToken) {
    return {};
  }
  return { "X-CSRF-Token": csrfToken };
}

export async function loginWithPassword(input: LoginRequest): Promise<ApiEnvelope<AuthSessionPayload>> {
  const parsed = loginRequestSchema.safeParse(input);
  if (!parsed.success) {
    const flattened = parsed.error.flatten().fieldErrors;
    const fieldErrors: ErrorField[] = [];
    if (flattened.identity?.[0]) {
      fieldErrors.push({ field: "identity", message: flattened.identity[0] });
    }
    if (flattened.password?.[0]) {
      fieldErrors.push({ field: "password", message: flattened.password[0] });
    }
    return invalidInputEnvelope("Invalid login input", fieldErrors);
  }

  const response = await fetch("/api/auth/login", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(parsed.data),
  });

  return parseEnvelope<AuthSessionPayload>(response);
}

export async function refreshSession(): Promise<ApiEnvelope<AuthSessionPayload>> {
  const response = await fetch("/api/auth/refresh", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      ...csrfHeader(),
    },
    body: JSON.stringify({}),
  });

  return parseEnvelope<AuthSessionPayload>(response);
}

export async function logoutSession(): Promise<ApiEnvelope<{ status: string }>> {
  const response = await fetch("/api/auth/logout", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      ...csrfHeader(),
    },
    body: JSON.stringify({}),
  });

  return parseEnvelope<{ status: string }>(response);
}

export async function fetchCurrentUser(): Promise<ApiEnvelope<ClaimsPayload>> {
  const response = await fetch("/api/auth/me", {
    method: "GET",
    cache: "no-store",
  });

  return parseEnvelope<ClaimsPayload>(response);
}