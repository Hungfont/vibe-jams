import { NextRequest } from "next/server";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { GET, POST } from "@/app/api/jam/[jamId]/permissions/route";

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
}));

vi.mock("@/lib/api/http", () => ({
  backendJson: mocks.backendJson,
}));

function buildGetRequest(): NextRequest {
  return new NextRequest("http://localhost:3000/api/jam/jam_1/permissions", {
    method: "GET",
    headers: {
      Authorization: "Bearer token",
    },
  });
}

function buildPostRequest(body: unknown): NextRequest {
  return new NextRequest("http://localhost:3000/api/jam/jam_1/permissions", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      Authorization: "Bearer token",
    },
    body: JSON.stringify(body),
  });
}

describe("/api/jam/[jamId]/permissions", () => {
  beforeEach(() => {
    mocks.resolveAuthHeaders.mockReset();
    mocks.validateClaims.mockReset();
    mocks.backendJson.mockReset();
  });

  it("returns 401 when auth context is missing", async () => {
    mocks.resolveAuthHeaders.mockReturnValue(null);

    const response = await GET(buildGetRequest(), {
      params: Promise.resolve({ jamId: "jam_1" }),
    });
    const payload = (await response.json()) as { success: boolean; error?: { code: string } };

    expect(response.status).toBe(401);
    expect(payload.success).toBe(false);
    expect(payload.error?.code).toBe("unauthorized");
    expect(mocks.backendJson).not.toHaveBeenCalled();
  });

  it("forwards GET permission lookup via gateway route", async () => {
    mocks.resolveAuthHeaders.mockReturnValue({
      authHeader: "Bearer token",
      cookieHeader: "auth_token=token",
    });
    mocks.validateClaims.mockResolvedValue({
      ok: true,
      claims: {
        userId: "host-1",
        plan: "premium",
        sessionState: "valid",
      },
    });
    mocks.backendJson.mockResolvedValue({
      ok: true,
      status: 200,
      data: {
        canControlPlayback: true,
        canReorderQueue: false,
        canChangeVolume: true,
      },
    });

    const response = await GET(buildGetRequest(), {
      params: Promise.resolve({ jamId: "jam_1" }),
    });
    const payload = (await response.json()) as {
      success: boolean;
      data?: { canControlPlayback: boolean; canReorderQueue: boolean; canChangeVolume: boolean };
    };

    expect(response.status).toBe(200);
    expect(payload.success).toBe(true);
    expect(payload.data?.canControlPlayback).toBe(true);
    expect(mocks.backendJson).toHaveBeenCalledWith(
      expect.objectContaining({
        service: "gateway",
        method: "GET",
        path: "/api/v1/jams/jam_1/permissions",
      }),
    );
  });

  it("returns 400 for invalid POST permission payload", async () => {
    mocks.resolveAuthHeaders.mockReturnValue({
      authHeader: "Bearer token",
      cookieHeader: "auth_token=token",
    });
    mocks.validateClaims.mockResolvedValue({
      ok: true,
      claims: {
        userId: "host-1",
        plan: "premium",
        sessionState: "valid",
      },
    });

    const response = await POST(buildPostRequest({ invalid: true }), {
      params: Promise.resolve({ jamId: "jam_1" }),
    });
    const payload = (await response.json()) as { success: boolean; error?: { code: string } };

    expect(response.status).toBe(400);
    expect(payload.success).toBe(false);
    expect(payload.error?.code).toBe("invalid_input");
    expect(mocks.backendJson).not.toHaveBeenCalled();
  });

  it("forwards POST permission update via gateway route", async () => {
    mocks.resolveAuthHeaders.mockReturnValue({
      authHeader: "Bearer token",
      cookieHeader: "auth_token=token",
    });
    mocks.validateClaims.mockResolvedValue({
      ok: true,
      claims: {
        userId: "host-1",
        plan: "premium",
        sessionState: "valid",
      },
    });
    mocks.backendJson.mockResolvedValue({
      ok: true,
      status: 200,
      data: {
        canControlPlayback: true,
        canReorderQueue: true,
        canChangeVolume: false,
      },
    });

    const response = await POST(
      buildPostRequest({
        canControlPlayback: true,
        canReorderQueue: true,
      }),
      {
        params: Promise.resolve({ jamId: "jam_1" }),
      },
    );

    const payload = (await response.json()) as {
      success: boolean;
      data?: { canControlPlayback: boolean; canReorderQueue: boolean; canChangeVolume: boolean };
    };

    expect(response.status).toBe(200);
    expect(payload.success).toBe(true);
    expect(payload.data?.canReorderQueue).toBe(true);
    expect(mocks.backendJson).toHaveBeenCalledWith(
      expect.objectContaining({
        service: "gateway",
        method: "POST",
        path: "/api/v1/jams/jam_1/permissions",
        body: {
          canControlPlayback: true,
          canReorderQueue: true,
        },
      }),
    );
  });
});
