import type { NextRequest } from "next/server";

import { resolveAuthHeaders } from "@/lib/api/auth";
import { backendJson } from "@/lib/api/http";
import { jsonError, jsonSuccess } from "@/lib/api/response";
import { claimsSchema } from "@/lib/auth/schemas";

export async function GET(request: NextRequest) {
  const auth = resolveAuthHeaders(request);
  if (!auth) {
    return jsonError("unauthorized", "missing authentication context", 401);
  }

  const result = await backendJson<unknown>({
    service: "auth",
    path: "/v1/auth/me",
    method: "GET",
    authHeader: auth.authHeader,
    cookieHeader: auth.cookieHeader,
  });
  if (!result.ok || !result.data) {
    return jsonError(result.error?.code ?? "unauthorized", result.error?.message ?? "unauthorized", result.status);
  }

  const claims = claimsSchema.safeParse(result.data);
  if (!claims.success) {
    return jsonError("dependency_invalid_response", "invalid auth response", 502);
  }

  return jsonSuccess(claims.data);
}