import type { NextRequest } from "next/server";

import { resolveAuthHeaders } from "@/lib/api/auth";
import { validateClaims } from "@/lib/api/claims";
import { backendJson } from "@/lib/api/http";
import { jsonError, jsonSuccess } from "@/lib/api/response";
import { queueReorderSchema } from "@/lib/jam/schemas";
import type { QueueSnapshot } from "@/lib/jam/types";

export async function POST(
  request: NextRequest,
  context: { params: Promise<{ jamId: string }> },
) {
  const auth = resolveAuthHeaders(request);
  if (!auth) {
    return jsonError("unauthorized", "missing authentication context", 401);
  }

  const claims = await validateClaims(auth.authHeader, auth.cookieHeader);
  if (!claims.ok || !claims.claims) {
    return jsonError(claims.code ?? "unauthorized", claims.message ?? "unauthorized", 401);
  }

  const body = (await request.json().catch(() => null)) as unknown;
  const parsed = queueReorderSchema.safeParse(body);
  if (!parsed.success) {
    return jsonError("invalid_input", "invalid reorder payload", 400);
  }

  const { jamId } = await context.params;
  const result = await backendJson<QueueSnapshot>({
    service: "jam",
    path: `/api/v1/jams/${encodeURIComponent(jamId)}/queue/reorder`,
    method: "POST",
    body: parsed.data,
    authHeader: auth.authHeader,
    cookieHeader: auth.cookieHeader,
  });

  if (!result.ok || !result.data) {
    return jsonError(result.error?.code ?? "internal_error", result.error?.message ?? "failed to reorder queue", result.status);
  }

  return jsonSuccess(result.data);
}
