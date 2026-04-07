import type { NextRequest } from "next/server";

import { backendJson } from "@/lib/api/http";
import { jsonError, jsonSuccess } from "@/lib/api/response";
import { authTokenPairSchema, loginRequestSchema } from "@/lib/auth/schemas";
import { issueCsrfToken, setAuthCookies } from "@/lib/auth/server-cookies";

export async function POST(request: NextRequest) {
  const body = (await request.json().catch(() => null)) as unknown;
  const parsed = loginRequestSchema.safeParse(body);
  if (!parsed.success) {
    return jsonError("invalid_input", "invalid login payload", 400);
  }

  const result = await backendJson<unknown>({
    service: "gateway",
    path: "/v1/auth/login",
    method: "POST",
    body: parsed.data,
  });
  if (!result.ok || !result.data) {
    return jsonError(result.error?.code ?? "unauthorized", result.error?.message ?? "invalid credentials", result.status);
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