import { NextRequest } from "next/server";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { GET } from "@/app/api/catalog/tracks/[trackId]/route";

const mocks = vi.hoisted(() => ({
  backendJson: vi.fn(),
}));

vi.mock("@/lib/api/http", () => ({
  backendJson: mocks.backendJson,
}));

function buildRequest(): NextRequest {
  return new NextRequest("http://localhost:3000/api/catalog/tracks/trk_1", {
    method: "GET",
  });
}

describe("GET /api/catalog/tracks/[trackId]", () => {
  beforeEach(() => {
    mocks.backendJson.mockReset();
  });

  it("returns 400 when trackId is missing", async () => {
    const response = await GET(buildRequest(), {
      params: Promise.resolve({ trackId: "" }),
    });
    const payload = (await response.json()) as { success: boolean; error?: { code: string } };

    expect(response.status).toBe(400);
    expect(payload.success).toBe(false);
    expect(payload.error?.code).toBe("invalid_input");
  });

  it("routes catalog lookup via gateway service", async () => {
    mocks.backendJson.mockResolvedValue({
      ok: true,
      status: 200,
      data: { trackId: "trk_1", title: "Track" },
    });

    const response = await GET(buildRequest(), {
      params: Promise.resolve({ trackId: "trk_1" }),
    });
    const payload = (await response.json()) as { success: boolean; data?: { trackId: string } };

    expect(response.status).toBe(200);
    expect(payload.success).toBe(true);
    expect(payload.data?.trackId).toBe("trk_1");
    expect(mocks.backendJson).toHaveBeenCalledWith(
      expect.objectContaining({
        service: "gateway",
        path: "/internal/v1/catalog/tracks/trk_1",
        method: "GET",
      }),
    );
  });
});
