import type { NextRequest } from "next/server";

import { resolveAuthHeaders } from "@/lib/api/auth";
import { backendJson } from "@/lib/api/http";
import { jsonError, jsonSuccess } from "@/lib/api/response";
import { permissionUpdateSchema } from "@/lib/jam/schemas";
import type { SessionPermissions } from "@/lib/jam/types";

export async function GET(
  request: NextRequest,
  context: { params: Promise<{ jamId: string }> },
) {
  const auth = resolveAuthHeaders(request);
  if (!auth) {
    return jsonError("unauthorized", "missing authentication context", 401);
  }

  const { jamId } = await context.params;
  const result = await backendJson<SessionPermissions>({
    service: "gateway",
    path: `/api/v1/jams/${encodeURIComponent(jamId)}/permissions`,
    method: "GET",
    authHeader: auth.authHeader,
    cookieHeader: auth.cookieHeader,
  });

  if (!result.ok || !result.data) {
    return jsonError(
      result.error?.code ?? "internal_error",
      result.error?.message ?? "failed to fetch permissions",
      result.status,
      result.error?.dependency,
      result.error?.retry,
    );
  }

  return jsonSuccess(result.data);
}

export async function POST(
  request: NextRequest,
  context: { params: Promise<{ jamId: string }> },
) {
  const auth = resolveAuthHeaders(request);
  if (!auth) {
    return jsonError("unauthorized", "missing authentication context", 401);
  }

  const body = (await request.json().catch(() => null)) as unknown;
  const parsed = permissionUpdateSchema.safeParse(body);
  if (!parsed.success) {
    return jsonError("invalid_input", "invalid permissions payload", 400);
  }
  if (
    parsed.data.canControlPlayback === undefined &&
    parsed.data.canReorderQueue === undefined &&
    parsed.data.canChangeVolume === undefined
  ) {
    return jsonError("invalid_input", "at least one permission field is required", 400);
  }

  const { jamId } = await context.params;
  const result = await backendJson<SessionPermissions>({
    service: "gateway",
    path: `/api/v1/jams/${encodeURIComponent(jamId)}/permissions`,
    method: "POST",
    body: parsed.data,
    authHeader: auth.authHeader,
    cookieHeader: auth.cookieHeader,
  });

  if (!result.ok || !result.data) {
    return jsonError(
      result.error?.code ?? "internal_error",
      result.error?.message ?? "failed to update permissions",
      result.status,
      result.error?.dependency,
      result.error?.retry,
    );
  }

  return jsonSuccess(result.data);
}
