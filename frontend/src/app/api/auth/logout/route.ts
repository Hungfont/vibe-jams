import type { NextRequest } from "next/server";

import { backendJson } from "@/lib/api/http";
import { jsonError, jsonSuccess } from "@/lib/api/response";
import { logoutRequestSchema } from "@/lib/auth/schemas";
import { clearAuthCookies, isValidCsrfRequest, readRefreshToken } from "@/lib/auth/server-cookies";

export async function POST(request: NextRequest) {
  if (!isValidCsrfRequest(request)) {
    return jsonError("forbidden", "invalid csrf token", 403);
  }

  const body = (await request.json().catch(() => ({}))) as unknown;
  const parsed = logoutRequestSchema.safeParse(body);
  if (!parsed.success) {
    return jsonError("invalid_input", "invalid logout payload", 400);
  }

  const refreshToken = parsed.data.refreshToken?.trim() || readRefreshToken(request);
  if (!refreshToken) {
    const response = jsonError("unauthorized", "missing refresh token", 401);
    clearAuthCookies(response);
    return response;
  }

  const result = await backendJson<unknown>({
    service: "gateway",
    path: "/v1/auth/logout",
    method: "POST",
    body: { refreshToken },
  });
  if (!result.ok) {
    const response = jsonError(result.error?.code ?? "unauthorized", result.error?.message ?? "unable to logout", result.status);
    clearAuthCookies(response);
    return response;
  }

  const response = jsonSuccess({ status: "ok" });
  clearAuthCookies(response);
  return response;
}