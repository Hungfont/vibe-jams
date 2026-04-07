import { NextRequest } from "next/server";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { POST } from "@/app/api/bff/jam/[jamId]/orchestration/route";

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
  return new NextRequest("http://localhost:3000/api/bff/jam/jam_1/orchestration", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      Authorization: "Bearer token",
    },
    body: JSON.stringify({}),
  });
}

function buildOrchestrationData() {
  return {
    claims: {
      userId: "host-1",
      plan: "premium",
      sessionState: "valid",
      scope: ["jam:read"],
    },
    sessionState: {
      session: {
        jamId: "jam_1",
        status: "active",
        hostUserId: "host-1",
        participants: [{ userId: "host-1", role: "host" }],
        sessionVersion: 1,
      },
      queue: {
        jamId: "jam_1",
        queueVersion: 1,
        items: [],
      },
      aggregateVersion: 1,
    },
    partial: false,
    dependencyStatuses: {
      auth: "ok",
      jam: "ok",
    },
  };
}

describe("POST /api/bff/jam/[jamId]/orchestration", () => {
  beforeEach(() => {
    mocks.resolveAuthHeaders.mockReset();
    mocks.backendJson.mockReset();
  });

  it("returns 401 when auth context is missing", async () => {
    mocks.resolveAuthHeaders.mockReturnValue(null);

    const response = await POST(buildRequest(), {
      params: Promise.resolve({ jamId: "jam_1" }),
    });
    const payload = (await response.json()) as { success: boolean; error?: { code: string } };

    expect(response.status).toBe(401);
    expect(payload.success).toBe(false);
    expect(payload.error?.code).toBe("unauthorized");
    expect(mocks.backendJson).not.toHaveBeenCalled();
  });

  it("unwraps upstream bff envelope into frontend envelope data", async () => {
    mocks.resolveAuthHeaders.mockReturnValue({
      authHeader: "Bearer token",
      cookieHeader: "auth_token=token",
    });
    mocks.backendJson.mockResolvedValue({
      ok: true,
      status: 200,
      data: {
        success: true,
        data: buildOrchestrationData(),
      },
    });

    const response = await POST(buildRequest(), {
      params: Promise.resolve({ jamId: "jam_1" }),
    });
    const payload = (await response.json()) as {
      success: boolean;
      data?: { claims?: { userId: string }; sessionState?: { session?: { jamId: string } }; success?: boolean };
    };

    expect(response.status).toBe(200);
    expect(payload.success).toBe(true);
    expect(payload.data?.claims?.userId).toBe("host-1");
    expect(payload.data?.sessionState?.session?.jamId).toBe("jam_1");
    expect(payload.data?.success).toBeUndefined();
    expect(mocks.backendJson).toHaveBeenCalledWith(
      expect.objectContaining({
        service: "gateway",
        path: "/v1/bff/mvp/sessions/jam_1/orchestration",
        method: "POST",
      }),
    );
  });

  it("returns 502 when upstream payload shape is invalid", async () => {
    mocks.resolveAuthHeaders.mockReturnValue({
      authHeader: "Bearer token",
      cookieHeader: "auth_token=token",
    });
    mocks.backendJson.mockResolvedValue({
      ok: true,
      status: 200,
      data: {
        success: true,
        data: {
          claims: { userId: "host-1" },
        },
      },
    });

    const response = await POST(buildRequest(), {
      params: Promise.resolve({ jamId: "jam_1" }),
    });
    const payload = (await response.json()) as { success: boolean; error?: { code: string; dependency?: string } };

    expect(response.status).toBe(502);
    expect(payload.success).toBe(false);
    expect(payload.error?.code).toBe("dependency_invalid_response");
    expect(payload.error?.dependency).toBe("bff");
  });
});
