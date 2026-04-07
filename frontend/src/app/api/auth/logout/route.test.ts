import { NextRequest } from "next/server";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { POST } from "@/app/api/auth/logout/route";

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

  return new NextRequest("http://localhost:3000/api/auth/logout", {
    method: "POST",
    headers,
    body: JSON.stringify({}),
  });
}

describe("POST /api/auth/logout", () => {
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

  it("forwards logout through api-gateway and clears cookies", async () => {
    httpMocks.backendJson.mockResolvedValue({
      ok: true,
      status: 200,
      data: {},
    });

    const response = await POST(buildRequest("refresh_token=token-1; csrf_token=token-csrf", "token-csrf"));
    const payload = (await response.json()) as { success: boolean; data?: { status: string } };
    const setCookie = response.headers.get("set-cookie") ?? "";

    expect(response.status).toBe(200);
    expect(payload.success).toBe(true);
    expect(payload.data?.status).toBe("ok");
    expect(setCookie).toContain("auth_token=");
    expect(setCookie).toContain("refresh_token=");
    expect(setCookie).toContain("csrf_token=");
    expect(httpMocks.backendJson).toHaveBeenCalledWith(
      expect.objectContaining({
        service: "gateway",
        path: "/v1/auth/logout",
        method: "POST",
        body: { refreshToken: "token-1" },
      }),
    );
  });
});
