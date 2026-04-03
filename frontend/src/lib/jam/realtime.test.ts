import { describe, expect, it } from "vitest";

import {
  isSnapshotRecovery,
  reduceRealtimeVersion,
  shouldApplyEvent,
  shouldRecover,
  type RoomRealtimeEvent,
} from "@/lib/jam/realtime";

function event(version: number, overrides?: Partial<RoomRealtimeEvent>): RoomRealtimeEvent {
  return {
    eventType: "jam.queue.updated",
    sessionId: "jam-1",
    aggregateVersion: version,
    occurredAt: "2026-04-03T00:00:00Z",
    payload: {},
    ...overrides,
  };
}

describe("realtime reducer", () => {
  it("ignores stale and duplicate versions", () => {
    expect(reduceRealtimeVersion(5, event(4))).toBe("ignore");
    expect(reduceRealtimeVersion(5, event(5))).toBe("ignore");
    expect(shouldApplyEvent(5, event(5))).toBe(false);
    expect(shouldRecover(5, event(5))).toBe(false);
  });

  it("applies exactly-next aggregate version", () => {
    expect(reduceRealtimeVersion(5, event(6))).toBe("apply");
    expect(shouldApplyEvent(5, event(6))).toBe(true);
    expect(shouldRecover(5, event(6))).toBe(false);
  });

  it("marks gaps for snapshot recovery", () => {
    expect(reduceRealtimeVersion(5, event(8))).toBe("recover");
    expect(shouldApplyEvent(5, event(8))).toBe(false);
    expect(shouldRecover(5, event(8))).toBe(true);
  });

  it("detects recovery snapshot events", () => {
    expect(
      isSnapshotRecovery(
        event(7, {
          eventType: "jam.session.snapshot",
          recovery: true,
          payload: {
            session: {
              jamId: "jam-1",
              status: "active",
              hostUserId: "host-1",
              participants: [{ userId: "host-1", role: "host" }],
              sessionVersion: 2,
            },
            queue: {
              jamId: "jam-1",
              queueVersion: 4,
              items: [],
            },
            aggregateVersion: 7,
          },
        }),
      ),
    ).toBe(true);

    expect(
      isSnapshotRecovery(
        event(7, {
          eventType: "jam.session.snapshot",
          recovery: false,
        }),
      ),
    ).toBe(false);
  });
});
