# rt-gateway

Realtime websocket fanout gateway for Jam session updates.

## Endpoints

- `GET /healthz`
- `GET /metrics/fanout`
- `GET /ws?sessionId=<id>&lastSeenVersion=<optional-int>`

> In current frontend architecture, browser websocket ingress is mediated by `api-gateway -> api-service (BFF)` proxy path (`/v1/bff/mvp/realtime/ws`). Direct browser targeting of rt-gateway `/ws` is non-compliant.

## Fanout Behavior

- Subscriptions are room-scoped by `jam:{sessionId}`.
- Queue/playback events are consumed by consumer group `rt-gateway-fanout`.
- Per-session ordering is enforced by `aggregateVersion`.
- Duplicate/stale events are suppressed.
- Version gaps trigger snapshot recovery from `jam-service` endpoint:
  - `GET /api/v1/jams/{jamId}/state`
- Reconnect with stale cursor (`lastSeenVersion < current`) sends snapshot fallback before incremental updates.

## Feature Flag Rollout

- `FEATURE_REALTIME_FANOUT_ENABLED=true` enables websocket fanout and consumer loop.
- Set to `false` to disable fanout quickly while keeping process alive.

## Local Verification

```bash
go test ./...
```

Load-scenario validation:

```bash
go test ./internal/fanout -run TestLoadScenario_FanoutP95LatencyUnderTarget -count=1
```
