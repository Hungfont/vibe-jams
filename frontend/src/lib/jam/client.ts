import type { ApiEnvelope } from "@/lib/api/types";
import type {
  BffOrchestrationData,
  PlaybackAccepted,
  PlaybackCommandRequest,
  QueueSnapshot,
  SessionStateSnapshot,
} from "@/lib/jam/types";

async function parseEnvelope<T>(response: Response): Promise<ApiEnvelope<T>> {
  const json = (await response.json()) as ApiEnvelope<T>;
  return json;
}

export interface RealtimeWsConfig {
  wsUrl: string;
  sessionId: string;
  lastSeenVersion: string;
}

export async function fetchOrchestration(jamId: string): Promise<ApiEnvelope<BffOrchestrationData>> {
  const response = await fetch(`/api/bff/jam/${encodeURIComponent(jamId)}/orchestration`, {
    method: "POST",
    cache: "no-store",
    headers: { "Content-Type": "application/json" },
  });
  return parseEnvelope<BffOrchestrationData>(response);
}

export async function fetchJamState(jamId: string): Promise<ApiEnvelope<SessionStateSnapshot>> {
  const response = await fetch(`/api/jam/${encodeURIComponent(jamId)}/state`, {
    cache: "no-store",
  });
  return parseEnvelope<SessionStateSnapshot>(response);
}

export async function fetchRealtimeWsConfig(jamId: string, lastSeenVersion: number): Promise<ApiEnvelope<RealtimeWsConfig>> {
  const params = new URLSearchParams({
    sessionId: jamId,
    lastSeenVersion: String(lastSeenVersion),
  });
  const response = await fetch(`/api/realtime/ws-config?${params.toString()}`, {
    cache: "no-store",
  });
  return parseEnvelope<RealtimeWsConfig>(response);
}

export async function fetchQueueSnapshot(jamId: string): Promise<ApiEnvelope<QueueSnapshot>> {
  const response = await fetch(`/api/jam/${encodeURIComponent(jamId)}/queue/snapshot`, {
    cache: "no-store",
  });
  return parseEnvelope<QueueSnapshot>(response);
}

export async function executePlayback(jamId: string, payload: PlaybackCommandRequest): Promise<ApiEnvelope<PlaybackAccepted>> {
  const response = await fetch(`/api/jam/${encodeURIComponent(jamId)}/playback/commands`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload),
  });
  return parseEnvelope<PlaybackAccepted>(response);
}

export async function endJamSession(jamId: string): Promise<ApiEnvelope<unknown>> {
  const response = await fetch(`/api/jam/${encodeURIComponent(jamId)}/end`, { method: "POST" });
  return parseEnvelope<unknown>(response);
}
