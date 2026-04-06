import type { NextRequest } from "next/server";

import { backendJson } from "@/lib/api/http";
import { jsonError, jsonSuccess } from "@/lib/api/response";
import { authTokenPairSchema, refreshRequestSchema } from "@/lib/auth/schemas";
import { isValidCsrfRequest, issueCsrfToken, readRefreshToken, setAuthCookies } from "@/lib/auth/server-cookies";

export async function POST(request: NextRequest) {
  if (!isValidCsrfRequest(request)) {
    return jsonError("forbidden", "invalid csrf token", 403);
  }

  const body = (await request.json().catch(() => ({}))) as unknown;
  const parsed = refreshRequestSchema.safeParse(body);
  if (!parsed.success) {
    return jsonError("invalid_input", "invalid refresh payload", 400);
  }

  const refreshToken = parsed.data.refreshToken?.trim() || readRefreshToken(request);
  if (!refreshToken) {
    return jsonError("unauthorized", "missing refresh token", 401);
  }

  const result = await backendJson<unknown>({
    service: "auth",
    path: "/v1/auth/refresh",
    method: "POST",
    body: { refreshToken },
  });
  if (!result.ok || !result.data) {
    return jsonError(result.error?.code ?? "unauthorized", result.error?.message ?? "invalid refresh token", result.status);
  }

  const tokenPair = authTokenPairSchema.safeParse(result.data);
  if (!tokenPair.success) {
    return jsonError("dependency_invalid_response", "invalid auth response", 502);
  }

  const csrfToken = issueCsrfToken();
  const response = jsonSuccess({
    claims: tokenPair.data.claims,
    expiresAt: tokenPair.data.expiresAt,
  });
  setAuthCookies(response, tokenPair.data.accessToken, tokenPair.data.refreshToken, tokenPair.data.expiresAt, csrfToken);

  return response;
}