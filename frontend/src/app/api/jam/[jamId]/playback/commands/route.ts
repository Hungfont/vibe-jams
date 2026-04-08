import type { NextRequest } from "next/server";

import { resolveAuthHeaders } from "@/lib/api/auth";
import { backendJson } from "@/lib/api/http";
import { jsonError, jsonSuccess } from "@/lib/api/response";
import { playbackCommandSchema } from "@/lib/jam/schemas";
import type { PlaybackAccepted } from "@/lib/jam/types";

export async function POST(
  request: NextRequest,
  context: { params: Promise<{ jamId: string }> },
) {
  const auth = resolveAuthHeaders(request);
  if (!auth) {
    return jsonError("unauthorized", "missing authentication context", 401);
  }

  const body = (await request.json().catch(() => null)) as unknown;
  const parsed = playbackCommandSchema.safeParse(body);
  if (!parsed.success) {
    return jsonError("invalid_input", "invalid playback payload", 400);
  }

  const { jamId } = await context.params;
  const result = await backendJson<PlaybackAccepted>({
    service: "gateway",
    path: `/v1/jam/sessions/${encodeURIComponent(jamId)}/playback/commands`,
    method: "POST",
    body: parsed.data,
    authHeader: auth.authHeader,
    cookieHeader: auth.cookieHeader,
  });

  if (!result.ok || !result.data) {
    return jsonError(
      result.error?.code ?? "internal_error",
      result.error?.message ?? "failed to execute playback command",
      result.status,
      result.error?.dependency,
      result.error?.retry,
    );
  }

  return jsonSuccess(result.data, 202);
}
