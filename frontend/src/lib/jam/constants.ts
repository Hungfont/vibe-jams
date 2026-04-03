export const ROOM_TABS = ["queue", "playback", "participants", "diagnostics"] as const;
export type RoomTab = (typeof ROOM_TABS)[number];

export const BLOCKING_CODES = new Set([
  "unauthorized",
  "premium_required",
  "session_ended",
]);
