import { NextRequest } from "next/server";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { POST } from "@/app/api/auth/refresh/route";

const httpMocks = vi.hoisted(() => ({
  backendJson: vi.fn(),
}));

vi.mock("@/lib/api/http", () => ({
  backendJson: httpMocks.backendJson,
}));

function buildRequest(cookieHeader: string, csrfHeader = ""): NextRequest {
  const headers: Record<string, string> = {
    "Content-Type": "application/json",
    Cookie: cookieHeader,
  };
  if (csrfHeader) {
    headers["X-CSRF-Token"] = csrfHeader;
  }

  return new NextRequest("http://localhost:3000/api/auth/refresh", {
    method: "POST",
    headers,
    body: JSON.stringify({}),
  });
}

describe("POST /api/auth/refresh", () => {
  beforeEach(() => {
    httpMocks.backendJson.mockReset();
  });

  it("returns 403 when CSRF header is missing or mismatched", async () => {
    const response = await POST(buildRequest("refresh_token=token-1; csrf_token=token-csrf"));
    const payload = (await response.json()) as { success: boolean; error?: { code: string } };

    expect(response.status).toBe(403);
    expect(payload.success).toBe(false);
    expect(payload.error?.code).toBe("forbidden");
    expect(httpMocks.backendJson).not.toHaveBeenCalled();
  });

  it("rotates cookies and returns claims when refresh succeeds", async () => {
    httpMocks.backendJson.mockResolvedValue({
      ok: true,
      status: 200,
      data: {
        accessToken: "next-access-token",
        refreshToken: "next-refresh-token",
        tokenType: "Bearer",
        expiresAt: "2099-01-01T00:00:00Z",
        claims: {
          userId: "user-premium-1",
          plan: "premium",
          sessionState: "valid",
          scope: ["jam:read", "jam:control"],
        },
      },
    });

    const response = await POST(buildRequest("refresh_token=token-1; csrf_token=token-csrf", "token-csrf"));
    const payload = (await response.json()) as {
      success: boolean;
      data?: { claims: { userId: string } };
    };
    const setCookie = response.headers.get("set-cookie") ?? "";

    expect(response.status).toBe(200);
    expect(payload.success).toBe(true);
    expect(payload.data?.claims.userId).toBe("user-premium-1");
    expect(setCookie).toContain("auth_token=");
    expect(setCookie).toContain("refresh_token=");
    expect(setCookie).toContain("csrf_token=");
    expect(httpMocks.backendJson).toHaveBeenCalledWith(
      expect.objectContaining({
        service: "auth",
        path: "/v1/auth/refresh",
        method: "POST",
        body: { refreshToken: "token-1" },
      }),
    );
  });
});