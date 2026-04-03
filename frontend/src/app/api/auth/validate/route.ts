import type { NextRequest } from "next/server";

import { resolveAuthHeaders } from "@/lib/api/auth";
import { validateClaims } from "@/lib/api/claims";
import { jsonError, jsonSuccess } from "@/lib/api/response";

export async function POST(request: NextRequest) {
  const auth = resolveAuthHeaders(request);
  if (!auth) {
    return jsonError("unauthorized", "missing authentication context", 401);
  }

  const claims = await validateClaims(auth.authHeader, auth.cookieHeader);
  if (!claims.ok || !claims.claims) {
    return jsonError(claims.code ?? "unauthorized", claims.message ?? "unauthorized", 401);
  }

  return jsonSuccess(claims.claims);
}
