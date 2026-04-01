## Context

`rt-gateway` currently handles websocket connectivity but does not provide a defined fanout pipeline from Kafka queue/playback topics to `jam:{sessionId}` rooms. Queue and playback services already publish session-scoped events keyed by `sessionId`, and clients need consistent realtime updates with reconnect safety when transient disconnects or consumer lag create version gaps.

This change spans multiple services (`rt-gateway`, `jam-service`) and cross-cutting concerns (stream ordering, recovery correctness, and latency under load), so design-level decisions are required before implementation.

## Goals / Non-Goals

**Goals:**
- Deliver session-scoped websocket updates to subscribers in room `jam:{sessionId}`.
- Consume queue/playback Kafka events in consumer group `rt-gateway-fanout` and fan out updates to connected clients.
- Preserve deterministic ordering for updates within the same `sessionId` stream.
- Detect version gaps and recover by fetching authoritative session snapshot from `jam-service`.
- Meet fanout performance target validated by load testing (p95 latency target).

**Non-Goals:**
- Exactly-once websocket delivery semantics.
- Cross-session global ordering.
- Replacing existing Kafka topic contracts or producer ownership in queue/playback services.
- Building client-side conflict resolution beyond reconnect + snapshot replacement behavior.

## Decisions

1. Session-scoped room model in `rt-gateway`
- Maintain room subscriptions keyed by `jam:{sessionId}`.
- On subscribe, attach connection to room and initialize per-connection cursor (`lastSeenVersion`) when provided.
- Rationale: aligns fanout partitioning with business aggregate boundary and existing Kafka keying (`sessionId`).

2. Fanout consumer group and event normalization
- Introduce consumer group `rt-gateway-fanout` that consumes jam queue/playback event streams.
- Normalize inbound events into a shared outbound websocket envelope including `sessionId`, `eventType`, `aggregateVersion`, `occurredAt`, and payload.
- Rationale: keeps transport to clients stable while preserving domain metadata required for ordering and replay logic.

3. Ordering and version window logic
- Keep `lastBroadcastVersion` per `sessionId` in gateway memory.
- If inbound `aggregateVersion == lastBroadcastVersion + 1`, broadcast and advance.
- If `aggregateVersion <= lastBroadcastVersion`, treat as duplicate/late and ignore.
- If `aggregateVersion > lastBroadcastVersion + 1`, trigger recovery path before allowing further broadcast for that session.
- Rationale: monotonic aggregate version is a low-cost deterministic guard against gaps and duplicates.

4. Snapshot recovery path for version gaps and reconnects
- Add/consume a session state read API in `jam-service` that returns authoritative queue + playback snapshot with current aggregate/version marker.
- During recovery, temporarily gate fanout for the affected `sessionId`, fetch snapshot, broadcast snapshot event, then resume stream processing from recovered version.
- For reconnecting clients with stale cursor, server MAY send snapshot immediately when gap is detected against room cursor.
- Rationale: authoritative snapshot prevents divergence and bounds complexity for clients.

5. Backpressure and latency controls
- Use bounded per-room fanout buffers and disconnect slow consumers that exceed threshold.
- Emit metrics for consumer lag, gap-recovery count, snapshot fetch latency, room size, and end-to-end fanout latency.
- Rationale: protects gateway stability and enables SLO validation.

## Risks / Trade-offs

- [Risk] Cross-topic event interleaving may violate strict sequence assumptions if versions are not globally monotonic per session. -> Mitigation: require and validate a single session-level monotonic `aggregateVersion`; fallback to snapshot on any anomaly.
- [Risk] Snapshot storms during broker lag or mass reconnect events. -> Mitigation: per-session recovery lock, short TTL snapshot cache, and rate limit concurrent snapshot fetches.
- [Risk] Memory growth for high-cardinality active sessions. -> Mitigation: LRU/TTL cleanup for room state and idle-session cursor eviction.
- [Risk] Slow websocket clients increase tail latency. -> Mitigation: bounded queues and deterministic slow-consumer eviction policy.

## Migration Plan

1. Deploy `jam-service` snapshot read endpoint and verify contract in staging.
2. Deploy `rt-gateway` fanout consumer in dark mode (consume + metrics, no client broadcast) to validate ordering and lag behavior.
3. Enable websocket fanout for a canary percentage of sessions.
4. Run load tests and verify p95 latency target, recovery correctness, and consumer lag thresholds.
5. Roll out globally; keep feature flag rollback path to disable fanout broadcasting while retaining websocket connectivity.

Rollback strategy:
- Disable realtime fanout feature flag in `rt-gateway`.
- Keep existing session APIs and command pipelines unchanged.
- Continue recording metrics to diagnose failure before re-enable.

## Open Questions

- Exact p95 target threshold value for fanout latency per environment (staging vs production).
- Whether snapshot endpoint should be internal HTTP, gRPC, or cached via Redis read-through.
- Client contract for snapshot event type/version marker naming.
