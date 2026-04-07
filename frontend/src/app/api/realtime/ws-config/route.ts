import type { NextRequest } from "next/server";

import { resolveAuthHeaders } from "@/lib/api/auth";
import { backendJson } from "@/lib/api/http";
import { jsonError, jsonSuccess } from "@/lib/api/response";

export async function GET(request: NextRequest) {
  const sessionId = request.nextUrl.searchParams.get("sessionId")?.trim();
  if (!sessionId) {
    return jsonError("invalid_input", "sessionId is required", 400);
  }

  const lastSeenVersion = request.nextUrl.searchParams.get("lastSeenVersion")?.trim() ?? "";

  const query = new URLSearchParams({ sessionId });
  if (lastSeenVersion.length > 0) {
    query.set("lastSeenVersion", lastSeenVersion);
  }

  const auth = resolveAuthHeaders(request);

  const result = await backendJson<{ wsUrl: string; sessionId: string; lastSeenVersion: string }>({
    service: "gateway",
    path: `/v1/bff/mvp/realtime/ws-config?${query.toString()}`,
    method: "GET",
    authHeader: auth?.authHeader,
    cookieHeader: auth?.cookieHeader,
  });
  if (!result.ok || !result.data) {
    return jsonError(
      result.error?.code ?? "dependency_unavailable",
      result.error?.message ?? "failed to resolve realtime bootstrap config",
      result.status,
      result.error?.dependency,
    );
  }

  return jsonSuccess({
    wsUrl: result.data.wsUrl,
    sessionId: result.data.sessionId,
    lastSeenVersion: result.data.lastSeenVersion,
  });
}
