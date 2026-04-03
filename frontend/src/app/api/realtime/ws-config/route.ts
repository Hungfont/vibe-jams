import type { NextRequest } from "next/server";

import { getDefaultWsUrl } from "@/lib/api/config";
import { jsonError, jsonSuccess } from "@/lib/api/response";

export async function GET(request: NextRequest) {
  const sessionId = request.nextUrl.searchParams.get("sessionId")?.trim();
  if (!sessionId) {
    return jsonError("invalid_input", "sessionId is required", 400);
  }

  const lastSeenVersion = request.nextUrl.searchParams.get("lastSeenVersion")?.trim() ?? "";
  const wsBase = getDefaultWsUrl();

  return jsonSuccess({
    wsUrl: `${wsBase}/ws`,
    sessionId,
    lastSeenVersion,
  });
}
