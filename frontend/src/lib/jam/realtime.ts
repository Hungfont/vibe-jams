import type { SessionStateSnapshot } from "@/lib/jam/types";

export interface RoomRealtimeEvent {
  eventType: string;
  sessionId: string;
  aggregateVersion: number;
  occurredAt: string;
  payload: unknown;
  recovery?: boolean;
}

export type RealtimeVersionDecision = "ignore" | "apply" | "recover";

export function reduceRealtimeVersion(currentVersion: number, event: RoomRealtimeEvent): RealtimeVersionDecision {
  if (event.aggregateVersion <= currentVersion) {
    return "ignore";
  }
  if (event.aggregateVersion === currentVersion + 1) {
    return "apply";
  }
  return "recover";
}

export function shouldApplyEvent(currentVersion: number, event: RoomRealtimeEvent): boolean {
  return reduceRealtimeVersion(currentVersion, event) === "apply";
}

export function shouldRecover(currentVersion: number, event: RoomRealtimeEvent): boolean {
  return reduceRealtimeVersion(currentVersion, event) === "recover";
}

export function isSnapshotRecovery(event: RoomRealtimeEvent): event is RoomRealtimeEvent & { payload: SessionStateSnapshot } {
  return event.eventType === "jam.session.snapshot" && event.recovery === true;
}
