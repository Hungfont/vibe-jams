import { NextRequest } from "next/server";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { POST } from "@/app/api/jam/create/route";

const mocks = vi.hoisted(() => ({
  resolveAuthHeaders: vi.fn(),
  validateClaims: vi.fn(),
  backendJson: vi.fn(),
}));

vi.mock("@/lib/api/auth", () => ({
  resolveAuthHeaders: mocks.resolveAuthHeaders,
}));

vi.mock("@/lib/api/claims", () => ({
  validateClaims: mocks.validateClaims,
  isPremiumPlan: () => true,
}));

vi.mock("@/lib/api/http", () => ({
  backendJson: mocks.backendJson,
}));

function buildRequest(): NextRequest {
  return new NextRequest("http://localhost:3000/api/jam/create", {
    method: "POST",
  });
}

describe("POST /api/jam/create", () => {
  beforeEach(() => {
    mocks.resolveAuthHeaders.mockReset();
    mocks.validateClaims.mockReset();
    mocks.backendJson.mockReset();
  });

  it("returns 401 when auth context is missing", async () => {
    mocks.resolveAuthHeaders.mockReturnValue(null);

    const response = await POST(buildRequest());
    const payload = (await response.json()) as { success: boolean; error?: { code: string } };

    expect(response.status).toBe(401);
    expect(payload.success).toBe(false);
    expect(payload.error?.code).toBe("unauthorized");
    expect(mocks.backendJson).not.toHaveBeenCalled();
  });

  it("routes create flow via gateway service", async () => {
    mocks.resolveAuthHeaders.mockReturnValue({
      authHeader: "Bearer token",
      cookieHeader: "auth_token=token",
    });
    mocks.validateClaims.mockResolvedValue({
      ok: true,
      claims: {
        userId: "user-1",
        plan: "premium",
        sessionState: "valid",
        scope: ["jam:control"],
      },
    });
    mocks.backendJson.mockResolvedValue({
      ok: true,
      status: 201,
      data: { jamId: "jam_1" },
    });

    const response = await POST(buildRequest());
    const payload = (await response.json()) as { success: boolean; data?: { jamId: string } };

    expect(response.status).toBe(201);
    expect(payload.success).toBe(true);
    expect(payload.data?.jamId).toBe("jam_1");
    expect(mocks.backendJson).toHaveBeenCalledWith(
      expect.objectContaining({
        service: "gateway",
        path: "/api/v1/jams/create",
        method: "POST",
      }),
    );
  });
});
