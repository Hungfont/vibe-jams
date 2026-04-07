import type { NextRequest } from "next/server";

import { resolveAuthHeaders } from "@/lib/api/auth";
import { validateClaims } from "@/lib/api/claims";
import { backendJson } from "@/lib/api/http";
import { jsonError, jsonSuccess } from "@/lib/api/response";
import { joinJamSchema } from "@/lib/jam/schemas";
import type { SessionSnapshot } from "@/lib/jam/types";

export async function POST(
  request: NextRequest,
  context: { params: Promise<{ jamId: string }> },
) {
  const auth = resolveAuthHeaders(request);
  if (!auth) {
    return jsonError("unauthorized", "missing authentication context", 401);
  }

  const claimResult = await validateClaims(auth.authHeader, auth.cookieHeader);
  if (!claimResult.ok || !claimResult.claims) {
    return jsonError(claimResult.code ?? "unauthorized", claimResult.message ?? "unauthorized", 401);
  }

  const { jamId } = await context.params;
  const valid = joinJamSchema.safeParse({ jamId });
  if (!valid.success) {
    return jsonError("invalid_input", "jamId is invalid", 400);
  }

  const result = await backendJson<SessionSnapshot>({
    service: "gateway",
    path: `/api/v1/jams/${encodeURIComponent(jamId)}/join`,
    method: "POST",
    authHeader: auth.authHeader,
    cookieHeader: auth.cookieHeader,
  });

  if (!result.ok || !result.data) {
    return jsonError(result.error?.code ?? "internal_error", result.error?.message ?? "failed to join jam", result.status);
  }

  return jsonSuccess(result.data);
}
