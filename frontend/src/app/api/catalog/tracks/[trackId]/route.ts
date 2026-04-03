import type { NextRequest } from "next/server";

import { backendJson } from "@/lib/api/http";
import { jsonError, jsonSuccess } from "@/lib/api/response";
import type { TrackLookup } from "@/lib/jam/types";

export async function GET(
  _request: NextRequest,
  context: { params: Promise<{ trackId: string }> },
) {
  const { trackId } = await context.params;
  if (!trackId || trackId.trim().length === 0) {
    return jsonError("invalid_input", "trackId is required", 400);
  }

  const result = await backendJson<TrackLookup>({
    service: "catalog",
    path: `/internal/v1/catalog/tracks/${encodeURIComponent(trackId)}`,
    method: "GET",
  });

  if (!result.ok || !result.data) {
    return jsonError(
      result.error?.code ?? "track_not_found",
      result.error?.message ?? "track lookup failed",
      result.status,
    );
  }

  return jsonSuccess(result.data);
}
