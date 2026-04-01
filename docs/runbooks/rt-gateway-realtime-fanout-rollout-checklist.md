# RT-Gateway Realtime Fanout Rollout Checklist (ES-P1-003)

## Feature Flag

- Flag: `FEATURE_REALTIME_FANOUT_ENABLED`
- Default target state for rollout: `true`
- Emergency rollback state: `false`

## Pre-Rollout

1. Verify `jam-service` exposes `GET /api/v1/jams/{jamId}/state` and returns `aggregateVersion`.
2. Confirm Kafka ACL includes consumer group `rt-gateway-fanout` with READ access to:
   - `jam.queue.events`
   - `jam.playback.events`
3. Run baseline tests:
   - `go test ./backend/rt-gateway/...`
   - `go test ./backend/jam-service/...`

## Canary Rollout

1. Deploy with fanout feature flag enabled for canary environment only.
2. Monitor:
   - `p95 fanout latency`
   - `gapDetectedCount`
   - `recoveryFailureCount`
   - `slowConsumerCount`
3. Validate reconnect behavior by forcing client reconnect with stale cursor and confirming snapshot fallback.

## Full Rollout

1. Increase traffic gradually by environment/cluster.
2. Keep snapshot endpoint and consumer lag dashboards visible throughout rollout window.
3. Record p95 latency and recovery metrics in release notes.

## Rollback

1. Set `FEATURE_REALTIME_FANOUT_ENABLED=false`.
2. Confirm websocket endpoint returns disabled state.
3. Keep Kafka topics and ACL unchanged for quick re-enable.
4. Capture incident metrics and root cause before reattempting rollout.
