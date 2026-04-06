import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { JamRoomClient } from "@/components/jam/jam-room-client";
import type { BffOrchestrationData, QueueSnapshot, SessionStateSnapshot } from "@/lib/jam/types";

const mocks = vi.hoisted(() => ({
  addQueueItem: vi.fn(),
  kickParticipant: vi.fn(),
  muteParticipant: vi.fn(),
  removeQueueItem: vi.fn(),
  reorderQueue: vi.fn(),
  endJamSession: vi.fn(),
  executePlayback: vi.fn(),
  fetchJamState: vi.fn(),
  fetchOrchestration: vi.fn(),
  fetchRealtimeWsConfig: vi.fn(),
}));

vi.mock("@/lib/jam/actions", () => ({
  addQueueItem: mocks.addQueueItem,
  kickParticipant: mocks.kickParticipant,
  muteParticipant: mocks.muteParticipant,
  removeQueueItem: mocks.removeQueueItem,
  reorderQueue: mocks.reorderQueue,
}));

vi.mock("@/lib/jam/client", () => ({
  endJamSession: mocks.endJamSession,
  executePlayback: mocks.executePlayback,
  fetchJamState: mocks.fetchJamState,
  fetchOrchestration: mocks.fetchOrchestration,
  fetchRealtimeWsConfig: mocks.fetchRealtimeWsConfig,
}));

class MockWebSocket {
  public onopen: ((event: Event) => void) | null = null;
  public onclose: ((event: CloseEvent) => void) | null = null;
  public onerror: ((event: Event) => void) | null = null;
  public onmessage: ((event: MessageEvent<string>) => void) | null = null;

  constructor(public readonly url: string) {}

  close() {}
}

function buildRoom(overrides?: Partial<BffOrchestrationData>): BffOrchestrationData {
  return {
    claims: {
      userId: "host-1",
      plan: "premium",
      sessionState: "valid",
    },
    sessionState: {
      session: {
        jamId: "jam-1",
        status: "active",
        hostUserId: "host-1",
        participants: [
          { userId: "host-1", role: "host" },
          { userId: "member-1", role: "member" },
        ],
        sessionVersion: 3,
      },
      queue: {
        jamId: "jam-1",
        queueVersion: 2,
        items: [
          { itemId: "i-1", trackId: "track-1", addedBy: "host-1" },
          { itemId: "i-2", trackId: "track-2", addedBy: "member-1" },
        ],
      },
      aggregateVersion: 5,
    },
    partial: false,
    dependencyStatuses: {
      auth: "ok",
      jam: "ok",
      playback: "ok",
      catalog: "ok",
    },
    ...overrides,
  };
}

function buildSnapshot(queueVersion: number): QueueSnapshot {
  return {
    jamId: "jam-1",
    queueVersion,
    items: [
      { itemId: "i-2", trackId: "track-2", addedBy: "member-1" },
      { itemId: "i-1", trackId: "track-1", addedBy: "host-1" },
    ],
  };
}

function buildState(queueVersion: number): SessionStateSnapshot {
  return {
    session: {
      jamId: "jam-1",
      status: "active",
      hostUserId: "host-1",
      participants: [
        { userId: "host-1", role: "host" },
        { userId: "member-1", role: "member" },
      ],
      sessionVersion: 3,
    },
    queue: buildSnapshot(queueVersion),
    aggregateVersion: 6,
  };
}

describe("JamRoomClient", () => {
  beforeEach(() => {
    vi.restoreAllMocks();
    vi.stubGlobal("WebSocket", MockWebSocket as unknown as typeof WebSocket);

    mocks.addQueueItem.mockReset();
    mocks.kickParticipant.mockReset();
    mocks.muteParticipant.mockReset();
    mocks.removeQueueItem.mockReset();
    mocks.reorderQueue.mockReset();
    mocks.endJamSession.mockReset();
    mocks.executePlayback.mockReset();
    mocks.fetchJamState.mockReset();
    mocks.fetchOrchestration.mockReset();
    mocks.fetchRealtimeWsConfig.mockReset();

    mocks.fetchOrchestration.mockResolvedValue({ success: true, data: buildRoom() });
    mocks.fetchJamState.mockResolvedValue({ success: true, data: buildState(3) });
    mocks.fetchRealtimeWsConfig.mockResolvedValue({
      success: true,
      data: {
        wsUrl: "ws://localhost:8085/ws",
        sessionId: "jam-1",
        lastSeenVersion: "5",
      },
    });
    mocks.endJamSession.mockResolvedValue({ success: true });
  });

  it("renders jam room first load from initial orchestration data", () => {
    render(
      <JamRoomClient
        jamId="jam-1"
        initialView="queue"
        initialData={buildRoom()}
      />,
    );

    expect(screen.getByRole("heading", { name: "Jam Room" })).toBeInTheDocument();
    expect(screen.getByText("track-1")).toBeInTheDocument();
    expect(screen.getByText("Participants 2")).toBeInTheDocument();
  });

  it("retries queue mutation after version conflict by refreshing snapshot", async () => {
    mocks.reorderQueue
      .mockResolvedValueOnce({
        success: false,
        error: { code: "version_conflict", message: "stale" },
      })
      .mockResolvedValueOnce({
        success: true,
        data: buildSnapshot(3),
      });

    render(
      <JamRoomClient
        jamId="jam-1"
        initialView="queue"
        initialData={buildRoom()}
      />,
    );

    await userEvent.click(screen.getAllByRole("button", { name: "Reverse" })[0]);

    await waitFor(() => {
      expect(mocks.fetchJamState).toHaveBeenCalledTimes(1);
      expect(mocks.reorderQueue).toHaveBeenCalledTimes(2);
    });
  });

  it("blocks playback interactions for non-host users", async () => {
    mocks.fetchOrchestration.mockResolvedValue(
      {
        success: true,
        data: buildRoom({
          claims: {
            userId: "member-1",
            plan: "premium",
            sessionState: "valid",
          },
        }),
      },
    );

    render(
      <JamRoomClient
        jamId="jam-host-only"
        initialView="playback"
        initialData={buildRoom({
          claims: {
            userId: "member-1",
            plan: "premium",
            sessionState: "valid",
          },
        })}
      />,
    );

    expect(screen.getAllByText("Host only control").length).toBeGreaterThan(0);

    const playButtons = screen.getAllByRole("button", { name: "Play" });
    await userEvent.click(playButtons[0]);
    expect(mocks.executePlayback).not.toHaveBeenCalled();
  });

  it("allows host moderation actions and renders muted state", async () => {
    const moderatedSession = {
      jamId: "jam-1",
      status: "active",
      hostUserId: "host-1",
      participants: [
        { userId: "host-1", role: "host" },
        { userId: "member-1", role: "member", muted: true },
      ],
      sessionVersion: 4,
    };
    mocks.muteParticipant.mockResolvedValue({ success: true, data: moderatedSession });

    render(
      <JamRoomClient
        jamId="jam-1"
        initialView="participants"
        initialData={buildRoom()}
      />,
    );

    await userEvent.type(screen.getByPlaceholderText("Moderation reason (optional)"), "spam links");
    await userEvent.click(screen.getByRole("button", { name: "Mute" }));

    await waitFor(() => {
      expect(mocks.muteParticipant).toHaveBeenCalledWith(
        "jam-1",
        expect.objectContaining({ targetUserId: "member-1", reason: "spam links" }),
      );
    });

    expect(screen.getByText("muted")).toBeInTheDocument();
  });
});
