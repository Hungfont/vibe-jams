import type { NextRequest } from "next/server";

import { resolveAuthHeaders } from "@/lib/api/auth";
import { isPremiumPlan, validateClaims } from "@/lib/api/claims";
import { backendJson } from "@/lib/api/http";
import { jsonError, jsonSuccess } from "@/lib/api/response";
import type { SessionSnapshot } from "@/lib/jam/types";

export async function POST(request: NextRequest) {
  const auth = resolveAuthHeaders(request);
  if (!auth) {
    return jsonError("unauthorized", "missing authentication context", 401);
  }

  const claimResult = await validateClaims(auth.authHeader, auth.cookieHeader);
  if (!claimResult.ok || !claimResult.claims) {
    return jsonError(claimResult.code ?? "unauthorized", claimResult.message ?? "unauthorized", 401);
  }

  if (!isPremiumPlan(claimResult.claims.plan)) {
    return jsonError("premium_required", "premium plan required", 403);
  }

  const result = await backendJson<SessionSnapshot>({
    service: "jam",
    path: "/api/v1/jams/create",
    method: "POST",
    authHeader: auth.authHeader,
    cookieHeader: auth.cookieHeader,
  });

  if (!result.ok || !result.data) {
    return jsonError(result.error?.code ?? "internal_error", result.error?.message ?? "failed to create jam", result.status);
  }

  return jsonSuccess(result.data, 201);
}
