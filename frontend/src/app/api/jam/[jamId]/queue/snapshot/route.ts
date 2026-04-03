import type { NextRequest } from "next/server";

import { backendJson } from "@/lib/api/http";
import { jsonError, jsonSuccess } from "@/lib/api/response";
import type { QueueSnapshot } from "@/lib/jam/types";

export async function GET(
  request: NextRequest,
  context: { params: Promise<{ jamId: string }> },
) {
  const { jamId } = await context.params;
  const result = await backendJson<QueueSnapshot>({
    service: "jam",
    path: `/api/v1/jams/${encodeURIComponent(jamId)}/queue/snapshot`,
    method: "GET",
    authHeader: request.headers.get("authorization") ?? undefined,
    cookieHeader: request.headers.get("cookie") ?? undefined,
  });

  if (!result.ok || !result.data) {
    return jsonError(result.error?.code ?? "internal_error", result.error?.message ?? "failed to load queue snapshot", result.status);
  }

  return jsonSuccess(result.data);
}
