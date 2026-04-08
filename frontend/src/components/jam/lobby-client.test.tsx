import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { LobbyClient } from "@/components/jam/lobby-client";

const routerMocks = vi.hoisted(() => ({
  push: vi.fn(),
}));

vi.mock("next/navigation", () => ({
  useRouter: () => ({
    push: routerMocks.push,
  }),
}));

function jsonResponse(payload: unknown): Response {
  return new Response(JSON.stringify(payload), {
    status: 200,
    headers: { "Content-Type": "application/json" },
  });
}

describe("LobbyClient", () => {
  beforeEach(() => {
    routerMocks.push.mockReset();
    vi.restoreAllMocks();
  });

  it("navigates to jam room after successful create flow", async () => {
    const fetchMock = vi.fn().mockResolvedValueOnce(
      jsonResponse({
        success: true,
        data: {
          jamId: "jam-create-1",
          status: "active",
          hostUserId: "host-1",
          participants: [{ userId: "host-1", role: "host" }],
          sessionVersion: 1,
        },
      }),
    );

    vi.stubGlobal("fetch", fetchMock);

    render(<LobbyClient />);
    const createSubmit = screen
      .getAllByRole("button", { name: "Create Jam" })
      .find((button) => button.className.includes("w-full"));
    expect(createSubmit).toBeDefined();
    await userEvent.click(createSubmit as HTMLButtonElement);

    await waitFor(() => {
      expect(routerMocks.push).toHaveBeenCalledWith("/jam/jam-create-1");
    });
    expect(fetchMock).toHaveBeenCalledTimes(1);
  });

  it("navigates to jam room after successful join flow", async () => {
    const fetchMock = vi.fn().mockResolvedValueOnce(
      jsonResponse({
        success: true,
        data: {
          jamId: "jam-join-1",
          status: "active",
          hostUserId: "host-1",
          participants: [
            { userId: "host-1", role: "host" },
            { userId: "member-1", role: "member" },
          ],
          sessionVersion: 2,
        },
      }),
    );

    vi.stubGlobal("fetch", fetchMock);

    render(<LobbyClient />);

    await userEvent.click(screen.getAllByRole("button", { name: "Join Jam" })[0]);
    await userEvent.type(screen.getByPlaceholderText("Enter jamId"), "jam-join-1");
    const joinSubmit = screen
      .getAllByRole("button", { name: "Join Jam" })
      .find((button) => button.className.includes("w-full"));
    expect(joinSubmit).toBeDefined();
    await userEvent.click(joinSubmit as HTMLButtonElement);

    await waitFor(() => {
      expect(routerMocks.push).toHaveBeenCalledWith("/jam/jam-join-1");
    });
    expect(fetchMock).toHaveBeenCalledTimes(1);
  });
});
