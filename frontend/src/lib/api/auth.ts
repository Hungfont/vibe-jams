import type { NextRequest } from "next/server";

function parseCookieToken(cookieHeader: string | null): string | null {
  if (!cookieHeader) {
    return null;
  }

  const parts = cookieHeader.split(";").map((part) => part.trim());
  for (const part of parts) {
    if (part.startsWith("auth_token=")) {
      return decodeURIComponent(part.slice("auth_token=".length));
    }
    if (part.startsWith("token=")) {
      return decodeURIComponent(part.slice("token=".length));
    }
  }

  return null;
}

export function resolveAuthHeaders(request: NextRequest): { authHeader: string; cookieHeader: string } | null {
  const authHeader = request.headers.get("authorization")?.trim() ?? "";
  const cookieHeader = request.headers.get("cookie") ?? "";

  if (authHeader) {
    return { authHeader, cookieHeader };
  }

  const token = parseCookieToken(cookieHeader);
  if (!token) {
    return null;
  }

  return {
    authHeader: `Bearer ${token}`,
    cookieHeader,
  };
}
