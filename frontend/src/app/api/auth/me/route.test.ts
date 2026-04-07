import { NextRequest } from "next/server";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { GET } from "@/app/api/auth/me/route";

const mocks = vi.hoisted(() => ({
  resolveAuthHeaders: vi.fn(),
  backendJson: vi.fn(),
}));

vi.mock("@/lib/api/auth", () => ({
  resolveAuthHeaders: mocks.resolveAuthHeaders,
}));

vi.mock("@/lib/api/http", () => ({
  backendJson: mocks.backendJson,
}));

function buildRequest(): NextRequest {
  return new NextRequest("http://localhost:3000/api/auth/me", {
    method: "GET",
    headers: {
      Authorization: "Bearer token",
      Cookie: "auth_token=token",
    },
  });
}

describe("GET /api/auth/me", () => {
  beforeEach(() => {
    mocks.resolveAuthHeaders.mockReset();
    mocks.backendJson.mockReset();
  });

  it("returns 401 when auth context is missing", async () => {
    mocks.resolveAuthHeaders.mockReturnValue(null);

    const response = await GET(buildRequest());
    const payload = (await response.json()) as { success: boolean; error?: { code: string } };

    expect(response.status).toBe(401);
    expect(payload.success).toBe(false);
    expect(payload.error?.code).toBe("unauthorized");
    expect(mocks.backendJson).not.toHaveBeenCalled();
  });

  it("forwards me lookup through api-gateway and returns claims", async () => {
    mocks.resolveAuthHeaders.mockReturnValue({
      authHeader: "Bearer token",
      cookieHeader: "auth_token=token",
    });
    mocks.backendJson.mockResolvedValue({
      ok: true,
      status: 200,
      data: {
        userId: "user-premium-1",
        plan: "premium",
        sessionState: "valid",
        scope: ["jam:read", "jam:control"],
      },
    });

    const response = await GET(buildRequest());
    const payload = (await response.json()) as {
      success: boolean;
      data?: { userId: string; scope?: string[] };
    };

    expect(response.status).toBe(200);
    expect(payload.success).toBe(true);
    expect(payload.data?.userId).toBe("user-premium-1");
    expect(payload.data?.scope).toEqual(["jam:read", "jam:control"]);
    expect(mocks.backendJson).toHaveBeenCalledWith(
      expect.objectContaining({
        service: "gateway",
        path: "/v1/auth/me",
        method: "GET",
        authHeader: "Bearer token",
        cookieHeader: "auth_token=token",
      }),
    );
  });
});
