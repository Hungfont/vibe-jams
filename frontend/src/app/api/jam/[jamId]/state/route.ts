import type { NextRequest } from "next/server";

import { backendJson } from "@/lib/api/http";
import { jsonError, jsonSuccess } from "@/lib/api/response";
import type { SessionStateSnapshot } from "@/lib/jam/types";

export async function GET(
  request: NextRequest,
  context: { params: Promise<{ jamId: string }> },
) {
  const { jamId } = await context.params;
  const authHeader = request.headers.get("authorization") ?? undefined;
  const cookieHeader = request.headers.get("cookie") ?? undefined;

  const result = await backendJson<SessionStateSnapshot>({
    service: "gateway",
    path: `/api/v1/jams/${encodeURIComponent(jamId)}/state`,
    method: "GET",
    authHeader,
    cookieHeader,
  });

  if (!result.ok || !result.data) {
    return jsonError(result.error?.code ?? "internal_error", result.error?.message ?? "failed to load jam state", result.status);
  }

  return jsonSuccess(result.data);
}
