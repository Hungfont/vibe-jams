import type { BackendService } from "@/lib/api/types";

const DEFAULTS: Record<BackendService, string> = {
  gateway: "http://localhost:8085",
  auth: "http://localhost:8081",
  catalog: "http://localhost:8083",
  jam: "http://localhost:8080",
  playback: "http://localhost:8082",
  bff: "http://localhost:8084",
  realtime: "http://localhost:8086",
};

const ENV_KEYS: Record<BackendService, string> = {
  gateway: "API_GATEWAY_URL",
  auth: "AUTH_SERVICE_URL",
  catalog: "CATALOG_SERVICE_URL",
  jam: "JAM_SERVICE_URL",
  playback: "PLAYBACK_SERVICE_URL",
  bff: "API_SERVICE_URL",
  realtime: "RT_GATEWAY_URL",
};

export function getBackendBaseUrl(service: BackendService): string {
  const envKey = ENV_KEYS[service];
  const value = process.env[envKey];
  const raw = value && value.trim().length > 0 ? value : DEFAULTS[service];
  return raw.replace(/\/$/, "");
}

export function getDefaultWsUrl(): string {
  const explicit = process.env.RT_GATEWAY_WS_URL;
  if (explicit && explicit.trim().length > 0) {
    return explicit.trim();
  }

  const base = getBackendBaseUrl("realtime");
  if (base.startsWith("https://")) {
    return `wss://${base.slice("https://".length)}`;
  }
  if (base.startsWith("http://")) {
    return `ws://${base.slice("http://".length)}`;
  }
  return base;
}
