import type { NextRequest } from "next/server";

import { resolveAuthHeaders } from "@/lib/api/auth";
import { backendJson } from "@/lib/api/http";
import { jsonError, jsonSuccess } from "@/lib/api/response";
import type { BffOrchestrationData } from "@/lib/jam/types";
import { bffOrchestrationDataSchema, bffUpstreamEnvelopeSchema } from "@/lib/jam/schemas";

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
    service: "gateway",
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

  const upstreamEnvelope = bffUpstreamEnvelopeSchema.safeParse(result.data);
  if (upstreamEnvelope.success) {
    if (!upstreamEnvelope.data.success) {
      return jsonError(
        upstreamEnvelope.data.error?.code ?? "internal_error",
        upstreamEnvelope.data.error?.message ?? "failed to orchestrate room",
        result.status >= 400 ? result.status : 502,
        upstreamEnvelope.data.error?.dependency,
      );
    }
    return jsonSuccess(upstreamEnvelope.data.data);
  }

  const directPayload = bffOrchestrationDataSchema.safeParse(result.data);
  if (directPayload.success) {
    return jsonSuccess(directPayload.data);
  }

  return jsonError("dependency_invalid_response", "invalid orchestration payload", 502, "bff");
}
