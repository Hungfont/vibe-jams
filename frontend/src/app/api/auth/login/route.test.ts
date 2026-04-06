import { NextRequest } from "next/server";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { POST } from "@/app/api/auth/login/route";

const httpMocks = vi.hoisted(() => ({
  backendJson: vi.fn(),
}));

vi.mock("@/lib/api/http", () => ({
  backendJson: httpMocks.backendJson,
}));

function buildRequest(payload: unknown): NextRequest {
  return new NextRequest("http://localhost:3000/api/auth/login", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(payload),
  });
}

describe("POST /api/auth/login", () => {
  beforeEach(() => {
    httpMocks.backendJson.mockReset();
  });

  it("returns 400 for invalid login payload", async () => {
    const response = await POST(buildRequest({ identity: "", password: "short" }));
    const payload = (await response.json()) as { success: boolean; error?: { code: string } };

    expect(response.status).toBe(400);
    expect(payload.success).toBe(false);
    expect(payload.error?.code).toBe("invalid_input");
    expect(httpMocks.backendJson).not.toHaveBeenCalled();
  });

  it("maps upstream auth error into envelope", async () => {
    httpMocks.backendJson.mockResolvedValue({
      ok: false,
      status: 401,
      error: {
        code: "unauthorized",
        message: "invalid credentials",
      },
    });

    const response = await POST(buildRequest({ identity: "premium@example.com", password: "wrong-pass" }));
    const payload = (await response.json()) as { success: boolean; error?: { code: string } };

    expect(response.status).toBe(401);
    expect(payload.success).toBe(false);
    expect(payload.error?.code).toBe("unauthorized");
  });

  it("issues auth cookies and returns normalized claims on success", async () => {
    httpMocks.backendJson.mockResolvedValue({
      ok: true,
      status: 200,
      data: {
        accessToken: "access-token",
        refreshToken: "refresh-token",
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

    const response = await POST(buildRequest({ identity: "premium@example.com", password: "premium-pass" }));
    const payload = (await response.json()) as {
      success: boolean;
      data?: { claims: { userId: string; scope?: string[] } };
    };
    const setCookie = response.headers.get("set-cookie") ?? "";

    expect(response.status).toBe(200);
    expect(payload.success).toBe(true);
    expect(payload.data?.claims.userId).toBe("user-premium-1");
    expect(payload.data?.claims.scope).toEqual(["jam:read", "jam:control"]);
    expect(setCookie).toContain("auth_token=");
    expect(setCookie).toContain("refresh_token=");
    expect(setCookie).toContain("csrf_token=");
  });
});