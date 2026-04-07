import { NextRequest } from "next/server";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { GET } from "@/app/api/realtime/ws-config/route";

const mocks = vi.hoisted(() => ({
  backendJson: vi.fn(),
}));

vi.mock("@/lib/api/http", () => ({
  backendJson: mocks.backendJson,
}));

describe("GET /api/realtime/ws-config", () => {
  beforeEach(() => {
    mocks.backendJson.mockReset();
  });

  it("returns 400 when sessionId is missing", async () => {
    const request = new NextRequest("http://localhost:3000/api/realtime/ws-config", {
      method: "GET",
    });

    const response = await GET(request);
    const payload = (await response.json()) as { success: boolean; error?: { code: string } };

    expect(response.status).toBe(400);
    expect(payload.success).toBe(false);
    expect(payload.error?.code).toBe("invalid_input");
  });

  it("routes realtime bootstrap config via gateway BFF endpoint", async () => {
    mocks.backendJson.mockResolvedValue({
      ok: true,
      status: 200,
      data: {
        wsUrl: "ws://localhost:8085/v1/bff/mvp/realtime/ws",
        sessionId: "jam_1",
        lastSeenVersion: "5",
      },
    });

    const request = new NextRequest(
      "http://localhost:3000/api/realtime/ws-config?sessionId=jam_1&lastSeenVersion=5",
      {
        method: "GET",
        headers: {
          cookie: "auth_token=cookie-token; refresh_token=refresh-token",
        },
      },
    );

    const response = await GET(request);
    const payload = (await response.json()) as {
      success: boolean;
      data?: { wsUrl: string; sessionId: string; lastSeenVersion: string };
    };

    expect(response.status).toBe(200);
    expect(payload.success).toBe(true);
    expect(payload.data?.wsUrl).toBe("ws://localhost:8085/v1/bff/mvp/realtime/ws");
    expect(mocks.backendJson).toHaveBeenCalledWith(
      expect.objectContaining({
        service: "gateway",
        authHeader: "Bearer cookie-token",
        cookieHeader: "auth_token=cookie-token; refresh_token=refresh-token",
      }),
    );
    const callArg = mocks.backendJson.mock.calls[0][0] as { path: string };
    expect(callArg.path).toContain("/v1/bff/mvp/realtime/ws-config?");
    expect(callArg.path).toContain("sessionId=jam_1");
    expect(callArg.path).toContain("lastSeenVersion=5");
  });
});
