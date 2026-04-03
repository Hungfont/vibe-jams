import type { NextRequest } from "next/server";

import { resolveAuthHeaders } from "@/lib/api/auth";
import { backendJson } from "@/lib/api/http";
import { jsonError, jsonSuccess } from "@/lib/api/response";
import type { BffOrchestrationData } from "@/lib/jam/types";

export async function POST(
  request: NextRequest,
  context: { params: Promise<{ jamId: string }> },
) {
  const auth = resolveAuthHeaders(request);
  if (!auth) {
    return jsonError("unauthorized", "missing authentication context", 401);
  }

  const body = (await request.json().catch(() => ({}))) as unknown;
  const { jamId } = await context.params;
  const result = await backendJson<BffOrchestrationData>({
    service: "bff",
    path: `/v1/bff/mvp/sessions/${encodeURIComponent(jamId)}/orchestration`,
    method: "POST",
    body,
    authHeader: auth.authHeader,
    cookieHeader: auth.cookieHeader,
  });

  if (!result.ok || !result.data) {
    return jsonError(
      result.error?.code ?? "internal_error",
      result.error?.message ?? "failed to orchestrate room",
      result.status,
      result.error?.dependency,
    );
  }

  return jsonSuccess(result.data);
}
